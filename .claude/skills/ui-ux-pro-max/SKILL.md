# UI/UX Pro Max - Design Intelligence (Mantine v9 Edition)

Comprehensive design guide for web applications built with **Mantine v9** and React/TypeScript. Contains 50+ styles, 161 color palettes, 57 font pairings, 161 product types with reasoning rules, 99 UX guidelines, and 25 chart types. Searchable database with priority-based recommendations.

**Tailored for FindFore:** Golf social platform using React 19 + TypeScript + Mantine v9 + Vite, mobile-first PWA.

## When to Apply

This Skill should be used when the task involves **UI structure, visual design decisions, interaction patterns, or user experience quality control**.

### Must Use

- Designing new pages or screens (feed, dashboard, profile, tee time flows)
- Creating or refactoring Mantine UI components (buttons, modals, forms, cards, etc.)
- Choosing color schemes, typography systems, spacing standards, or layout systems
- Reviewing UI code for UX, accessibility, or visual consistency
- Implementing navigation structures, animations, or responsive behavior
- Making product-level design decisions (style, information hierarchy, brand expression)
- Improving perceived quality, clarity, or usability of interfaces

### Recommended

- UI looks "not professional enough" but the reason is unclear
- Receiving feedback on usability or experience
- Pre-launch UI quality optimization
- Building reusable Mantine component compositions or theme extensions
- Dark mode refinement

### Skip

- Pure backend logic (Go handlers, sqlc queries)
- API or database design only
- Performance optimization unrelated to the interface
- Infrastructure, DevOps, or deployment work

**Decision criteria**: If the task will change how a feature **looks, feels, moves, or is interacted with**, this Skill should be used.

## Rule Categories by Priority

| Priority | Category | Impact | Domain | Key Checks (Must Have) | Anti-Patterns (Avoid) |
|----------|----------|--------|--------|------------------------|------------------------|
| 1 | Accessibility | CRITICAL | `ux` | Contrast 4.5:1, Alt text, Keyboard nav, Aria-labels | Removing focus rings, Icon-only buttons without labels |
| 2 | Touch & Interaction | CRITICAL | `ux` | Min size 44x44px, 8px+ spacing, Loading feedback | Reliance on hover only, Instant state changes (0ms) |
| 3 | Performance | HIGH | `ux` | WebP/AVIF, Lazy loading, Reserve space (CLS < 0.1) | Layout thrashing, Cumulative Layout Shift |
| 4 | Style Selection | HIGH | `style`, `product` | Match product type, Consistency, SVG icons (no emoji) | Mixing flat & skeuomorphic randomly, Emoji as icons |
| 5 | Layout & Responsive | HIGH | `ux` | Mobile-first breakpoints, Viewport meta, No horizontal scroll | Horizontal scroll, Fixed px container widths, Disable zoom |
| 6 | Typography & Color | MEDIUM | `typography`, `color` | Base 16px, Line-height 1.5, Semantic color tokens | Text < 12px body, Gray-on-gray, Raw hex in components |
| 7 | Animation | MEDIUM | `ux` | Duration 150-300ms, Motion conveys meaning, Spatial continuity | Decorative-only animation, Animating width/height, No reduced-motion |
| 8 | Forms & Feedback | MEDIUM | `ux` | Visible labels, Error near field, Helper text, Progressive disclosure | Placeholder-only label, Errors only at top, Overwhelm upfront |
| 9 | Navigation Patterns | HIGH | `ux` | Predictable back, Bottom nav <=5, Deep linking | Overloaded nav, Broken back behavior, No deep links |
| 10 | Charts & Data | LOW | `chart` | Legends, Tooltips, Accessible colors | Relying on color alone to convey meaning |

## Mantine v9 Implementation Notes

When implementing designs from this skill's recommendations, translate to Mantine v9 patterns:

### Theme & Color Tokens

- Use `MantineProvider` with a custom theme object for all branding
- Define colors as Mantine color tuples (10 shades per color) in `theme.colors`
- Use `primaryColor`, `primaryShade` for the main brand color
- Use CSS variables: `var(--mantine-color-*)` for semantic tokens
- Dark mode via `colorScheme` prop on `MantineProvider` and `useMantineColorScheme()`
- Never hardcode hex values in components -- always reference theme tokens

### Components (Mantine v9 Equivalents)

| Design Concept | Mantine v9 Component |
|---|---|
| Cards / Surfaces | `Card`, `Paper`, `Surface` |
| Navigation (bottom) | `AppShell.Navbar` + custom bottom nav |
| Modal / Sheet | `Modal`, `Drawer` |
| Forms | `TextInput`, `Select`, `Checkbox`, `Radio`, `Switch` |
| Buttons / CTAs | `Button`, `ActionIcon`, `UnstyledButton` |
| Feedback / Toast | `Notifications` (from `@mantine/notifications`) |
| Loading states | `Skeleton`, `LoadingOverlay`, `Progress` |
| Data display | `Table`, `Timeline`, `Badge`, `Avatar` |
| Layout | `AppShell`, `Container`, `Grid`, `SimpleGrid`, `Stack`, `Group`, `Flex` |
| Typography | `Title`, `Text`, `Highlight` |
| Overlay | `Overlay`, `Tooltip`, `Popover`, `HoverCard` |

### Spacing & Layout

- Use Mantine's spacing scale: `xs`, `sm`, `md`, `lg`, `xl` (maps to 10/12/16/20/32 by default)
- Customize via `theme.spacing` for the 4/8dp rhythm system
- Use `Container` with `size` prop for consistent content width
- Use `AppShell` for page-level layout with responsive navbar/header

### Responsive Design

- Use Mantine's responsive props: `visibleFrom`, `hiddenFrom`
- Breakpoints defined in `theme.breakpoints` (xs: 576, sm: 768, md: 992, lg: 1200, xl: 1408)
- Use `useMediaQuery` hook from `@mantine/hooks` for JS-level responsiveness
- `SimpleGrid` with `cols` object for responsive grids: `cols={{ base: 1, sm: 2, lg: 3 }}`

### Dark Mode

- Use `useMantineColorScheme()` to toggle
- Light/dark variants in theme via `theme.other` or CSS module `.light` / `.dark`
- Test both modes -- Mantine auto-switches `color` and `backgroundColor` but custom styles need manual handling
- Use `light-dark()` CSS function or `[data-mantine-color-scheme="dark"]` selectors

### Accessibility in Mantine

- Mantine components have built-in accessibility (focus rings, aria attributes)
- Use `VisuallyHidden` for screen-reader-only text
- `FocusTrap` for modal focus management
- All form components support `label`, `description`, `error` props natively
- Use `aria-label` on `ActionIcon` components (icon-only buttons)

## How to Use

### Prerequisites

```bash
python3 --version
```

### Step 1: Analyze Requirements

Extract from the user request:
- **Product type**: Social, entertainment, tool, productivity, or hybrid
- **Target audience**: Golfers, mobile users, specific demographics
- **Style keywords**: golf-inspired, modern, dark mode, community, etc.
- **Stack**: React + Mantine v9 (always for this project)

### Step 2: Generate Design System

```bash
python3 .claude/skills/ui-ux-pro-max/scripts/search.py "<product_type> <industry> <keywords>" --design-system [-p "Project Name"]
```

Example for FindFore:
```bash
python3 .claude/skills/ui-ux-pro-max/scripts/search.py "golf social community mobile-first" --design-system -p "FindFore"
```

### Step 2b: Persist Design System

```bash
python3 .claude/skills/ui-ux-pro-max/scripts/search.py "<query>" --design-system --persist -p "FindFore"
```

With page-specific overrides:
```bash
python3 .claude/skills/ui-ux-pro-max/scripts/search.py "<query>" --design-system --persist -p "FindFore" --page "tee-time-create"
```

### Step 3: Domain-Specific Searches

```bash
python3 .claude/skills/ui-ux-pro-max/scripts/search.py "<keyword>" --domain <domain> [-n <max_results>]
```

| Need | Domain | Example |
|------|--------|---------|
| Product type patterns | `product` | `--domain product "social golf community"` |
| More style options | `style` | `--domain style "glassmorphism dark"` |
| Color palettes | `color` | `--domain color "social vibrant green"` |
| Font pairings | `typography` | `--domain typography "modern sporty"` |
| Chart recommendations | `chart` | `--domain chart "real-time dashboard"` |
| UX best practices | `ux` | `--domain ux "animation accessibility"` |
| Individual Google Fonts | `google-fonts` | `--domain google-fonts "sans serif modern"` |
| Landing structure | `landing` | `--domain landing "hero social-proof"` |
| React performance | `react` | `--domain react "rerender memo list"` |
| Web accessibility | `web` | `--domain web "aria focus form"` |

### Step 4: Stack Guidelines (React)

```bash
python3 .claude/skills/ui-ux-pro-max/scripts/search.py "<keyword>" --stack react
```

Then translate React guidance into Mantine v9 component patterns.

## Quick Reference

### 1. Accessibility (CRITICAL)

- `color-contrast` - Minimum 4.5:1 ratio for normal text; use Mantine's built-in theme contrast
- `focus-states` - Mantine provides focus rings by default; never override `theme.focusRing`
- `alt-text` - Descriptive alt text for meaningful images
- `aria-labels` - `aria-label` on all `ActionIcon` components (icon-only buttons)
- `keyboard-nav` - Tab order matches visual order; Mantine handles this for most components
- `form-labels` - Use Mantine form component `label` prop, never placeholder-only
- `skip-links` - Skip to main content for keyboard users
- `heading-hierarchy` - Use Mantine `Title` with sequential `order` props (1-6)
- `color-not-only` - Don't convey info by color alone (add icon/text)
- `reduced-motion` - Respect `prefers-reduced-motion`; Mantine `Transition` handles this
- `escape-routes` - Mantine `Modal`/`Drawer` provide close on Escape by default

### 2. Touch & Interaction (CRITICAL)

- `touch-target-size` - Min 44x44px; use Mantine `size` prop (sm/md/lg) to ensure adequate targets
- `touch-spacing` - Minimum 8px gap; use Mantine `spacing` scale between interactive elements
- `hover-vs-tap` - Use click/tap for primary interactions; don't rely on `HoverCard` for critical info
- `loading-buttons` - Use Mantine `Button` `loading` prop during async operations
- `error-feedback` - Use Mantine form `error` prop on inputs for inline error display
- `cursor-pointer` - Mantine buttons handle this; add `style={{ cursor: 'pointer' }}` on custom clickables
- `safe-area-awareness` - Keep primary touch targets away from notch and gesture bar areas (PWA)

### 3. Performance (HIGH)

- `image-optimization` - Use WebP/AVIF, lazy load with `loading="lazy"`, use Mantine `Image` component
- `image-dimension` - Use Mantine `Image` with `w` and `h` or `aspectRatio` to prevent CLS
- `font-loading` - Use `font-display: swap`; preload critical fonts
- `lazy-loading` - Use React.lazy + Suspense for route-level code splitting
- `skeleton-loading` - Use Mantine `Skeleton` component for loading placeholders
- `virtualize-lists` - For long lists (50+ items), use virtualization libraries

### 4. Style Selection (HIGH)

- `style-match` - Match style to golf social product type via `--design-system`
- `consistency` - Use Mantine theme object for consistent styling across all pages
- `no-emoji-icons` - Use `@tabler/icons-react` (Mantine's recommended icon library), not emojis
- `dark-mode-pairing` - Design light/dark variants together in Mantine theme
- `primary-action` - Each screen should have one primary `Button`; others use `variant="outline"` or `variant="subtle"`

### 5. Layout & Responsive (HIGH)

- `viewport-meta` - Vite template handles this; verify `width=device-width, initial-scale=1`
- `mobile-first` - Design mobile-first; use Mantine responsive props (`visibleFrom`, `hiddenFrom`)
- `breakpoint-consistency` - Use Mantine's theme breakpoints consistently
- `readable-font-size` - Minimum 16px body text on mobile (Mantine default `fz="md"`)
- `horizontal-scroll` - No horizontal scroll; use Mantine `Container` with appropriate `size`
- `spacing-scale` - Use Mantine's `theme.spacing` (xs/sm/md/lg/xl) consistently
- `container-width` - Use Mantine `Container` with `size` prop for consistent max-width

### 6. Typography & Color (MEDIUM)

- `line-height` - Use Mantine `Text` `lh` prop; default 1.55 is appropriate
- `font-pairing` - Search `--domain typography` and apply via `theme.fontFamily` / `theme.headings.fontFamily`
- `font-scale` - Use Mantine `fz` prop values (xs/sm/md/lg/xl) for consistent scale
- `color-semantic` - Define colors in Mantine theme.colors; reference as `color="primary"` not `color="#hex"`
- `color-dark-mode` - Use `primaryShade: { light: 6, dark: 8 }` for mode-specific shade selection

### 7. Animation (MEDIUM)

- `duration-timing` - 150-300ms for micro-interactions; use Mantine `Transition` component
- `transform-performance` - Use `transform`/`opacity` only; Mantine transitions handle this
- `loading-states` - Show `Skeleton` or `LoadingOverlay` when loading exceeds 300ms
- `easing` - Use Mantine transition presets (`fade`, `slide-up`, `scale`, `pop`)
- `reduced-motion` - Mantine respects `prefers-reduced-motion` in its `Transition` component

### 8. Forms & Feedback (MEDIUM)

- `input-labels` - Always use `label` prop on Mantine form components
- `error-placement` - Use `error` prop on Mantine inputs (displays below field automatically)
- `submit-feedback` - Use `Button` `loading` prop, then `Notification` for success/error
- `required-indicators` - Use `withAsterisk` prop on required Mantine form inputs
- `empty-states` - Use Mantine `Center` + `Stack` with helpful message and action button
- `toast-dismiss` - Use `@mantine/notifications` with `autoClose` (3000-5000ms)
- `confirmation-dialogs` - Use `modals.openConfirmModal()` from `@mantine/modals`
- `progressive-disclosure` - Use Mantine `Accordion`, `Collapse`, or `Stepper` for complex flows

### 9. Navigation Patterns (HIGH)

- `bottom-nav-limit` - Max 5 items in bottom navigation; use icons + labels
- `back-behavior` - Use react-router-dom `useNavigate(-1)` for predictable back behavior
- `deep-linking` - All key screens reachable via URL path (react-router-dom routes)
- `nav-state-active` - Use Mantine `NavLink` with `active` prop for current page indication
- `modal-escape` - Mantine `Modal`/`Drawer` handle close on Escape and overlay click by default
- `search-accessible` - Use Mantine `Spotlight` or `TextInput` with search icon for discovery

### 10. Charts & Data (LOW)

- `chart-type` - Match chart to data type; search `--domain chart` for recommendations
- `color-guidance` - Use accessible color palettes from Mantine theme for chart colors
- `legend-visible` - Always show legends near charts
- `tooltip-on-interact` - Provide tooltips showing exact values on hover/tap
- `responsive-chart` - Charts must reflow on small screens

## Pre-Delivery Checklist (Mantine v9)

### Visual Quality
- [ ] No emojis used as icons (use `@tabler/icons-react`)
- [ ] All icons come from a consistent icon family (Tabler Icons)
- [ ] Semantic theme tokens used consistently (no ad-hoc hardcoded colors)
- [ ] Mantine `Button` variants used correctly (filled for primary, outline/subtle for secondary)

### Interaction
- [ ] All tappable elements provide feedback (Mantine handles this for its components)
- [ ] Touch targets >= 44x44px (use `size="md"` or larger on Mantine components)
- [ ] Loading states use Mantine `loading` prop or `Skeleton` / `LoadingOverlay`
- [ ] Disabled states use Mantine `disabled` prop (consistent styling)
- [ ] `ActionIcon` components have `aria-label` prop

### Light/Dark Mode
- [ ] Primary text contrast >= 4.5:1 in both modes
- [ ] Both themes tested (toggle via `useMantineColorScheme()`)
- [ ] Custom CSS uses `light-dark()` function or `[data-mantine-color-scheme]` selectors
- [ ] Modal/Drawer overlays readable in both modes

### Layout
- [ ] Mobile-first responsive design verified
- [ ] Mantine `Container` used for consistent content width
- [ ] `AppShell` used for page-level layout structure
- [ ] Verified on 375px, 768px, 1024px, 1440px
- [ ] Mantine spacing scale (xs/sm/md/lg/xl) used consistently

### Accessibility
- [ ] All meaningful images have alt text
- [ ] Form fields use Mantine `label`, `description`, and `error` props
- [ ] Color is not the only indicator (icons/text supplement color meaning)
- [ ] `prefers-reduced-motion` respected
- [ ] Heading hierarchy uses Mantine `Title` with sequential `order` values

## Search Reference

### Available Domains

| Domain | Use For | Example Keywords |
|--------|---------|------------------|
| `product` | Product type recommendations | social, community, golf, sports, booking, scheduling |
| `style` | UI styles, colors, effects | glassmorphism, minimalism, dark mode, vibrant |
| `typography` | Font pairings, Google Fonts | modern, sporty, elegant, professional |
| `color` | Color palettes by product type | social, sports, community, green, nature |
| `landing` | Page structure, CTA strategies | hero, testimonial, pricing, social-proof |
| `chart` | Chart types, library recommendations | trend, comparison, timeline, funnel |
| `ux` | Best practices, anti-patterns | animation, accessibility, z-index, loading |
| `google-fonts` | Individual Google Fonts lookup | sans serif, modern, variable font |
| `react` | React performance | suspense, memo, rerender, bundle, dynamic import |
| `web` | Web accessibility | aria, focus, semantic, form, input type |

### Available Stacks

| Stack | When to Use |
|-------|-------------|
| `react` | React-specific performance and patterns (translate to Mantine v9) |
| `nextjs` | If migrating to Next.js in future |

## Tips for Better Results

### Query Strategy

- Combine product + industry + tone: `"golf social community vibrant mobile-first"` not just `"app"`
- Use `--design-system` first, then `--domain` to deep-dive specific areas
- Always translate results into Mantine v9 component patterns

### Common FindFore Queries

```bash
# Full design system for golf social app
python3 .claude/skills/ui-ux-pro-max/scripts/search.py "golf social community mobile-first premium" --design-system -p "FindFore"

# Color palettes for sports/social
python3 .claude/skills/ui-ux-pro-max/scripts/search.py "social sports green community" --domain color

# Font pairings for modern sporty feel
python3 .claude/skills/ui-ux-pro-max/scripts/search.py "modern sporty clean" --domain typography

# UX for mobile social app
python3 .claude/skills/ui-ux-pro-max/scripts/search.py "mobile social feed scroll touch" --domain ux

# React performance for feeds/lists
python3 .claude/skills/ui-ux-pro-max/scripts/search.py "list virtualization infinite scroll" --stack react
```
