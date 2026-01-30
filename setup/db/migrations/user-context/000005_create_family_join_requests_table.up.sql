CREATE TABLE
  IF NOT EXISTS family_join_requests (
    id uuid PRIMARY KEY,
    family_id uuid NOT NULL,
    user_id uuid NOT NULL,
    status int NOT NULL,
    responded_user_id uuid,
    responded_at timestamp,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL
  );

CREATE INDEX IF NOT EXISTS idx_family_join_requests_family_id ON family_join_requests (family_id);