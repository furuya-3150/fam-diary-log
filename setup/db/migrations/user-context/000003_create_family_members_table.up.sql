-- family_membersテーブル作成
CREATE TABLE
  family_members (
    id BIGSERIAL PRIMARY KEY,
    family_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role INT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE (family_id, user_id)
  );

CREATE INDEX idx_family_members_family_id ON family_members (family_id);

CREATE INDEX idx_family_members_user_id ON family_members (user_id);