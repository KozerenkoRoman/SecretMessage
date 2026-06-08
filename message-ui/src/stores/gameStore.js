import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { jwtDecode } from 'jwt-decode';

const EMPTY_GAME_STATE = Object.freeze({
  is_started: false,
  players: {},
  current_player_id: '',
  my_hand: [],
  discard_pile: [],
  seconds_left: 0,
  chancellor_options: [],
});

function makeEmptyGameState() {
  return {
    is_started: false,
    players: {},
    current_player_id: '',
    my_hand: [],
    discard_pile: [],
    seconds_left: 0,
    chancellor_options: [],
  };
}

function extractUserIdFromToken(token) {
  if (!token || typeof token !== 'string') return '';
  try {
    const decoded = jwtDecode(token);
    const id = decoded?.user_id;
    return typeof id === 'string' ? id : '';
  } catch (e) {
    console.warn('[gameStore] Не вдалося декодувати JWT:', e);
    return '';
  }
}

export const useGameStore = defineStore('gameStore', () => {
  const socket = ref(null);
  const isConnected = ref(false);
  const error = ref(null);
  const currentRoomID = ref('');
  const lobbyRooms = ref([]);
  let reconnectTimeout = null;
  const isIntentionallyClosed = ref(false);
  const gameState = ref(makeEmptyGameState());
  const revealedCardData = ref(null);
  const authToken = ref(localStorage.getItem('token') || '');

  // Обчислювальний ID нашого користувача з JWT-токену
  const myID = computed(() => extractUserIdFromToken(authToken.value));

  const isGameStarted = computed(() => {
    return (
      gameState.value?.is_started ||
      (Array.isArray(gameState.value?.turn_order) && gameState.value.turn_order.length > 0) ||
      false
    );
  });

  // Обчислювальна властивість для отримання карт у нашій руці
  const myCards = computed(() => {
    const id = myID.value;
    const playersData = gameState.value?.players;
    if (id && playersData && typeof playersData === 'object' && !Array.isArray(playersData)) {
      const me = playersData[id];
      if (me && Array.isArray(me.hand)) return me.hand;
    }
    return Array.isArray(gameState.value?.my_hand) ? gameState.value.my_hand : [];
  });

  const activePlayers = computed(() => {
    const playersData = gameState.value?.players;
    if (!playersData) return [];
    if (Array.isArray(playersData)) {
      return playersData.filter((p) => p && typeof p === 'object' && typeof p.id === 'string');
    }
    return Object.keys(playersData)
      .map((id) => {
        const raw = playersData[id];
        if (!raw || typeof raw !== 'object') return null;
        return { ...raw, id };
      })
      .filter((p) => p !== null);
  });

  function refreshAuthToken() {
    authToken.value = localStorage.getItem('token') || '';
  }

  function clearError() {
    error.value = null;
  }

  function setErrorFromPacket(packet) {
    if (!packet || typeof packet !== 'object') {
      error.value = { code: 'ERR_INTERNAL', message: 'Unknown error', details: null };
      return;
    }
    const code = (typeof packet.error_code === 'string' && packet.error_code) || (typeof packet.code === 'string' && packet.code) || 'ERR_INTERNAL';
    error.value = {
      code,
      message: typeof packet.message === 'string' ? packet.message : '',
      details: packet.details ?? null,
      requestId: typeof packet.request_id === 'string' ? packet.request_id : undefined,
    };
    console.error('[WS Серверна помилка]:', error.value);
  }

  function connectToHub(roomID = null) {
    const targetRoomID = roomID || '';
    if (socket.value && socket.value.readyState === WebSocket.OPEN) {
      console.log(`[WS] Сокет уже відкритий. Міняємо кімнату з "${currentRoomID.value}" на "${targetRoomID}"`);
      currentRoomID.value = targetRoomID;
      if (targetRoomID) {
        sendWSMessage('JOIN', targetRoomID, null, null);
      }
      return;
    }

    if (socket.value && socket.value.readyState === WebSocket.CONNECTING) {
      console.log(`[WS] Сокет зараз підключається. Оновлюємо цільову кімнату на: ${targetRoomID}`);
      currentRoomID.value = targetRoomID;
      return;
    }

    currentRoomID.value = targetRoomID;
    isIntentionallyClosed.value = false;
    refreshAuthToken();

    const token = authToken.value;
    if (!token) {
      console.warn('[WS] Спроба підключення без JWT токена');
      return;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const backendHost = window.location.hostname === 'localhost' ? 'localhost:3000' : window.location.host;
    const wsUrl = `${protocol}//${window.location.host}/ws?token=${token}`;

    console.log(`[WS] Створення нового підключення до шлюзу: ${wsUrl}`);
    try {
      socket.value = new WebSocket(wsUrl);
    } catch (e) {
      console.error('[WS] Критична помилка створення WebSocket:', e);
      return;
    }

    socket.value.onopen = () => {
      isConnected.value = true;
      error.value = null;
      console.log('[WS] Сокет успішно відкрито з бекендом.');
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
        reconnectTimeout = null;
      }
      if (currentRoomID.value) {
        sendWSMessage('JOIN', currentRoomID.value, null, null);
      }
    };

    socket.value.onmessage = (event) => {
      try {
        const packet = JSON.parse(event.data);
        console.log('[WS] Отримано пакунок події від Go:', packet);

        if (packet.status === 'error' || packet.type === 'ERROR') {
          setErrorFromPacket(packet);
          return;
        }

        if (packet.type === 'LOBBY_LIST_UPDATED' || packet.type === 'LOBBY_UPDATED') {
          lobbyRooms.value = Array.isArray(packet.data) ? packet.data : (packet.payload || []);
          return;
        }

        if (packet.type === 'ROOM_UPDATED' || packet.type === 'STATE_UPDATE' || packet.state) {
          let rawState = packet.state || packet.payload;
          if (!rawState || typeof rawState !== 'object') return;

          // 1. Повністю ізольовано клонуємо стан від сервера
          let updatedState = JSON.parse(JSON.stringify(rawState));
          const currentUuid = myID.value;

          // 2. Гарантуємо повне очищення та синхронізацію дублюючих полів
          if (updatedState.players && updatedState.players[currentUuid]) {
            // Синхронізуємо my_hand суворо з тим, що знає сервер про нашу руку
            updatedState.my_hand = [...updatedState.players[currentUuid].hand];
          }

          // 3. Обробляємо події (тільки для збереження даних у діалоги/нотифікації)
          if (Array.isArray(packet.events)) {
            packet.events.forEach((ev) => {
              if (!ev || typeof ev !== 'object') return;

              // Обробка ефекту Барона
              if (ev.type === 'ROUND_COMPARED') {
                const payload = ev.payload || {};
                revealedCardData.value = {
                  eventType: 'ROUND_COMPARED',
                  playerCard: Number(payload.player_card),
                  targetCard: Number(payload.target_card),
                  targetId: payload.target_id,
                  playerId: payload.player_id,
                };
              }
              // Обробка ефекту Священника
              else if (ev.type === 'PRIEST_EFFECT' || ev.type === 'CARD_REVEALED') {
                const payload = ev.payload || {};
                if (payload.viewer_id === currentUuid) {
                  const revealedCard = payload.card !== undefined ? payload.card : payload.card_type;
                  if (revealedCard !== undefined && revealedCard !== null) {
                    revealedCardData.value = {
                      eventType: 'CARD_REVEALED',
                      cardType: Number(revealedCard),
                      targetId: payload.target_id
                    };
                  }
                }
              }
              // Обробка ефекту Канцлера
              else if (ev.type === 'CHANCELLOR_DRAWN') {
                const payload = ev.payload || {};
                if (payload.player_id === currentUuid) {
                  console.log('[WS DEBUG] Подія Канцлера: відкриваємо інтерфейс вибору карт.');
                  // Тут за потреби можна виставити локальний флаг увімкнення вікна Канцлера:
                  // isChancellorModalOpen.value = true;
                }
              }
            });
          }

          // 4. Записуємо чистий стан у стор. Старі карти затираються повністю!
          gameState.value = updatedState;
        }
      } catch (err) {
        console.error('[WS] Помилка десеріалізації:', err);
      }
    };

    socket.value.onclose = (event) => {
      isConnected.value = false;
      socket.value = null;
      console.log('[WS] Сокет закрився.', event);
      if (!isIntentionallyClosed.value) {
        reconnectTimeout = setTimeout(() => {
          connectToHub(currentRoomID.value);
        }, 3000);
      }
    };

    socket.value.onerror = (err) => {
      console.error('[WS] Помилка з\'єднання:', err);
    };
  }

  function sendWSMessage(type, overrideRoomID = null, actionPayload = null, chancellorPayload = null) {
    if (!socket.value || socket.value.readyState !== WebSocket.OPEN) {
      console.warn('[WS] Спроба відправки повідомлення у закритий сокет.');
      return;
    }
    const finalRoomID = overrideRoomID || currentRoomID.value;
    const message = {
      room_id: finalRoomID || '',
      type,
      request_id: typeof crypto?.randomUUID === 'function' ? crypto.randomUUID() : Math.random().toString(36).substring(2),
      action: actionPayload,
      chancellor_action: chancellorPayload,
    };
    console.log('[WS] Відправка на бекенд:', message);
    socket.value.send(JSON.stringify(message));
  }

  function clearRevealedData() {
    revealedCardData.value = null;
  }

  function leaveCurrentRoom() {
    if (currentRoomID.value) {
      sendWSMessage('LEAVE', currentRoomID.value, null, null);
      currentRoomID.value = '';
      revealedCardData.value = null;
      gameState.value = makeEmptyGameState();
    }
  }

  function disconnect() {
    if (currentRoomID.value && socket.value && socket.value.readyState === WebSocket.OPEN) {
      sendWSMessage('LEAVE', currentRoomID.value, null, null);
    }
    isIntentionallyClosed.value = true;
    revealedCardData.value = null;
    if (reconnectTimeout) {
      clearTimeout(reconnectTimeout);
      reconnectTimeout = null;
    }
    if (socket.value) {
      socket.value.close();
      socket.value = null;
    }
    isConnected.value = false;
    currentRoomID.value = '';
  }

  return {
    gameState,
    lobbyRooms,
    revealedCardData,
    isConnected,
    error,
    currentRoomID,
    myID,
    isGameStarted,
    activePlayers,
    myCards,
    refreshAuthToken,
    clearError,
    clearRevealedData,
    leaveCurrentRoom,
    connectToHub,
    sendWSMessage,
    disconnect,
  };
});