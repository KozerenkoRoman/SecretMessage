export const CARD_I18N_KEYS = {
  SPY: { name: "cards.SPY.name", desc: "cards.SPY.desc" },
  GUARD: { name: "cards.GUARD.name", desc: "cards.GUARD.desc" },
  PRIEST: { name: "cards.PRIEST.name", desc: "cards.PRIEST.desc" },
  BARON: { name: "cards.BARON.name", desc: "cards.BARON.desc" },
  HANDMAID: { name: "cards.HANDMAID.name", desc: "cards.HANDMAID.desc" },
  PRINCE: { name: "cards.PRINCE.name", desc: "cards.PRINCE.desc" },
  CHANCELLOR: { name: "cards.CHANCELLOR.name", desc: "cards.CHANCELLOR.desc" },
  KING: { name: "cards.KING.name", desc: "cards.KING.desc" },
  COUNTESS: { name: "cards.COUNTESS.name", desc: "cards.COUNTESS.desc" },
  PRINCESS: { name: "cards.PRINCESS.name", desc: "cards.PRINCESS.desc" }
};

const getCardImg = (fileName) => {
  return new URL(`../assets/cards/${fileName}`, import.meta.url).href;
};

export const CARD_INFO_NUMBERS = {
  0: { nameKey: CARD_I18N_KEYS.SPY.name, descKey: CARD_I18N_KEYS.SPY.desc, type: "SPY", value: 0, color: "bg-gray-700", emoji: "🦅", targetType: "SELF", image: getCardImg("00_Spy.png") },
  1: { nameKey: CARD_I18N_KEYS.GUARD.name, descKey: CARD_I18N_KEYS.GUARD.desc, type: "GUARD", value: 1, color: "bg-red-700", emoji: "⚔️", targetType: "OPPONENT", requiresGuess: true, image: getCardImg("01_Guard.png") },
  2: { nameKey: CARD_I18N_KEYS.PRIEST.name, descKey: CARD_I18N_KEYS.PRIEST.desc, type: "PRIEST", value: 2, color: "bg-blue-600", emoji: "⛪", targetType: "OPPONENT", image: getCardImg("02_Priest.png") },
  3: { nameKey: CARD_I18N_KEYS.BARON.name, descKey: CARD_I18N_KEYS.BARON.desc, type: "BARON", value: 3, color: "bg-green-600", emoji: "🎭", targetType: "OPPONENT", image: getCardImg("03_Baron.png") },
  4: { nameKey: CARD_I18N_KEYS.HANDMAID.name, descKey: CARD_I18N_KEYS.HANDMAID.desc, type: "HANDMAID", value: 4, color: "bg-yellow-600", emoji: "🛡️", targetType: "SELF", image: getCardImg("04_Handmaid.png") },
  5: { nameKey: CARD_I18N_KEYS.PRINCE.name, descKey: CARD_I18N_KEYS.PRINCE.desc, type: "PRINCE", value: 5, color: "bg-cyan-600", emoji: "👑", targetType: "ANY", image: getCardImg("05_Prince.png") },
  6: { nameKey: CARD_I18N_KEYS.CHANCELLOR.name, descKey: CARD_I18N_KEYS.CHANCELLOR.desc, type: "CHANCELLOR", value: 6, color: "bg-purple-600", emoji: "📜", targetType: "SELF", image: getCardImg("06_Minister.png") },
  7: { nameKey: CARD_I18N_KEYS.KING.name, descKey: CARD_I18N_KEYS.KING.desc, type: "KING", value: 8, color: "bg-amber-700", emoji: "⚜️", targetType: "OPPONENT", image: getCardImg("07_King.png") },
  8: { nameKey: CARD_I18N_KEYS.COUNTESS.name, descKey: CARD_I18N_KEYS.COUNTESS.desc, type: "COUNTESS", value: 7, color: "bg-orange-500", emoji: "💃", targetType: "SELF", image: getCardImg("08_Countess.png") },
  9: { nameKey: CARD_I18N_KEYS.PRINCESS.name, descKey: CARD_I18N_KEYS.PRINCESS.desc, type: "PRINCESS", value: 9, color: "bg-pink-600", emoji: "👸", targetType: "SELF", image: getCardImg("09_Princess.png") }
}

export const CARD_INFO_NAMES = {
  "SPY": { nameKey: CARD_I18N_KEYS.SPY.name, descKey: CARD_I18N_KEYS.SPY.desc, type: "SPY", value: 0, color: "bg-gray-700", emoji: "🦅", targetType: "SELF", image: getCardImg("00_Spy.png") },
  "GUARD": { nameKey: CARD_I18N_KEYS.GUARD.name, descKey: CARD_I18N_KEYS.GUARD.desc, type: "GUARD", value: 1, color: "bg-red-700", emoji: "⚔️", targetType: "OPPONENT", requiresGuess: true, image: getCardImg("01_Guard.png") },
  "PRIEST": { nameKey: CARD_I18N_KEYS.PRIEST.name, descKey: CARD_I18N_KEYS.PRIEST.desc, type: "PRIEST", value: 2, color: "bg-blue-600", emoji: "⛪", targetType: "OPPONENT", image: getCardImg("02_Priest.png") },
  "BARON": { nameKey: CARD_I18N_KEYS.BARON.name, descKey: CARD_I18N_KEYS.BARON.desc, type: "BARON", value: 3, color: "bg-green-600", emoji: "🎭", targetType: "OPPONENT", image: getCardImg("03_Baron.png") },
  "HANDMAID": { nameKey: CARD_I18N_KEYS.HANDMAID.name, descKey: CARD_I18N_KEYS.HANDMAID.desc, type: "HANDMAID", value: 4, color: "bg-yellow-600", emoji: "🛡️", targetType: "SELF", image: getCardImg("04_Handmaid.png") },
  "PRINCE": { nameKey: CARD_I18N_KEYS.PRINCE.name, descKey: CARD_I18N_KEYS.PRINCE.desc, type: "PRINCE", value: 5, color: "bg-cyan-600", emoji: "👑", targetType: "ANY", image: getCardImg("05_Prince.png") },
  "CHANCELLOR": { nameKey: CARD_I18N_KEYS.CHANCELLOR.name, descKey: CARD_I18N_KEYS.CHANCELLOR.desc, type: "CHANCELLOR", value: 6, color: "bg-purple-600", emoji: "📜", targetType: "SELF", image: getCardImg("06_Minister.png") },
  "KING": { nameKey: CARD_I18N_KEYS.KING.name, descKey: CARD_I18N_KEYS.KING.desc, type: "KING", value: 7, border: "border-amber-600", color: "bg-amber-700", emoji: "⚜️", targetType: "OPPONENT", image: getCardImg("07_King.png") },
  "COUNTESS": { nameKey: CARD_I18N_KEYS.COUNTESS.name, descKey: CARD_I18N_KEYS.COUNTESS.desc, type: "COUNTESS", value: 8, color: "bg-orange-500", emoji: "💃", targetType: "SELF", image: getCardImg("08_Countess.png") },
  "PRINCESS": { nameKey: CARD_I18N_KEYS.PRINCESS.name, descKey: CARD_I18N_KEYS.PRINCESS.desc, type: "PRINCESS", value: 9, color: "bg-pink-600", emoji: "👸", targetType: "SELF", image: getCardImg("09_Princess.png") }
};

const resolveCard = (raw, t) => {
  if (!raw) return raw;
  if (typeof t !== "function") return raw;
  return {
    ...raw,
    name: t(raw.nameKey),
    desc: t(raw.descKey),
  };
};

export const getCardInfoHelper = (key, t) => {
  if (key === undefined || key === null) return null;
  const raw = (!isNaN(key) && key !== "") ? CARD_INFO_NUMBERS[Number(key)] : CARD_INFO_NAMES[String(key)];
  return resolveCard(raw, t);
};
