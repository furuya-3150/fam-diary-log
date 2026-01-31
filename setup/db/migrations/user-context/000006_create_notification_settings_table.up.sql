CREATE TABLE notification_settings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  family_id UUID NOT NULL,
  post_created_enabled BOOLEAN NOT NULL DEFAULT true,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  CONSTRAINT uq_notification_user_family UNIQUE (user_id, family_id)
);
CREATE INDEX idx_notification_settings_family_id ON notification_settings (family_id);