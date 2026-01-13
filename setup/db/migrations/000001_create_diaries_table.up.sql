CREATE TABLE
  diaries (
    id UUID NOT NULL PRIMARY KEY,
    user_id UUID NOT NULL,
    family_id UUID NOT NULL,
    title VARCHAR(255) NULL,
    content TEXT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
  );

CREATE INDEX idx_family_id_created_at ON diaries (family_id, created_at);