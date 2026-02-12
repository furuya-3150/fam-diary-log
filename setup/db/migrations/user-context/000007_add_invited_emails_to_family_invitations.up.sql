-- Add invited_emails column to family_invitations table after invitation_token
ALTER TABLE family_invitations
ADD COLUMN invited_emails jsonb NOT NULL DEFAULT '[]'