/**
 * Tailwind CSS configuration with semantic colour palette.
 */
module.exports = {
  content: [
    "./src/**/*.vue",
    "./src/**/*.js",
    "./src/**/*.ts",
  ],
  theme: {
    extend: {
      colors: {
        "brand-bg": "#0f172a",
        "brand-bg-dark": "#020617",
        "brand-surface": "#1e293b",
        "brand-surface-dim": "#334155",
        "brand-text-primary": "#f1f5f9",
        "brand-text-muted": "#e2e8f0",
        "brand-text-subtle": "#cbd5e1",
        "brand-border": "#334155",
        "brand-accent": "#f59e0b",
        "brand-accent-light": "#fbbf24",
        "brand-warning": "#eab308",
        "brand-info": "#3b82f6",
        "brand-success": "#10b981",
        "brand-danger": "#ef4444",
        // Card colours
        "brand-card-gray": "#374151",
        "brand-card-red": "#b91c1c",
        "brand-card-blue": "#2563eb",
        "brand-card-green": "#16a34a",
        "brand-card-yellow": "#ca8a04",
        "brand-card-cyan": "#0891b2",
        "brand-card-purple": "#7c3aed",
        "brand-card-amber": "#b45309",
        "brand-card-orange": "#f97316",
        "brand-card-pink": "#d946ef",
      },
    },
  },
  plugins: [],
};
