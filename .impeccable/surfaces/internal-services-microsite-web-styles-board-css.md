---
version: 1
slug: "internal-services-microsite-web-styles-board-css"
primary_target: "internal/services/microsite_web/styles/board.css"
related_targets: []
---

# Surface brief: board

## Scope
Link-in-Bio template (theme.style = board), rendered at /m/<slug>. Owner content replaces the sample content; owner theme tokens (accent, card_fill, text, button_text) drive colour.

## Visitor mode
Persuade. Creators and personal brands; the visitor arrives from a social profile on a phone and taps the newest link.

## Constraints
Mobile-first, works without JavaScript, self-hosted fonts only with system fallbacks for non-Latin scripts, reduced motion respected, owner's brand leads (Trunc only in the removable badge).

## Direction contract
THESIS: links are departures, and the newest one is boarding now; it refuses the stack of identical pills.
OWN-WORLD: a station board in the owner's card_fill, defaulting to SBB-style deep blue rather than black; Martian Mono in white, uppercase, in fixed columns (number, destination, status chip NEW / LIVE / DEPARTED); hairline row rules; accent yellow for the NOW BOARDING row; no neon, no glow.
STORY: see who it is, then catch the newest thing first.
FIRST VIEWPORT: a header strip carries the name and avatar like a station name plate; the first row is the highlighted link, larger, with a one-time split-flap reveal (CSS steps, about 600ms, skipped under reduced motion); other rows are tight board lines; socials appear as a PLATFORMS row.
FORM: candidate 5 of 7 (assigned); seed:3d36f4265d193ccf
FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance
