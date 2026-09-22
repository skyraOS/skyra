import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        // Three-region palette. Tweak in week 2 once we see them in motion.
        region: {
          user: "#1a2332",
          overlap: "#2d2438",
          being: "#1f1a2e",
        },
        logos: {
          world: "#5b8def",
          being: "#a78bfa",
          operator: "#34d399",
        },
        ink: {
          DEFAULT: "#e6e8ee",
          dim: "#8a8fa0",
          faint: "#4a4f5e",
        },
        surface: {
          DEFAULT: "#0b0d12",
          raised: "#141821",
          edge: "#222632",
        },
      },
      fontFamily: {
        mono: ["ui-monospace", "SFMono-Regular", "Menlo", "monospace"],
      },
    },
  },
  plugins: [],
};

export default config;
