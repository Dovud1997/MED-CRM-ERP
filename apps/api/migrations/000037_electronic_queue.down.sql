DROP TABLE IF EXISTS queue_ticker_messages,queue_audio_templates,queue_display_media,queue_media_assets,queue_call_history,queue_entries,queue_display_rooms,queue_displays,queue_rooms CASCADE;
DELETE FROM permissions WHERE code IN ('queue:read','queue:manage','queue:call','queue:settings','queue:media','queue:audio','queue:display');
