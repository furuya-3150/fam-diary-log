-- enable pgcrypto for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE
  IF NOT EXISTS family_invitations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    family_id uuid NOT NULL,
    inviter_user_id uuid NOT NULL,
    invitation_token varchar(255) NOT NULL UNIQUE,
    expires_at timestamp NOT NULL,
    created_at timestamp NOT NULL DEFAULT now (),
    updated_at timestamp NOT NULL DEFAULT now ()
  );

CREATE INDEX IF NOT EXISTS idx_family_invitations_family_id ON family_invitations (family_id);