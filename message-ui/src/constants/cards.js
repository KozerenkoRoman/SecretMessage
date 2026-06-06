export const CARD_LOCALIZATION = {
  SPY: { name: "Шпигун", desc: "Отримайте один бал по закінченні раунду." },
  GUARD: { name: "Вартовий", desc: "Вгадайте карту іншого гравця." },
  PRIEST: { name: "Священник", desc: "Подивіться руку іншого гравця." },
  BARON: { name: "Барон", desc: "Порівняйте карти з іншим гравцем." },
  HANDMAID: { name: "Служниця", desc: "Захист від усіх ефектів до наступного ходу." },
  PRINCE: { name: "Принц", desc: "Оберіть гравця (можна себе), щоб він скинули карту." },
  CHANCELLOR: { name: "Канцлер", desc: "Візьміть 2 карти, залиште 1, інші вниз колоди." },
  KING: { name: "Король", desc: "Обміняйтеся картами з іншим гравцем." },
  COUNTESS: { name: "Графиня", desc: "Скиньте, якщо в руці є Принц або Король." },
  PRINCESS: { name: "Принцеса", desc: "Якщо ви скинете цю карту — ви вилітаєте." }
};

const getCardImg = (fileName) => {
  return new URL(`../assets/cards/${fileName}`, import.meta.url).href;
};

export const CARD_INFO_NUMBERS = {
  0: { name: CARD_LOCALIZATION.SPY.name, type: "SPY", value: 0, color: "bg-gray-700", emoji: "🦅", desc: CARD_LOCALIZATION.SPY.desc, targetType: "SELF", image: getCardImg("00_Spy.png") },
  1: { name: CARD_LOCALIZATION.GUARD.name, type: "GUARD", value: 1, color: "bg-red-700", emoji: "⚔️", desc: CARD_LOCALIZATION.GUARD.desc, targetType: "OPPONENT", requiresGuess: true, image: getCardImg("01_Guard.png") },
  2: { name: CARD_LOCALIZATION.PRIEST.name, type: "PRIEST", value: 2, color: "bg-blue-600", emoji: "⛪", desc: CARD_LOCALIZATION.PRIEST.desc, targetType: "OPPONENT", image: getCardImg("02_Priest.png") },
  3: { name: CARD_LOCALIZATION.BARON.name, type: "BARON", value: 3, color: "bg-green-600", emoji: "🎭", desc: CARD_LOCALIZATION.BARON.desc, targetType: "OPPONENT", image: getCardImg("03_Baron.png") },
  4: { name: CARD_LOCALIZATION.HANDMAID.name, type: "HANDMAID", value: 4, color: "bg-yellow-600", emoji: "🛡️", desc: CARD_LOCALIZATION.HANDMAID.desc, targetType: "SELF", image: getCardImg("04_Handmaid.png") },
  5: { name: CARD_LOCALIZATION.PRINCE.name, type: "PRINCE", value: 5, color: "bg-cyan-600", emoji: "👑", desc: CARD_LOCALIZATION.PRINCE.desc, targetType: "ANY", image: getCardImg("05_Prince.png") },
  6: { name: CARD_LOCALIZATION.CHANCELLOR.name, type: "CHANCELLOR", value: 6, color: "bg-purple-600", emoji: "📜", desc: CARD_LOCALIZATION.CHANCELLOR.desc, targetType: "SELF", image: getCardImg("06_Minister.png") },
  7: { name: CARD_LOCALIZATION.KING.name, type: "KING", value: 8, color: "bg-amber-700", emoji: "⚜️", desc: CARD_LOCALIZATION.KING.desc, targetType: "OPPONENT", image: getCardImg("07_King.png") },
  8: { name: CARD_LOCALIZATION.COUNTESS.name, type: "COUNTESS", value: 7, color: "bg-orange-500", emoji: "💃", desc: CARD_LOCALIZATION.COUNTESS.desc, targetType: "SELF", image: getCardImg("08_Countess.png") },
  9: { name: CARD_LOCALIZATION.PRINCESS.name, type: "PRINCESS", value: 9, color: "bg-pink-600", emoji: "👸", desc: CARD_LOCALIZATION.PRINCESS.desc, targetType: "SELF", image: getCardImg("09_Princess.png") }
}

export const CARD_INFO_NAMES = {
  "SPY": { name: CARD_LOCALIZATION.SPY.name, value: 0, color: "bg-gray-700", emoji: "🦅", desc: CARD_LOCALIZATION.SPY.desc, targetType: "SELF", image: getCardImg("00_Spy.png") },
  "GUARD": { name: CARD_LOCALIZATION.GUARD.name, value: 1, color: "bg-red-700", emoji: "⚔️", desc: CARD_LOCALIZATION.GUARD.desc, targetType: "OPPONENT", requiresGuess: true, image: getCardImg("01_Guard.png") },
  "PRIEST": { name: CARD_LOCALIZATION.PRIEST.name, value: 2, color: "bg-blue-600", emoji: "⛪", desc: CARD_LOCALIZATION.PRIEST.desc, targetType: "OPPONENT", image: getCardImg("02_Priest.png") },
  "BARON": { name: CARD_LOCALIZATION.BARON.name, value: 3, color: "bg-green-600", emoji: "🎭", desc: CARD_LOCALIZATION.BARON.desc, targetType: "OPPONENT", image: getCardImg("03_Baron.png") },
  "HANDMAID": { name: CARD_LOCALIZATION.HANDMAID.name, value: 4, color: "bg-yellow-600", emoji: "🛡️", desc: CARD_LOCALIZATION.HANDMAID.desc, targetType: "SELF", image: getCardImg("04_Handmaid.png") },
  "PRINCE": { name: CARD_LOCALIZATION.PRINCE.name, value: 5, color: "bg-cyan-600", emoji: "👑", desc: CARD_LOCALIZATION.PRINCE.desc, targetType: "ANY", image: getCardImg("05_Prince.png") },
  "CHANCELLOR": { name: CARD_LOCALIZATION.CHANCELLOR.name, value: 6, color: "bg-purple-600", emoji: "📜", desc: CARD_LOCALIZATION.CHANCELLOR.desc, targetType: "SELF", image: getCardImg("06_Minister.png") },
  "KING": { name: CARD_LOCALIZATION.KING.name, value: 7, border: "border-amber-600", color: "bg-amber-700", emoji: "⚜️", desc: CARD_LOCALIZATION.KING.desc, targetType: "OPPONENT", image: getCardImg("07_King.png") },
  "COUNTESS": { name: CARD_LOCALIZATION.COUNTESS.name, value: 8, color: "bg-orange-500", emoji: "💃", desc: CARD_LOCALIZATION.COUNTESS.desc, targetType: "SELF", image: getCardImg("08_Countess.png") },
  "PRINCESS": { name: CARD_LOCALIZATION.PRINCESS.name, value: 9, color: "bg-pink-600", emoji: "👸", desc: CARD_LOCALIZATION.PRINCESS.desc, targetType: "SELF", image: getCardImg("09_Princess.png") }
};

export const getCardInfoHelper = (key) => {
  if (key === undefined || key === null) return null;
  return (!isNaN(key) && key !== "") ? CARD_INFO_NUMBERS[Number(key)] : CARD_INFO_NAMES[String(key)];
};