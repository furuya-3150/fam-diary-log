-- User Context Seeder
-- このファイルは開発環境用のサンプルデータを投入します
-- ユーザーの作成
INSERT INTO
  users (
    id,
    email,
    name,
    provider,
    provider_id,
    created_at,
    updated_at
  )
VALUES
  (
    '11111111-1111-1111-1111-111111111111',
    'father@example.com',
    '山田太郎',
    'google',
    'google_111',
    NOW (),
    NOW ()
  ),
  (
    '22222222-2222-2222-2222-222222222222',
    'mother@example.com',
    '山田花子',
    'google',
    'google_222',
    NOW (),
    NOW ()
  ),
  (
    '33333333-3333-3333-3333-333333333333',
    'alice@example.com',
    'Alice Smith',
    'google',
    'google_333',
    NOW (),
    NOW ()
  ),
  (
    '44444444-4444-4444-4444-444444444444',
    'bob@example.com',
    'Bob Johnson',
    'google',
    'google_444',
    NOW (),
    NOW ()
  ),
  (
    '55555555-5555-5555-5555-555555555555',
    'charlie@example.com',
    'Charlie Brown',
    'google',
    'google_555',
    NOW (),
    NOW ()
  ) ON CONFLICT (email) DO NOTHING;

-- ファミリーメンバーの作成
-- 山田家: 太郎(管理者)、花子(メンバー)、Charlie(メンバー)
-- Smith Family: Alice(管理者)、Bob(メンバー)
INSERT INTO
  family_members (
    id,
    family_id,
    user_id,
    role,
    created_at,
    updated_at
  )
VALUES
  (
    gen_random_uuid (),
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    '11111111-1111-1111-1111-111111111111',
    1,
    NOW (),
    NOW ()
  ),
  (
    gen_random_uuid (),
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    '22222222-2222-2222-2222-222222222222',
    2,
    NOW (),
    NOW ()
  ),
  (
    gen_random_uuid (),
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    '55555555-5555-5555-5555-555555555555',
    2,
    NOW (),
    NOW ()
  ),
  (
    gen_random_uuid (),
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    '33333333-3333-3333-3333-333333333333',
    1,
    NOW (),
    NOW ()
  ),
  (
    gen_random_uuid (),
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    '44444444-4444-4444-4444-444444444444',
    2,
    NOW (),
    NOW ()
  ) ON CONFLICT (family_id, user_id) DO NOTHING;

-- ファミリー招待の作成（アクティブな招待）
INSERT INTO
  family_invitations (
    id,
    family_id,
    inviter_user_id,
    invitation_token,
    invited_emails,
    expires_at,
    created_at,
    updated_at
  )
VALUES
  (
    gen_random_uuid (),
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    '11111111-1111-1111-1111-111111111111',
    'test-invitation-token-001',
    '["invited1@example.com", "invited2@example.com"]',
    NOW () + INTERVAL '7 days',
    NOW (),
    NOW ()
  ),
  (
    gen_random_uuid (),
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    '33333333-3333-3333-3333-333333333333',
    'test-invitation-token-002',
    '["friend@example.com"]',
    NOW () + INTERVAL '7 days',
    NOW (),
    NOW ()
  ) ON CONFLICT (invitation_token) DO NOTHING;