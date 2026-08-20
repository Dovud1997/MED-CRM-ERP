UPDATE queue_displays
SET settings = settings
  - 'slogan'
  - 'heroTitle'
  - 'heroSubtitle'
  - 'phone'
  - 'instagram'
  - 'telegram'
  - 'qrMediaId'
  - 'showSlogan'
  - 'showPhone'
  - 'showInstagram'
  - 'showTelegram'
  - 'showQr',
updated_at = now();
