---
name: DocAI
colors:
  surface: "#f7f9fb"
  surface-dim: "#d8dadc"
  surface-bright: "#f7f9fb"
  surface-container-lowest: "#ffffff"
  surface-container-low: "#f2f4f6"
  surface-container: "#eceef0"
  surface-container-high: "#e6e8ea"
  surface-container-highest: "#e0e3e5"
  on-surface: "#191c1e"
  on-surface-variant: "#454652"
  inverse-surface: "#2d3133"
  inverse-on-surface: "#eff1f3"
  outline: "#757684"
  outline-variant: "#c5c5d4"
  surface-tint: "#4355b9"
  primary: "#24389c"
  on-primary: "#ffffff"
  primary-container: "#3f51b5"
  on-primary-container: "#cacfff"
  inverse-primary: "#bac3ff"
  secondary: "#4c56af"
  on-secondary: "#ffffff"
  secondary-container: "#959efd"
  on-secondary-container: "#27308a"
  tertiary: "#313e7e"
  on-tertiary: "#ffffff"
  tertiary-container: "#495697"
  on-tertiary-container: "#c9d0ff"
  error: "#ba1a1a"
  on-error: "#ffffff"
  error-container: "#ffdad6"
  on-error-container: "#93000a"
  primary-fixed: "#dee0ff"
  primary-fixed-dim: "#bac3ff"
  on-primary-fixed: "#00105c"
  on-primary-fixed-variant: "#293ca0"
  secondary-fixed: "#e0e0ff"
  secondary-fixed-dim: "#bdc2ff"
  on-secondary-fixed: "#000767"
  on-secondary-fixed-variant: "#343d96"
  tertiary-fixed: "#dee1ff"
  tertiary-fixed-dim: "#b9c3ff"
  on-tertiary-fixed: "#021355"
  on-tertiary-fixed-variant: "#354282"
  background: "#f7f9fb"
  on-background: "#191c1e"
  surface-variant: "#e0e3e5"
  slate-900: "#0F172A"
  slate-700: "#334155"
  slate-400: "#94A3B8"
  status-pending: "#94A3B8"
  status-processing: "#3B82F6"
  status-done: "#10B981"
  status-error: "#EF4444"
typography:
  headline-lg:
    fontFamily: Geist
    fontSize: 30px
    fontWeight: "600"
    lineHeight: 38px
    letterSpacing: -0.02em
  headline-md:
    fontFamily: Geist
    fontSize: 24px
    fontWeight: "600"
    lineHeight: 32px
    letterSpacing: -0.01em
  headline-sm:
    fontFamily: Geist
    fontSize: 18px
    fontWeight: "600"
    lineHeight: 26px
  body-lg:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: "400"
    lineHeight: 24px
  body-md:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: "400"
    lineHeight: 20px
  label-md:
    fontFamily: Geist
    fontSize: 12px
    fontWeight: "500"
    lineHeight: 16px
    letterSpacing: 0.02em
  metadata-mono:
    fontFamily: Geist
    fontSize: 12px
    fontWeight: "400"
    lineHeight: 16px
  headline-lg-mobile:
    fontFamily: Geist
    fontSize: 24px
    fontWeight: "600"
    lineHeight: 32px
rounded:
  sm: 0.125rem
  DEFAULT: 0.25rem
  md: 0.375rem
  lg: 0.5rem
  xl: 0.75rem
  full: 9999px
spacing:
  baseline: 4px
  gutter: 16px
  margin-page: 24px
  panel-gap: 1px
  list-item-py: 12px
  list-item-px: 16px
---

## Brand & Style

The design system is engineered for a **Local-First AI** experience, emphasizing
privacy, high-speed utility, and technical precision. The brand personality is
**Intelligent, Secure, and Efficient**, catering to power users and
organizations who prioritize data sovereignty.

The design style is **Modern / Corporate** with a heavy influence from
**Minimalism** to manage high information density. It avoids unnecessary
flourishes, opting instead for a "glass-and-metal" aesthetic that feels like a
robust local tool rather than a transient web app.

Key principles include:

- **Density over Decoration:** Maximizing screen real estate for metadata and
  content.
- **Functional Clarity:** Using status-driven visual cues to communicate the AI
  lifecycle.
- **Institutional Security:** A structured, grid-heavy layout that evokes
  stability.

## Colors

The palette is rooted in **Deep Slates and Indigos** to establish a
professional, "IDE-like" environment.

- **Primary & Secondary:** A range of Indigos are used for primary actions and
  deep structural elements like sidebars.
- **Surface Strategy:** We use a neutral base of `#F8FAFC` (Slate 50) for
  backgrounds, with white containers to create subtle contrast.
- **Functional Palette:** Status colors are non-negotiable for the document
  lifecycle. `status-processing` uses a vibrant blue, often paired with a subtle
  pulse animation, while `status-done` uses a high-visibility emerald to signify
  successful AI extraction.
- **Metadata Text:** `slate-700` is the default for body text, while `slate-400`
  is reserved for secondary metadata and timestamps to maintain hierarchy.

## Typography

This design system utilizes a dual-sans-serif approach to balance readability
with a technical aesthetic.

- **Geist** is used for headlines, labels, and metadata. Its technical, slightly
  condensed nature allows for high-density information without sacrificing
  clarity. It feels "engineered," fitting for a local AI app.
- **Inter** is used for body copy and long-form extracted text. Its high
  x-height and neutrality make it ideal for reading summaries and document
  contents.
- **Monospace Nuance:** Geist naturally carries a monospaced feel in its
  numerals and labels, which should be leveraged for file sizes, timestamps, and
  entity tags.

## Layout & Spacing

The layout philosophy is a **Functional Grid** that prioritizes side-by-side
content views.

- **The Split-View:** The core Detail View is a 50/50 or 40/60 split between the
  "Extracted Text" (reading pane) and the "Metadata Panel" (data pane).
- **Density:** Spacing is tight (4px baseline) to allow for large document
  lists. List items use 12px vertical padding to ensure high visibility of
  status badges and titles in a compact space.
- **Breakpoints:**
  - **Desktop (1280px+):** Three-pane layout (Sidebar, List, Detail).
  - **Tablet (768px - 1279px):** Two-pane layout (List and Detail toggle).
  - **Mobile (<767px):** Single-pane stacked layout; side-by-side views reflow
    to a tabbed interface (Content vs. Metadata).

## Elevation & Depth

To maintain a "local utility" feel, the system uses **Tonal Layers** and
**Low-Contrast Outlines** instead of heavy shadows.

- **Surface Tiers:** The main background is the lowest tier. Functional panels
  (Sidebar, Metadata Panel) sit on top with a 1px border (`slate-200`) and no
  shadow.
- **Active States:** Elements being "processed" or "selected" may use a very
  subtle, tight ambient shadow (4px blur, 5% opacity) to provide a "lift"
  sensation without appearing floaty.
- **Glassmorphism (Limited):** Use backdrop blurs solely for the "File Upload
  Zone" overlay to keep the focus on the drop target while maintaining context
  of the current list.

## Shapes

The design system uses a **Soft (0.25rem)** roundedness for most UI components.
This provides a professional edge that feels precise but modern.

- **Standard Elements:** Buttons, input fields, and card containers use the base
  4px (0.25rem) radius.
- **Status Badges & Tags:** These use `rounded-lg` (8px) to distinguish them as
  discrete interactive or informational "pills" within the otherwise rectangular
  grid.
- **File Upload Zone:** Uses a dashed 1px border with a slightly larger radius
  (12px) to signify a receptive drop area.

## Components

### Buttons & Inputs

- **Primary Action:** Solid `indigo-700` background with white text. High
  contrast for "Upload" or "Process" actions.
- **Ghost Buttons:** Used for secondary actions (e.g., "Download Original") to
  keep the interface clean.
- **Input Fields:** Flat, white backgrounds with a 1px `slate-300` border. On
  focus, the border shifts to `primary_color`.

### High-Density Lists

- **Rows:** Each document row should display the icon, filename (bold), status
  badge, and file size.
- **Hover State:** A subtle background shift to `slate-100` to indicate
  interactivity.

### Status Badges

- **Anatomy:** A small 8px dot (the status color) + text label.
- **Processing State:** The processing badge should feature a "spin" or "pulse"
  icon to communicate active AI background tasks.

### Metadata Panels

- **Layout:** Vertical stacks of "Label" (Geist, uppercase, 10px) and "Value"
  (Inter, 14px).
- **Tags:** Small, rounded boxes for AI-generated entities with a light slate
  background to keep them unobtrusive.

### Content Views

- **Side-by-Side:** Extracted text is presented in a clean, high-readability
  column with adjustable font sizes. Metadata remains fixed on the right to
  allow for quick scanning of AI insights while reading.
