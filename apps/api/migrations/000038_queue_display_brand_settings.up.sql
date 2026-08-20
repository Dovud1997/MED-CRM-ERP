UPDATE queue_displays
SET settings = jsonb_build_object(
  'slogan', 'SIZNING SOG''LIG''INGIZ — BIZNING G''AMXO''RLIGIMIZ',
  'heroTitle', 'Современная клиника',
  'heroSubtitle', 'Лучшие условия для вас',
  'phone', '',
  'instagram', '',
  'telegram', '',
  'qrMediaId', '',
  'showSlogan', true,
  'showPhone', true,
  'showInstagram', true,
  'showTelegram', true,
  'showQr', true
) || settings || jsonb_build_object(
  'secondaryColor', '#8f0735',
  'backgroundColor', '#fff7fa'
),
updated_at = now();
