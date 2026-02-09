-- enable pgcrypto for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE
  IF NOT EXISTS diary_analyses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    diary_id UUID NOT NULL,
    user_id UUID NOT NULL,
    family_id UUID NOT NULL,
    char_count INTEGER NOT NULL DEFAULT 0,
    sentence_count INTEGER NOT NULL DEFAULT 0,
    accuracy_score INTEGER NOT NULL DEFAULT 0,
    writing_time_seconds INTEGER NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
  );

CREATE INDEX IF NOT EXISTS idx_diary_analyses_user_id_created_at ON diary_analyses (user_id, created_at);