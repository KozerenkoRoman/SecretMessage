import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { jwtDecode } from 'jwt-decode';
import { CARD_INFO_NUMBERS } from '../constants/cards';
import { mergeState, mergeArray } from '../utils/merge';

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
  const isIntentionallyClosed = ref(false);
  const gameState = ref(makeEmptyGameState());
  const revealedCardData = ref(null);
  const authToken = ref(localStorage.getItem('token') || '');

  const latestTurnAlert = ref(null);
  let alertTimeout = null;

  /*
    Сигнальні ref'и для UI.
    Раніше підписники (RoomManager) визначали "ми отримали стан від сервера"
    за фактом перепризначення gameState.value (через spread). Після того як
    ми перейшли на in-place mergeState (для усунення мерехтіння карт),
    кореневе посилання НЕ змінюється, тож shallow watch на gameState
    більше не тригериться. Тому надаємо явні сигнали:
      - hasReceivedState: ми хоча б раз отримали валідний ROOM_UPDATED
      - stateVersion:     монотонний лічильник для watcher'ів, які хочуть
                          реагувати на КОЖНЕ оновлення стану, а не лише
                          на зміну посилання.
  */
  const hasReceivedState = ref(false);
  const stateVersion = ref(0);

  /*
    Похідні (derived) поля стану дошки. Раніше ця логіка жила у
    BoardView.vue усередині важкого watch({ deep: true }), що:
      - перебудовував усе при кожному ROOM_UPDATED;
      - блукав по ВСІХ гравцях .find()'ом і spread'ив об'єкти;
      - писав у локальний стан компонента, який гасився при ремаунті.
    Тут ми обчислюємо все ОДИН раз при отриманні WS-пакета й тримаємо
    стабільні посилання, тож компонент лише читає плоскі ref'и.
  */
  // Глобальна послідовність відбою з монотонним seq.
  const discardSequence = ref([]);
  // pid -> остання зіграна карта (число або об'єкт {type:...}).
  const lastPlayedCardsByPlayer = ref({});
  // pid -> скільки карт у discard_pile ми вже зафіксували
  // (внутрішній прапор, не експонується назовні).
  const _discardSeenLengths = Object.create(null);
  let _discardSeqCounter = 0;

  // Стабільні UID для слотів руки поточного гравця.
  // [{ uid, cardType, index }] — :key='slot.uid' у v-for НЕ змінюється,
  // поки кількість карт стабільна, тому DOM-вузли не пересоздаються.
  const handSlots = ref([]);
  let _handSlotCounter = 0;
  const _nextHandUid = () => `hand-${++_handSlotCounter}`;

  // Локальний таймер: тримаємо ВІДОКРЕМЛЕНО від gameState.seconds_left,
  // щоб тиканина не мутувала об'єкт, що мерджиться сервером (інакше
  // отримаємо feedback-loop і повторні рендери щосекунди).
  const secondsLeft = ref(0);

  // Перенесено всередину стору для коректного скидання
  const gameLog = ref([]);

  const myID = computed(() => extractUserIdFromToken(authToken.value));

  let timerInterval = null;
  let reconnectTimeout = null;

  function _resetBoardDerived() {
    discardSequence.value = [];
    lastPlayedCardsByPlayer.value = {};
    latestTurnAlert.value = null;
    for (const k of Object.keys(_discardSeenLengths)) delete _discardSeenLengths[k];
    _discardSeqCounter = 0;
    handSlots.value = [];
  }

  const isGameStarted = computed(() => {
    return (
      gameState.value?.is_started ||
      (Array.isArray(gameState.value?.turn_order) && gameState.value.turn_order.length > 0) ||
      false
    );
  });

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

  function clearLog() {
    gameLog.value = [];
  }

  /**
   * Робимо легкий snapshot ключових полів СТАРОГО стану ДО злиття,
   * щоб після mergeState мати з чим порівнювати (mergeState мутує
   * gameState.value in-place, тож після нього "старого" стану вже нема).
   * Повертає лише те, що реально потрібне для дерев'яної логіки дошки.
   */
  function _snapshotForBoardDerivation(state) {
    if (!state || typeof state !== 'object') return { discardLenByPid: {} };
    const playersData = state.players;
    const out = { discardLenByPid: Object.create(null) };
    if (!playersData) return out;
    if (Array.isArray(playersData)) {
      for (const p of playersData) {
        if (!p || typeof p !== 'object' || typeof p.id !== 'string') continue;
        out.discardLenByPid[p.id] = Array.isArray(p.discard_pile) ? p.discard_pile.length : 0;
      }
    } else {
      for (const id of Object.keys(playersData)) {
        const p = playersData[id];
        if (!p || typeof p !== 'object') continue;
        out.discardLenByPid[id] = Array.isArray(p.discard_pile) ? p.discard_pile.length : 0;
      }
    }
    return out;
  }

  /**
   * Оновлює derived-поля для дошки після того, як newState уже злито в
   * gameState.value. Все відбувається ОДИН раз на пакет ROOM_UPDATED.
   * @param newState - власне gameState.value (in-place після merge)
   * @param prevSnap - snapshot, зроблений ДО merge (через _snapshotForBoardDerivation)
   */
  function _updateBoardDerived(newState, prevSnap) {
    if (!newState || typeof newState !== 'object') return;

    const playersData = newState.players;
    if (!playersData || typeof playersData !== 'object') return;

    // Нормалізуємо до плоского масиву [{id, ...}], не мутуючи playersData.
    let playersList;
    if (Array.isArray(playersData)) {
      playersList = playersData.filter(
        (p) => p && typeof p === 'object' && typeof p.id === 'string' && p.id.length > 0
      );
    } else {
      playersList = [];
      for (const id of Object.keys(playersData)) {
        const raw = playersData[id];
        if (raw && typeof raw === 'object') playersList.push({ raw, id });
      }
    }

    const turnOrder = Array.isArray(newState.turn_order) ? newState.turn_order : [];
    const currentTurnIdx = typeof newState.current_turn === 'number' ? newState.current_turn : -1;
    const currentActiveID = currentTurnIdx >= 0 && currentTurnIdx < turnOrder.length
      ? turnOrder[currentTurnIdx]
      : newState.current_player_id || '';

    const localId = myID.value;

    // ---- discard sequence + lastPlayedCardsByPlayer ----
    // Йдемо у порядку turn_order (далі - усі інші id-ки), щоб події у відбої
    // лягали детерміновано в одному порядку у всіх клієнтів.
    const seenIds = new Set();
    const ordered = [];
    for (const id of turnOrder) {
      const found = playersList.find((p) =>
        Array.isArray(playersData) ? p.id === id : p.id === id
      );
      if (found) {
        ordered.push(found);
        seenIds.add(id);
      }
    }
    for (const p of playersList) {
      if (!seenIds.has(p.id)) ordered.push(p);
    }

    // lpc - реактивний proxy від ref, тому правки ключів (lpc[pid] = ...,
    // delete lpc[pid]) автоматично трекаються Vue. Жодного ручного
    // re-assign'у наприкінці не потрібно.
    const lpc = lastPlayedCardsByPlayer.value;

    for (const entry of ordered) {
      const pid = entry.id;
      const p = Array.isArray(playersData) ? entry : entry.raw;
      const discard = Array.isArray(p.discard_pile) ? p.discard_pile : null;
      const curLen = discard ? discard.length : 0;
      const seenLen = _discardSeenLengths[pid] || 0;
      const prevLen = prevSnap?.discardLenByPid?.[pid] ?? seenLen;

      if (curLen > seenLen && discard) {
        for (let i = seenLen; i < curLen; i++) {
          const rawCard = discard[i];
          const isObject = typeof rawCard === 'object' && rawCard !== null;
          discardSequence.value.push({
            seq: _discardSeqCounter++,
            type: isObject ? rawCard.type : rawCard,
            playerId: pid,
          });
        }
        _discardSeenLengths[pid] = curLen;
      } else if (curLen < seenLen) {
        // Колоду / partію перетасували - синхронізуємо лічильник, але
        // НЕ чіпаємо discardSequence: він глобальний по партії та чиститься
        // явно (round-bump / leave / нова кімната).
        _discardSeenLengths[pid] = curLen;
      }

      // lastPlayedCardsByPlayer - тільки для опонентів, на основі дельти
      // ВІДНОСНО ПОПЕРЕДНЬОГО ПАКЕТА (а не до seenLen).
      if (pid === localId) continue;
      if (curLen > prevLen && discard) {
        const lastCard = discard[curLen - 1];
        if (lpc[pid] !== lastCard) lpc[pid] = lastCard;
      } else if (pid === currentActiveID && lpc[pid] !== undefined) {
        delete lpc[pid];
      }
      if (curLen === 0 && lpc[pid] !== undefined) {
        delete lpc[pid];
      }
    }

    // ---- handSlots для локального гравця ----
    const myHand = (() => {
      if (!localId) return [];
      if (Array.isArray(playersData)) {
        const me = playersData.find((p) => p && p.id === localId);
        return me && Array.isArray(me.hand) ? me.hand : [];
      }
      const me = playersData[localId];
      if (me && Array.isArray(me.hand)) return me.hand;
      return Array.isArray(newState.my_hand) ? newState.my_hand : [];
    })();

    const slots = handSlots.value;
    if (slots.length !== myHand.length) {
      // Кількість карт реально змінилась - перевипускаємо UID-и.
      // Реюзаємо UID-и для тих позицій, що збереглися (стабільний :key).
      const fresh = [];
      for (let i = 0; i < myHand.length; i++) {
        const existing = slots[i];
        fresh.push({
          uid: existing ? existing.uid : _nextHandUid(),
          cardType: myHand[i],
          index: i,
        });
      }
      handSlots.value = fresh;
    } else {
      // Та сама кількість - точково оновлюємо лише ті слоти, де
      // cardType/index реально змінилися. UID не змінюється ніколи у
      // цій гілці, тож DOM-вузли карт не пересоздаються.
      for (let i = 0; i < myHand.length; i++) {
        const slot = slots[i];
        if (!slot) {
          slots[i] = { uid: _nextHandUid(), cardType: myHand[i], index: i };
        } else if (slot.cardType !== myHand[i] || slot.index !== i) {
          // Створюємо новий об'єкт-обгортку (щоб тригернути реактивність
          // у v-for, який ітерує по slots), але зберігаємо УЖЕ виданий uid.
          slots[i] = { uid: slot.uid, cardType: myHand[i], index: i };
        }
      }
    }
  }

  function setErrorFromPacket(packet) {
    if (!packet || typeof packet !== 'object') {
      error.value = { code: 'ERR_INTERNAL', message: 'Unknown error', details: null };
      return;
    }
    const code =
      (typeof packet.error_code === 'string' && packet.error_code) ||
      (typeof packet.code === 'string' && packet.code) ||
      'ERR_INTERNAL';
    error.value = {
      code,
      message: typeof packet.message === 'string' ? packet.message : '',
      details: packet.details ?? null,
      requestId: typeof packet.request_id === 'string' ? packet.request_id : undefined,
    };
    console.error('[WS Server Error]:', error.value);
  }

  function connectToHub(roomID = null) {
    const targetRoomID = roomID || '';
    if (socket.value && socket.value.readyState === WebSocket.OPEN) {
      console.log(`[WS] Сокет уже відкритий. Міняємо кімнату з "${currentRoomID.value}" на "${targetRoomID}"`);

      if (currentRoomID.value !== targetRoomID) {
        clearLog();
        // Нова кімната - старий state вже не релевантний; чекаємо на свіжий
        // ROOM_UPDATED перш ніж вважати з'єднання "готовим".
        hasReceivedState.value = false;
        stateVersion.value = 0;
        _resetBoardDerived();
      }

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

    clearLog();
    currentRoomID.value = targetRoomID;
    isIntentionallyClosed.value = false;
    hasReceivedState.value = false;
    stateVersion.value = 0;
    _resetBoardDerived();
    refreshAuthToken();
    const token = authToken.value;
    if (!token) {
      console.warn('[WS] Спроба підключення без JWT токена');
      return;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const backendHost = window.location.hostname === 'localhost' ? 'localhost:3000' : window.location.host;
    const wsUrl = `${protocol}//${backendHost}/ws?token=${token}`;
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

          // Знімок попереднього стану (лише потрібні поля) ДО merge,
          // бо mergeState мутує gameState.value in-place.
          const prevSnap = _snapshotForBoardDerivation(gameState.value);
          const prevRound = typeof gameState.value?.round_number === 'number'
            ? gameState.value.round_number
            : null;
          const prevDeckLen = Array.isArray(gameState.value?.deck)
            ? gameState.value.deck.length
            : null;

          // Структурне злиття: НЕ робимо JSON-клон, інакше всі масиви
          // (зокрема hand[]) отримують нові посилання на кожному
          // ROOM_UPDATED, що ламає reactivity-діффінг та може провокувати
          // мерехтіння карт через будь-які transition-залежні стилі.
          // Замість цього мутуємо існуючі поля коли їхня структура збіглася.
          const updatedState = mergeState(gameState.value, rawState);

          // Якщо почався новий раунд або колода свіжо перегенерована,
          // глобальний відбій логічно скидається.
          const newRound = typeof updatedState.round_number === 'number'
            ? updatedState.round_number
            : null;
          const newDeckLen = Array.isArray(updatedState.deck)
            ? updatedState.deck.length
            : null;
          const roundBumped = prevRound !== null && newRound !== null && newRound > prevRound;
          const deckGrew = prevDeckLen !== null && newDeckLen !== null && newDeckLen > prevDeckLen;
          if (roundBumped || deckGrew) {
            discardSequence.value = [];
            for (const k of Object.keys(_discardSeenLengths)) delete _discardSeenLengths[k];
            _discardSeqCounter = 0;
            lastPlayedCardsByPlayer.value = {};
          }

          const getPlayerName = (id) => {
            if (!id) return '...';
            return updatedState.players?.[id]?.username || `Гравець (${id.substring(0, 4)})`;
          };

          const getCardKey = (cardType) => {
            const cardData = CARD_INFO_NUMBERS[Number(cardType)];
            return cardData ? cardData.nameKey : 'cards.unknown';
          };

          const currentUuid = myID.value;
          if (updatedState.players && updatedState.players[currentUuid]) {
            const srcHand = updatedState.players[currentUuid].hand;
            if (Array.isArray(srcHand)) {
              if (!Array.isArray(updatedState.my_hand)) {
                updatedState.my_hand = [];
              }
              // Зливаємо у наявний масив, щоб не зламати посилання,
              // вже узгоджене mergeState'ом вище.
              mergeArray(updatedState.my_hand, srcHand);
            }
          }

          if (Array.isArray(packet.events)) {
            let lastDrawnPlayerId = null;

            // Шукаємо подію CARD_PLAYED для банера всередині поточного пакета
            const cardPlayedEvent = packet.events.find((e) => e && e.type === "CARD_PLAYED");
            if (cardPlayedEvent && cardPlayedEvent.payload) {
              const { card, player_id, } = cardPlayedEvent.payload;

              let guessedCardId = null;
              let princeDiscardedCardId = null;
              if (CARD_INFO_NUMBERS[Number(card)]?.type === "GUARD") {
                // Якщо зіграно стражника, шукаємо результат у цьому ж пакеті подій
                const guardEvent = packet.events.find(
                  (e) => e && (e.type === "GUARD_MISS" || e.type === "GUARD_HIT")
                );
                if (guardEvent && guardEvent.payload && guardEvent.payload.guess !== undefined) {
                  guessedCardId = Number(guardEvent.payload.guess);
                }
              }

              const hasDiscardedCard = cardPlayedEvent.payload.discarded_card !== undefined && cardPlayedEvent.payload.discarded_card !== null;
              if (CARD_INFO_NUMBERS[Number(card)]?.type === "PRINCE" && hasDiscardedCard) {
                princeDiscardedCardId = Number(cardPlayedEvent.payload.discarded_card);
              }
              if (alertTimeout) clearTimeout(alertTimeout);

              // Записуємо дані. Оскільки під назву гравця потрібен username, беремо його ліниво:
              const foundPlayerName = rawState.players?.[player_id]?.username ||
                gameState.value.players?.[player_id]?.username || '...';

              latestTurnAlert.value = {
                cardType: Number(card),
                playerName: foundPlayerName,
                guessCard: guessedCardId,
                discardedCard: princeDiscardedCardId,
              };

              alertTimeout = setTimeout(() => {
                latestTurnAlert.value = null;
              }, 3500);
            }

            packet.events.forEach((ev) => {
              if (!ev || typeof ev !== 'object') return;
              const payload = ev.payload || {};
              switch (ev.type) {
                case 'ROUND_COMPARED':
                  revealedCardData.value = {
                    eventType: 'ROUND_COMPARED',
                    playerCard: Number(payload.player_card),
                    targetCard: Number(payload.target_card),
                    targetId: payload.target_id,
                    playerId: payload.player_id,
                  };
                  addToLog('log.round_compared', { player: getPlayerName(payload.player_id), target: getPlayerName(payload.target_id) });
                  break;
                case 'CARD_REVEALED':
                case 'PRIEST_EFFECT':
                  if (currentUuid === payload.viewer_id || currentUuid === payload.target_id) {
                    if (payload.card) {
                      revealedCardData.value = { eventType: 'PRIEST_EFFECT', cardType: Number(payload.card), targetId: payload.target_id, viewerId: payload.viewer_id };
                    }
                  }
                  addToLog('log.priest_effect', { player: getPlayerName(payload.viewer_id), target: getPlayerName(payload.target_id) });
                  break;
                case 'CARD_PLAYED':
                  const isPrince = Number(payload.card) === 5;
                  if (isPrince && payload.discarded_card !== undefined && payload.discarded_card !== null) {
                    addToLog('log.prince_effect', {
                      player: getPlayerName(payload.player_id),
                      target: getPlayerName(payload.target_id),
                      discarded_card: getCardKey(payload.discarded_card)
                    });
                  } else {
                    addToLog(payload.target_id ? 'log.card_played_targeted' : 'log.card_played', {
                      player: getPlayerName(payload.player_id),
                      card: getCardKey(payload.card),
                      target: getPlayerName(payload.target_id)
                    });
                  }
                  break;
                case 'GUARD_HIT':
                  addToLog('log.guard_hit', { player: getPlayerName(payload.player_id), target: getPlayerName(payload.target_id), guess: getCardKey(payload.guess) });
                  break;
                case 'GUARD_MISS':
                  addToLog('log.guard_miss', { player: getPlayerName(payload.player_id), target: getPlayerName(payload.target_id), guess: getCardKey(payload.guess) });
                  break;
                case 'BARON_RESULT':
                  addToLog('log.baron_result', { winner: getPlayerName(payload.winner_id), loser: getPlayerName(payload.loser_id), loser_card: getCardKey(payload.loser_card) });
                  break;
                case 'PLAYER_ELIMINATED':
                  addToLog('log.player_eliminated', { player: getPlayerName(payload.player_id), reason: `reasons.${payload.reason}` });
                  break;
                case 'HANDS_SWAPPED':
                  addToLog('log.hands_swapped', { player: getPlayerName(payload.player_id), target: getPlayerName(payload.target_id) });
                  break;
                case 'SPY_BONUS':
                  addToLog('log.spy_bonus', { player: getPlayerName(payload.player_id), points: payload.points });
                  break;
                case 'ROUND_END':
                  addToLog('log.round_end', { winner: getPlayerName(payload.winner_id), reason: `reasons.${payload.reason}` });
                  break;
                case 'PLAYER_LEFT':
                  addToLog('log.player_left', { player: getPlayerName(payload.player_id) });
                  break;
                case 'CARD_DRAWN':
                  if (payload.player_id === lastDrawnPlayerId) {
                    console.log(`[Log Skipper] Пропущено дублюючу подію CARD_DRAWN для гравця: ${payload.player_id}`);
                    break;
                  }
                  lastDrawnPlayerId = payload.player_id;
                  addToLog('log.card_drawn', { player: getPlayerName(payload.player_id) });
                  break;
                case 'CHANCELLOR_DRAWN':
                  addToLog('log.chancellor_drawn', { player: getPlayerName(payload.player_id) });
                  break;
                case 'CHANCELLOR_RESOLVED':
                  addToLog('log.chancellor_resolved', { player: getPlayerName(payload.player_id) });
                  break;
              }
            });
          }

          // Перерахунок derived-полів дошки (discardSequence,
          // lastPlayedCardsByPlayer, handSlots). Робиться РАЗ на пакет.
          _updateBoardDerived(gameState.value, prevSnap);

          // mergeState вже застосував зміни in-place до gameState.value,
          // тому окремо перезаписувати об'єкт не потрібно.
          if (typeof updatedState.seconds_left === 'number') {
            // Дрейф > 1с -> рестартуємо локальний тікер під серверну
            // істину. Тікер мутує ЛИШЕ secondsLeft, не gameState, тому
            // не провокує feedback-loop у merge.
            const drift = Math.abs(secondsLeft.value - updatedState.seconds_left);
            if (drift > 1) startLocalTimer(updatedState.seconds_left);
          }

          // Сигналізуємо UI, що стан отримано/оновлено. Це КРИТИЧНО для
          // RoomManager: він знімає loading-екран саме за цим прапором,
          // а не за зміною посилання gameState (його більше немає).
          hasReceivedState.value = true;
          stateVersion.value++;
        }
      } catch (err) {
        console.error('[WS] Помилка десеріалізації:', err);
      }
    };

    socket.value.onclose = (event) => {
      isConnected.value = false;
      socket.value = null;
      if (timerInterval) clearInterval(timerInterval);
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

  function startLocalTimer(initialSeconds) {
    if (timerInterval) {
      clearInterval(timerInterval);
      timerInterval = null;
    }
    if (typeof initialSeconds !== 'number' || initialSeconds <= 0) {
      secondsLeft.value = 0;
      return;
    }

    // ВАЖЛИВО: тикаємо ВИКЛЮЧНО локальний secondsLeft. НЕ мутуємо
    // gameState.seconds_left, інакше тікер інвалідуватиме всю реактивну
    // піддерево щосекунди (re-render storm) і конфліктуватиме з merge.
    secondsLeft.value = initialSeconds;

    timerInterval = setInterval(() => {
      if (secondsLeft.value > 0) {
        secondsLeft.value--;
      } else {
        clearInterval(timerInterval);
        timerInterval = null;
      }
    }, 1000);
  }

  function stopLocalTimer() {
    if (timerInterval) {
      clearInterval(timerInterval);
      timerInterval = null;
    }
    secondsLeft.value = 0;
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
    stopLocalTimer();
    clearError();
    clearLog();
    if (currentRoomID.value) {
      sendWSMessage('LEAVE', currentRoomID.value, null, null);
      currentRoomID.value = '';
      revealedCardData.value = null;
      gameState.value = makeEmptyGameState();
      hasReceivedState.value = false;
      stateVersion.value = 0;
      _resetBoardDerived();
    }
  }

  function disconnect() {
    stopLocalTimer();
    if (currentRoomID.value && socket.value && socket.value.readyState === WebSocket.OPEN) {
      sendWSMessage('LEAVE', currentRoomID.value, null, null);
    }
    isIntentionallyClosed.value = true;
    revealedCardData.value = null;
    clearLog();
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
    hasReceivedState.value = false;
    stateVersion.value = 0;
    _resetBoardDerived();
  }

  function addToLog(messageKey, namedArgs = {}) {
    setTimeout(() => {
      gameLog.value.push({
        id: typeof crypto?.randomUUID === 'function' ? crypto.randomUUID() : Math.random().toString(36).substring(2),
        timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
        messageKey,
        namedArgs,
      });
    }, 0);
  }

  async function addBotToRoom(roomID) {
    try {
      refreshAuthToken();
      const token = authToken.value;

      if (!token) {
        throw new Error("Користувач не авторизований для додавання бота");
      }

      const response = await fetch(`/api/rooms/${roomID}/bot`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        }
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `Помилка сервера: ${response.status}`);
      }

      console.log(`[gameStore] Бота успішно додано в кімнату ${roomID}`);
      return true;
    } catch (err) {
      console.error('[gameStore] Помилка при додаванні бота:', err);
      // Прокидаємо помилку в реактивне поле, щоб GameErrorModal її вивів
      error.value = {
        code: 'ERR_ADD_BOT',
        message: err.message || 'Не вдалося додати бота'
      };
      return false;
    }
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
    hasReceivedState,
    stateVersion,
    latestTurnAlert,
    discardSequence,
    lastPlayedCardsByPlayer,
    handSlots,
    secondsLeft,
    refreshAuthToken,
    clearError,
    clearLog,
    clearRevealedData,
    leaveCurrentRoom,
    connectToHub,
    sendWSMessage,
    disconnect,
    addBotToRoom,
  };
});