# Theme

## Compact token summary

| Token | Light | Dark | Purpose |
|---|---:|---:|---|
| `background` | `#ffffff` | `#1a1210` | App background |
| `foreground` | `#231f20` | `#ffeac5` | Default text |
| `surface` | `#ffeac5` | `#2a1f19` | Cards and raised surfaces |
| `primary` | `#603f26` | `#e0b988` | Actions and strong titles |
| `accent` | `#c49a6c` | `#c49a6c` | Secondary brand detail |
| `highlight` | `#fedbb5` | `#fedbb5` | Hover and selected tints |
| `ink` | `#231f20` | `#e8dccb` | Borders and icon detail |

- Body face: Inter through `next/font/google`, exposed as `font-sans`.
- Display face: Anton 400 through `next/font/google`, exposed as `font-display`.
- Shape: friendly `rounded-xl`/`rounded-2xl`; flat surfaces and minimal shadows.
- Spacing and breakpoints: Tailwind defaults, mobile-first.
- Theme mode: automatic `prefers-color-scheme`; no manual toggle.
- Brand mood: warm brown/cream, simplistic and cute-looking while professional.

## Raw source

### `src/app/globals.css`

```css
@import "tailwindcss";

:root {
  --background: #ffffff;
  --foreground: #231f20;
  --surface: #ffeac5;
  --primary: #603f26;
  --accent: #c49a6c;
  --highlight: #fedbb5;
  --ink: #231f20;
}

@theme inline {
  --color-background: var(--background);
  --color-foreground: var(--foreground);
  --color-surface: var(--surface);
  --color-primary: var(--primary);
  --color-accent: var(--accent);
  --color-highlight: var(--highlight);
  --color-ink: var(--ink);
  --font-sans: var(--font-inter);
  --font-display: var(--font-anton);
}

@media (prefers-color-scheme: dark) {
  :root {
    --background: #1a1210;
    --foreground: #ffeac5;
    --surface: #2a1f19;
    --primary: #e0b988;
    --accent: #c49a6c;
    --highlight: #fedbb5;
    --ink: #e8dccb;
  }
}

body {
  background: var(--background);
  color: var(--foreground);
  font-family: var(--font-sans), Arial, Helvetica, sans-serif;
}
```

### `postcss.config.mjs`

```js
const config = {
  plugins: {
    "@tailwindcss/postcss": {},
  },
};

export default config;
```

There is no `tailwind.config.*`; Tailwind v4 tokens live in `globals.css`.
