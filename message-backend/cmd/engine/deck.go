// =============================================================================
// engine/deck.go
//
// Чиста, детермінована побудова колоди для одного раунду.
//
// Цей файл є ЄДИНИМ джерелом правди про склад колоди "Secret Message".
// Він замінює дубльовану логіку, яка раніше жила одночасно у room.StartGame()
// та room.NextRound() (16 карт + 6 Guard, перемішування Fisher-Yates,
// одна "burn"-карта зверху).
//
// Дизайн:
//   • Функція повертає вже перемішану колоду без burn-карти + саму burn-карту.
//   • RNG передається явно — це робить функцію тестованою та детермінованою
//     при заміні реального rand на mockRNG.
//   • Колода будується через `make([]CardType, 0, deckTotal)` — рівно
//     одна аллокація бекінг-масиву на весь deck (без append-grow).
//   • Не залежить від `GameState`, тому її можна викликати з будь-якого шару
//     (room, AI-симулятор, replay-тести).
// =============================================================================

package engine

// totalDeckSize — канонічний розмір колоди Secret Message:
//
//	1 Princess + 1 King + 1 Countess + 2 Chancellor + 2 Prince +
//	2 Baron + 2 Handmaid + 2 Priest + 2 Spy + 6 Guard = 21 карта.
//
// (Це точне відображення складу, який використовувався у старих
// дублях логіки в room.StartGame()/room.NextRound() — тут ми НЕ змінюємо
// баланс гри, лише виносимо логіку в одне місце.)
const totalDeckSize = 21

// composedDeckTemplate описує мультимножину карт у нерозтасованій колоді.
// Виноситься у package-level масив, щоб не виділяти 21 значення при кожному
// раунді. УВАГА: у PrepareNewDeck ми ОБОВ'ЯЗКОВО копіюємо цей шаблон,
// тому модифікації під час перемішування не торкаються константи.
var composedDeckTemplate = [...]CardType{
	CardPrincess,
	CardKing,
	CardCountess,
	CardChancellor, CardChancellor,
	CardPrince, CardPrince,
	CardBaron, CardBaron,
	CardHandmaid, CardHandmaid,
	CardPriest, CardPriest,
	CardSpy, CardSpy,
	CardGuard, CardGuard, CardGuard, CardGuard, CardGuard, CardGuard,
}

// PrepareNewDeck будує і перемішує нову колоду одного раунду гри.
//
// Результат:
//
//	deck — слайс з (totalDeckSize - 1) карт, готовий до роздачі гравцям;
//	       індекс 0 — це "верх" колоди (звідки добирають карти).
//	burn — одна карта, яка прибирається з гри (state.BurnCard).
//
// Алгоритм:
//  1. Копіюємо незмінний шаблон у новий слайс (одна аллокація).
//  2. Перемішуємо in-place алгоритмом Fisher–Yates із поданим RNG.
//  3. Беремо першу карту як burn, інше повертаємо як deck.
//
// Складність: O(n), n = totalDeckSize.
//
// Викликати ТРЕБА під час старту партії (StartGame) та на початку кожного
// нового раунду (NextRound). Сигнатура зумисно мінімальна: ніяких
// побічних ефектів, ніяких залежностей від Room/GameState.
func PrepareNewDeck(rng RNG) (deck []CardType, burn CardType) {
	// Одна аллокація на 22 елементи; capacity == length, бо ми точно знаємо
	// фінальний розмір. Це спрощує життя GC порівняно з append-grow.
	cards := make([]CardType, totalDeckSize)
	copy(cards, composedDeckTemplate[:])

	// Fisher–Yates. Йдемо з кінця до 1; rng.Intn(i+1) дає рівномірний
	// розподіл по [0, i].
	for i := len(cards) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}

	// Burn — це одна карта зверху колоди (як у правилах Love Letter).
	// Решта йде у deck. Слайс [1:] не алокує — лише змінює header.
	burn = cards[0]
	deck = cards[1:]
	return deck, burn
}
