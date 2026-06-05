// src/constants/cards.js

const getCardImg = (fileName) => {
  return new URL(`../assets/cards/${fileName}`, import.meta.url).href;
};

export const CARD_INFO = {
  0: { name: "Шпигун", type: "SPY", value: 0, color: "bg-gray-700", emoji: "🦅", desc: "Отримайте один бал по закінченні раунду.", targetType: "SELF", image: getCardImg("00_Spy.png") },
  1: { name: "Вартовий", type: "GUARD", value: 1, color: "bg-red-700", emoji: "⚔️", desc: "Вгадайте карту іншого гравця.", targetType: "OPPONENT", requiresGuess: true, image: getCardImg("01_Guard.png") },
  2: { name: "Священник", type: "PRIEST", value: 2, color: "bg-blue-600", emoji: "⛪", desc: "Подивіться руку іншого гравця.", targetType: "OPPONENT", image: getCardImg("02_Priest.png") },
  3: { name: "Барон", type: "BARON", value: 3, color: "bg-green-600", emoji: "🎭", desc: "Порівняйте карти з іншим гравцем.", targetType: "OPPONENT", image: getCardImg("03_Baron.png") },
  4: { name: "Служниця", type: "HANDMAID", value: 4, color: "bg-yellow-600", emoji: "🛡️", desc: "Захист від усіх ефектів до наступного ходу.", targetType: "SELF", image: getCardImg("04_Handmaid.png") },
  5: { name: "Принц", type: "PRINCE", value: 5, color: "bg-cyan-600", emoji: "👑", desc: "Оберіть гравця (можна себе), щоб він скинули карту.", targetType: "ANY", image: getCardImg("05_Prince.png") },
  6: { name: "Канцлер", type: "CHANCELLOR", value: 6, color: "bg-purple-600", emoji: "📜", desc: "Візьміть 2 карти, залиште 1, інші вниз колоди.", targetType: "SELF", image: getCardImg("06_Minister.png") },
  7: { name: "Король", type: "KING", value: 8, color: "bg-amber-700", emoji: "⚜️", desc: "Обміняйтеся картами з іншим гравцем.", targetType: "OPPONENT", image: getCardImg("07_King.png") },
  8: { name: "Графиня", type: "COUNTESS", value: 7, color: "bg-orange-500", emoji: "💃", desc: "Скиньте, якщо в руці є Принц або Король.", targetType: "SELF", image: getCardImg("08_Countess.png") },
  9: { name: "Принцеса", type: "PRINCESS", value: 9, color: "bg-pink-600", emoji: "👸", desc: "Якщо ви скинете цю карту — ви вилітаєте.", targetType: "SELF", image: getCardImg("09_Princess.png") },

  "SPY": { name: "Шпигун", value: 0, color: "bg-gray-700", emoji: "🦅", desc: "Отримайте один бал по закінченні раунду.", targetType: "SELF", image: getCardImg("00_Spy.png") },
  "GUARD": { name: "Вартовий", value: 1, color: "bg-red-700", emoji: "⚔️", desc: "Вгадайте карту іншого гравця.", targetType: "OPPONENT", requiresGuess: true, image: getCardImg("01_Guard.png") },
  "PRIEST": { name: "Священник", value: 2, color: "bg-blue-600", emoji: "⛪", desc: "Подивіться руку іншого гравця.", targetType: "OPPONENT", image: getCardImg("02_Priest.png") },
  "BARON": { name: "Барон", value: 3, color: "bg-green-600", emoji: "🎭", desc: "Порівняйте карти з іншим гравцем.", targetType: "OPPONENT", image: getCardImg("03_Baron.png") },
  "HANDMAID": { name: "Служниця", value: 4, color: "bg-yellow-600", emoji: "🛡️", desc: "Захист від усіх ефектів до наступного ходу.", targetType: "SELF", image: getCardImg("04_Handmaid.png") },
  "PRINCE": { name: "Принц", value: 5, color: "bg-cyan-600", emoji: "👑", desc: "Оберіть гравця (можна себе), щоб він скинули карту.", targetType: "ANY", image: getCardImg("05_Prince.png") },
  "CHANCELLOR": { name: "Канцлер", value: 6, color: "bg-purple-600", emoji: "📜", desc: "Візьміть 2 карти, залиште 1, інші вниз колоди.", targetType: "SELF", image: getCardImg("06_Minister.png") },
  "KING": { name: "Король", value: 7, border: "border-amber-600", color: "bg-amber-700", emoji: "⚜️", desc: "Обміняйтеся картами з іншим гравцем.", targetType: "OPPONENT", image: getCardImg("07_King.png") },
  "COUNTESS": { name: "Графиня", value: 8, color: "bg-orange-500", emoji: "💃", desc: "Скиньте, якщо в руці є Принц або Король.", targetType: "SELF", image: getCardImg("08_Countess.png") },
  "PRINCESS": { name: "Принцеса", value: 9, color: "bg-pink-600", emoji: "👸", desc: "Якщо ви скинете цю карту — ви вилітаєте.", targetType: "SELF", image: getCardImg("09_Princess.png") }
};