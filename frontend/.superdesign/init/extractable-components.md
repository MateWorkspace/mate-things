# Extractable components

## Layout components

No reusable authenticated app shell exists yet. The planned `AppShell`, `AppBar`, `Sidebar`, permission-filtered navigation group, profile dialog, breadcrumb, and mobile drawer should become extractable layout components once implemented.

## Button

- Source: `src/components/ui/button.tsx`
- Category: basic
- Description: Rounded branded action button with primary and secondary variants.
- Extractable props: `variant`, `disabled`
- Hardcoded: radius, padding, brand colors, hover/active/focus treatment

## Input

- Source: `src/components/ui/input.tsx`
- Category: basic
- Description: Rounded text input with branded border and focus ring.
- Extractable props: native input value/state props
- Hardcoded: radius, padding, colors and focus treatment

## Label

- Source: `src/components/ui/label.tsx`
- Category: basic
- Description: Compact, medium-weight form label.
- Extractable props: none beyond native label props
- Hardcoded: typography, margin and muted foreground color

## Toast

- Source: `src/components/ui/toast.tsx`
- Category: basic
- Description: Animated success/error notification card with dismiss and progress.
- Extractable props: `variant`, `title`, `message`, `leaving`
- Hardcoded: icons, timing and visual treatment

## ToastProvider

- Source: `src/components/ui/toast-provider.tsx`
- Category: layout
- Description: Global notification context and fixed top-right toast stack.
- Extractable props: none
- Hardcoded: stack position, timing and supported variants
