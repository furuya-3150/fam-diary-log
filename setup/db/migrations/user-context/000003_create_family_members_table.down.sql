-- family_membersテーブル削除
-- インデックス削除
DROP INDEX IF EXISTS idx_family_members_family_id;

DROP INDEX IF EXISTS idx_family_members_user_id;

-- テーブル削除
DROP TABLE IF EXISTS family_members;