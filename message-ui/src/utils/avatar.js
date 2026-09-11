/**
 * getAvatarUrl - повертає URL зображення аватара для рендеру.
 *
 * Приймає або:
 *   - рядок (avatar_seed) - зворотньо сумісний виклик, повертає DiceBear URL;
 *   - об'єкт { avatar_url, avatar_seed } - якщо avatar_url заповнений
 *     (користувач завантажив власне зображення), він має пріоритет над
 *     процедурно згенерованим DiceBear-аватаром за seed.
 */
export function getAvatarUrl(seedOrPlayer) {
    if (seedOrPlayer && typeof seedOrPlayer === 'object') {
        if (seedOrPlayer.avatar_url) {
            return seedOrPlayer.avatar_url;
        }
        return getAvatarUrl(seedOrPlayer.avatar_seed);
    }

    const seed = seedOrPlayer;
    if (!seed) {
        return `https://api.dicebear.com/9.x/avataaars/svg?seed=fallback`;
    }
    return `https://api.dicebear.com/9.x/avataaars/svg?seed=${encodeURIComponent(seed)}`;
}

export function generateRandomSeed() {
    return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
}