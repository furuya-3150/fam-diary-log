ALTER TABLE notification_settings DROP CONSTRAINT IF EXISTS fk_notification_family;
ALTER TABLE notification_settings DROP CONSTRAINT IF EXISTS fk_notification_user;
DROP TABLE IF EXISTS notification_settings;
