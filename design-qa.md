# Design QA — экран электронной очереди

- Source visual truth: `C:\Users\khadj\Downloads\ChatGPT Image 10 авг. 2026 г., 01_22_03 (2).png`
- Implementation screenshot: `D:\MED\queue-tv-implementation.png`
- Combined comparison: `D:\MED\queue-tv-comparison.png`
- Viewport: 1920 × 1080 CSS px, desktop/TV, device scale factor 1.
- Source pixels: 1680 × 941.
- Implementation pixels: 1920 × 1080.
- Normalization: implementation resized to 1680 × 945 for the combined comparison; aspect-ratio difference is below 0.5%.
- State: working public display, Uzbek interface, no active queue call, contact values intentionally empty.

## Full-view comparison evidence

The implemented screen preserves the reference composition: white rounded header, three-column operational area, prominent current-number card, central clinic media, recent-call list, optional contact strip and persistent announcement strip. The palette intentionally changes from violet to the existing ONA VA BOLA raspberry, burgundy and white brand colors.

## Focused-region comparison evidence

Header, current-number card, hero/media card, recent-call card and ticker were checked individually. Typography is readable at TV distance; alignment, rounded cards, large queue number hierarchy and media crop match the reference structure. The contact region is hidden when all enabled contact fields are empty, avoiding an unusable blank panel; it appears as soon as an administrator enters at least one contact or selects a QR asset.

## Required fidelity surfaces

- Fonts and typography: Arial/Segoe UI system stack, strong uppercase headings, large clock and queue number. Hierarchy and legibility pass.
- Spacing and layout rhythm: reference three-column ratio and rounded-card rhythm are preserved; empty states do not overflow. Pass.
- Colors and visual tokens: intentional brand adaptation to `#e80d4f`, `#8f0735`, white and pale pink. Contrast passes.
- Image quality and asset fidelity: generated 16:9 clinic-interior photo is sharp, correctly cropped and project-local; existing brand logo is reused.
- Copy and content: Russian, Uzbek and English display labels remain supported; admin-editable slogan and hero copy are reflected by the public API.

## Findings

- No actionable P0/P1/P2 visual mismatch remains.
- P3: optional weather widget from the reference is not included because it needs a reliable weather source and was outside the requested editable contact scope.

## Comparison history

1. Initial implementation showed an empty contact frame when no contact values existed and used the legacy dark secondary color.
2. Fixed by conditionally rendering the contact strip only when content exists and migrating the display palette to the ONA VA BOLA burgundy/pink tokens.
3. Post-fix evidence: `D:\MED\queue-tv-implementation.png` and `D:\MED\queue-tv-comparison.png`.

## Interactions tested

- Public display loads from the token URL.
- One-click screen/audio activation works.
- Admin opens “ТВ-экраны” → “Настроить экран”.
- Phone, Instagram, Telegram, QR selector and all visibility switches render.
- Saving settings returns “Настройки ТВ-экрана сохранены”.
- Docker migration 38 applied successfully.

## Console check

The production localhost page loaded and rendered correctly. A transient fetch error was observed only while the local verification server was deliberately restarted; it was not reproduced after the Docker deployment.

final result: passed
