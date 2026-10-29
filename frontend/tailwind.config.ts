import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      // The TUI's palette (tui/render.go), as desktop/Sources/Skyra/Sky/Theme.swift has it.
      colors: {
        sky: "#080a0e",
        panel: "#0f141d",
        ink: {
          DEFAULT: "#d6dbe5",
          dim: "#788092",
          faint: "#3c4250",
        },
        accent: "#7ec4cc",
        green: "#60c4a6",
        // A refused call: the Boundary doing its job, not a failure.
        amber: "#dea868",
        // Something broke: the backend, custody.
        fault: "#e27070",
      },
      fontFamily: {
        mono: ["ui-monospace", "SFMono-Regular", "Menlo", "monospace"],
      },
    },
  },
  plugins: [],
};

export default config;
