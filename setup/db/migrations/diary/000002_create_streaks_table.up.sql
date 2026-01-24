CREATE TABLE
  streaks (
    user_id UUID NOT NULL,
    family_id UUID NOT NULL,
    current_streak INTEGER NOT NULL DEFAULT 0,
    last_post_date DATE NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, family_id)
  );

CREATE INDEX idx_family_id_user_id ON streaks (family_id, user_id);