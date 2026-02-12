-- Remove invited_emails column from family_invitations table
ALTER TABLE family_invitations
DROP COLUMN IF EXISTS invited_emails