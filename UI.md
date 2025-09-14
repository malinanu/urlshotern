URL Shortener Documentation Dashboard

\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_

🎯 Goal

Create a responsive documentation dashboard for a URL Shortener.

It should allow users to browse structured documentation (Business Requirements, Product Requirements, Technical Specs) in an elegant, modern interface.

\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_

🖥️ Deliverables

•	Figma File with:

o	Desktop frame (1440 × 900)

o	Mobile frame (375 × 812)

o	Component library (sidebar items, header, content card, buttons)

o	Color \& typography tokens

o	Interaction prototypes (hover, focus, sidebar open/close)

•	Export Assets:

o	Color tokens (JSON or text sheet)

o	Typography scale

o	Icon set (SVGs)

o	Grid/spacing documentation

\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_

🧩 Layout Specification

Desktop (1440 × 900)

Section	Details

Sidebar	Fixed left, 240 px wide, contains Table of Contents with hierarchical links. Active item highlighted with a subtle blue background + left accent bar.

Header	Full-width top bar with Search field (left) and Export button (right). Height 64 px.

Main Content	Max width 1100 px. Includes page title, section headings, and markdown-like content blocks (cards, code snippets).

Optional Utility Panel	Right column (280 px) for quick stats or version info.

Mobile (375 × 812)

•	Hamburger menu in header opens slide-in sidebar.

•	Content stacked vertically with generous padding.

•	Sticky export button on header.

\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_

🎨 Visual Style

Element	Spec

Typography	Inter font family. 

H1: 32 px / Bold / 1.25 line-height 

H2: 24 px / Semi-bold 

Body: 16 px / Regular / 1.6 line-height

Color Palette	• Primary Blue #3B82F6 

• Background #F8F9FA 

• Text Dark #111827 

• Muted Text #6B7280 

• Border/Divider #E5E7EB

Spacing System	8-point grid. Major gutters 32 px. Sidebar padding 24 px.

Corners \& Elevation	Cards: 8 px rounded corners, soft shadow (0 4 8 rgba(0,0,0,0.08)).

Icons	Feather or Lucide style, 24 px line icons.

Interactions	Hover: subtle lift (-2 px, shadow intensity +10%). Sidebar slide animation 160 ms ease-in-out.

\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_

🔄 States \& Components

•	Sidebar Item: default / hover / active / focus.

•	Buttons: primary (blue), secondary (gray), disabled.

•	Search Field: default / focus with outline.

•	Cards: default, hover.

•	Header: default / scrolled (if sticky).

\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_

🧩 Components to Include in Library

•	SidebarItem

•	HeaderBar

•	SearchInput

•	ExportButton

•	ContentCard

•	MarkdownBlock

•	MobileDrawer

Each with Auto Layout and constraints for responsive resizing.

\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_

📱 Responsive Guidelines

•	Breakpoints: 320, 375, 768, 1024, 1440 px.

•	Sidebar collapses <1024 px.

•	Typography scales fluidly with viewport.

\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_

🔎 Accessibility

•	Minimum contrast ratio 4.5:1 for body text.

•	Keyboard focus ring visible on interactive elements.

•	Landmarks: header, nav, main labeled.

\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_

🗂 File Organization

•	Pages in Figma:

1\.	Moodboard \& References

2\.	Wireframes (Lo-fi)

3\.	High-Fidelity Desktop

4\.	High-Fidelity Mobile

5\.	Components \& Tokens

\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_

💡 Tips for the Designer

•	Start with grayscale wireframes to validate layout.

•	Use Figma’s Auto Layout for cards, sidebar items, and content blocks for easy scaling.

•	Create reusable color \& typography styles for developer handoff.

•	Prototype micro-interactions using Smart Animate.





