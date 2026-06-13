export function getAvatarUrl(seed) {
    if (!seed) {
        return `https://api.dicebear.com/9.x/avataaars/svg?seed=fallback`;
    }
    return `https://api.dicebear.com/9.x/avataaars/svg?seed=${encodeURIComponent(seed)}`;
}

export function generateRandomSeed() {
    return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
}