
/**
 * Поверхнево-рекурсивне in-place злиття серверного state у локальний.
 * - Примітиви: присвоюємо лише коли значення відрізняється.
 * - Масиви: якщо довжина і кожен елемент-примітив рівні — НЕ чіпаємо
 *   існуюче посилання (ключовий момент для стабільної reactivity та
 *   уникання небажаних DOM-перебудов / blink-ефектів). Інакше мутуємо
 *   через splice, зберігаючи саме посилання на масив.
 * - Об'єкти: рекурсивно зливаємо ключі; видаляємо ті, яких більше нема.
 *
 * Повертає той самий target (для зручності).
 */
export function mergeState(target, source) {
    if (!target || typeof target !== 'object') return source;
    if (!source || typeof source !== 'object') return target;

    if (Array.isArray(source)) {
        if (!Array.isArray(target)) return source;
        return mergeArray(target, source);
    }

    // Видаляємо ключі, яких більше нема в новому стані
    for (const k of Object.keys(target)) {
        if (!(k in source)) delete target[k];
    }

    for (const k of Object.keys(source)) {
        const sv = source[k];
        const tv = target[k];

        if (sv === null || typeof sv !== 'object') {
            if (tv !== sv) target[k] = sv;
            continue;
        }

        if (Array.isArray(sv)) {
            if (Array.isArray(tv)) {
                mergeArray(tv, sv);
            } else {
                target[k] = sv.slice();
            }
            continue;
        }

        if (tv && typeof tv === 'object' && !Array.isArray(tv)) {
            mergeState(tv, sv);
        } else {
            target[k] = mergeState({}, sv);
        }
    }
    return target;
}

export function mergeArray(target, source) {
    // Швидкий шлях: однакова довжина та однакові примітиви/ID-збіги -
    // нічого не робимо, посилання зберігається.
    if (target.length === source.length) {
        let identical = true;
        for (let i = 0; i < source.length; i++) {
            const a = target[i];
            const b = source[i];
            if (a === b) continue;
            if (a && b && typeof a === 'object' && typeof b === 'object') {
                identical = false; // далі смержимо поелементно
                break;
            }
            identical = false;
            break;
        }
        if (identical) return target;
    }

    // Узгоджуємо довжину одним splice — це краще, ніж створювати новий масив.
    if (target.length > source.length) {
        target.splice(source.length, target.length - source.length);
    }
    for (let i = 0; i < source.length; i++) {
        const sv = source[i];
        const tv = target[i];
        if (sv === null || typeof sv !== 'object') {
            if (tv !== sv) target[i] = sv;
        } else if (Array.isArray(sv)) {
            if (Array.isArray(tv)) mergeArray(tv, sv);
            else target[i] = sv.slice();
        } else if (tv && typeof tv === 'object' && !Array.isArray(tv)) {
            mergeState(tv, sv);
        } else {
            target[i] = mergeState({}, sv);
        }
    }
    return target;
}