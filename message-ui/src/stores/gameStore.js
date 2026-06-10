import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { jwtDecode } from 'jwt-decode';
import { CARD_INFO_NUMBERS } from '../constants/cards';

const gameLog = ref([]);
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
          const getPlayerName = (id) => {
            if (!id) return '...';
            return updatedState.players?.[id]?.username || `Гравець (${id.substring(0, 4)})`;
          };


          const getCardKey = (cardType) => {
            const cardData = CARD_INFO_NUMBERS[Number(cardType)];
            return cardData ? cardData.nameKey : 'cards.unknown';
          };

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
              const payload = ev.payload || {};

              console.log(`[WS] Обробляємо подію: ${ev.type}`);

              switch (ev.type) {
                case 'ROUND_COMPARED':
                  revealedCardData.value = {
                    eventType: 'ROUND_COMPARED',
                    playerCard: Number(payload.player_card),
                    targetCard: Number(payload.target_card),
                    targetId: payload.target_id,
                    playerId: payload.player_id,
                  };
                  addToLog('log.round_compared', {
                    player: getPlayerName(payload.player_id),
                    target: getPlayerName(payload.target_id)
                  });
                  break;

                case 'CARD_REVEALED':
                case 'PRIEST_EFFECT':
                  if (currentUuid === payload.viewer_id || currentUuid === payload.target_id) {
                    if (payload.card) {
                      revealedCardData.value = {
                        eventType: 'PRIEST_EFFECT',
                        cardType: Number(payload.card),
                        targetId: payload.target_id,
                        viewerId: payload.viewer_id
                      };
                    }
                  }
                  addToLog('log.priest_effect', {
                    player: getPlayerName(payload.viewer_id),
                    target: getPlayerName(payload.target_id)
                  });
                  break;

                case 'CARD_PLAYED':
                  addToLog('log.card_played', {
                    player: getPlayerName(payload.player_id),
                    card: getCardKey(payload.card),
                    hasTarget: !!payload.target_id,
                    target: getPlayerName(payload.target_id)
                  });
                  break;

                case 'GUARD_HIT':
                  addToLog('log.guard_hit', {
                    player: getPlayerName(payload.player_id),
                    target: getPlayerName(payload.target_id),
                    guess: getCardKey(payload.guess)
                  });
                  break;

                case 'GUARD_MISS':
                  addToLog('log.guard_miss', {
                    player: getPlayerName(payload.player_id),
                    target: getPlayerName(payload.target_id),
                    guess: getCardKey(payload.guess)
                  });
                  break;

                case 'BARON_RESULT':
                  addToLog('log.baron_result', {
                    winner: getPlayerName(payload.winner_id),
                    loser: getPlayerName(payload.loser_id)
                  });
                  break;

                case 'PLAYER_ELIMINATED':
                  addToLog('log.player_eliminated', {
                    player: getPlayerName(payload.player_id),
                    reason: `reasons.${payload.reason}`
                  });
                  break;

                case 'HANDS_SWAPPED':
                  addToLog('log.hands_swapped', {
                    player: getPlayerName(payload.player_id),
                    target: getPlayerName(payload.target_id)
                  });
                  break;

                case 'SPY_BONUS':
                  addToLog('log.spy_bonus', {
                    player: getPlayerName(payload.player_id),
                    points: payload.points
                  });
                  break;

                case 'ROUND_END':
                  addToLog('log.round_end', {
                    winner: getPlayerName(payload.winner_id),
                    reason: `reasons.${payload.reason}`
                  });
                  break;

                case 'PLAYER_LEFT':
                  addToLog('log.player_left', {
                    player: getPlayerName(payload.player_id)
                  });
                  break;

                case 'CARD_DRAWN':
                  addToLog('log.card_drawn', {
                    player: getPlayerName(payload.player_id)
                  });
                  break;

                case 'CHANCELLOR_DRAWN':
                case 'CHANCELLOR_RESOLVED':
                  console.log(`[WS] Подія канцлера ${ev.type} зафіксована, але лог не реалізовано.`);
                  break;
              }
            });
          }

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

  function addToLog(messageKey, namedArgs = {}) {
    gameLog.value.push({
      id: crypto?.randomUUID ? crypto.randomUUID() : Math.random().toString(),
      timestamp: new Date().toLocaleTimeString(),
      messageKey,
      namedArgs
    });
  }

  return {
    gameState,
    gameLog,
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