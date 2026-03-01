-- Diary Seeder
-- このファイルは開発環境用の日記サンプルデータを投入します

-- 過去30日分の日記を作成（山田家のメンバー用）
-- 山田太郎の日記（過去2週間）
INSERT INTO diaries (id, user_id, family_id, title, content, writing_time_seconds, created_at, updated_at) VALUES
  ('d1111111-0001-0000-0000-000000000001', 'd48fa8d0-96a7-4a8c-82f4-2465f1aabbd8', 'f6113ded-5185-42e8-89fa-1a3f36414e66', '今日は良い天気', '今日は晴れていて気持ちが良かった。家族で公園に散歩に行った。子供たちも楽しそうで嬉しかった。', 180, NOW(), NOW()),
  ('d1111111-0002-0000-0000-000000000002', 'd48fa8d0-96a7-4a8c-82f4-2465f1aabbd8', 'f6113ded-5185-42e8-89fa-1a3f36414e66', '仕事が忙しい', '今日は一日中会議だった。疲れたけど、プロジェクトが順調に進んでいるのでよかった。', 120, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
  ('d1111111-0003-0000-0000-000000000003', 'd48fa8d0-96a7-4a8c-82f4-2465f1aabbd8', 'f6113ded-5185-42e8-89fa-1a3f36414e66', '週末の予定', '週末は家族でキャンプに行く予定。準備が少し大変だけど楽しみ。', 90, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
  ('d1111111-0004-0000-0000-000000000004', 'd48fa8d0-96a7-4a8c-82f4-2465f1aabbd8', 'f6113ded-5185-42e8-89fa-1a3f36414e66', '新しいレシピに挑戦', '今日は夕食に新しい料理を作ってみた。家族に好評でよかった。', 150, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
  ('d1111111-0005-0000-0000-000000000005', 'd48fa8d0-96a7-4a8c-82f4-2465f1aabbd8', 'f6113ded-5185-42e8-89fa-1a3f36414e66', '運動不足解消', 'ジムに久しぶりに行った。体が重かったけど、これから定期的に行こうと思う。', 200, NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),
  ('d1111111-0006-0000-0000-000000000006', 'd48fa8d0-96a7-4a8c-82f4-2465f1aabbd8', 'f6113ded-5185-42e8-89fa-1a3f36414e66', '読書の時間', '久しぶりに小説を読んだ。ゆっくりとした時間が持てて良かった。', 100, NOW() - INTERVAL '10 days', NOW() - INTERVAL '10 days'),
  ('d1111111-0007-0000-0000-000000000007', 'd48fa8d0-96a7-4a8c-82f4-2465f1aabbd8', 'f6113ded-5185-42e8-89fa-1a3f36414e66', '友人との再会', '学生時代の友人と久しぶりに会った。昔話に花が咲いて楽しかった。', 140, NOW() - INTERVAL '14 days', NOW() - INTERVAL '14 days'),

-- 山田花子の日記（過去2週間）
  ('d2222222-0001-0000-0000-000000000001', '22222222-2222-2222-2222-222222222222', 'f6113ded-5185-42e8-89fa-1a3f36414e66', 'ガーデニング', '庭の花が綺麗に咲いた。水やりを毎日続けた成果が出て嬉しい。', 160, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
  ('d2222222-0002-0000-0000-000000000002', '22222222-2222-2222-2222-222222222222', 'f6113ded-5185-42e8-89fa-1a3f36414e66', 'お菓子作り', '今日はクッキーを焼いた。子供たちと一緒に作って楽しかった。', 240, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
  ('d2222222-0003-0000-0000-000000000003', '22222222-2222-2222-2222-222222222222', 'f6113ded-5185-42e8-89fa-1a3f36414e66', 'カフェ巡り', '新しいカフェを見つけた。雰囲気が良くて、また行きたい。', 110, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
  ('d2222222-0004-0000-0000-000000000004', '22222222-2222-2222-2222-222222222222', 'f6113ded-5185-42e8-89fa-1a3f36414e66', 'ヨガレッスン', 'ヨガ教室に参加した。体がほぐれてリフレッシュできた。', 180, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),
  ('d2222222-0005-0000-0000-000000000005', '22222222-2222-2222-2222-222222222222', 'f6113ded-5185-42e8-89fa-1a3f36414e66', '家族写真', '家族で写真館に行った。久しぶりの家族写真が撮れて良い思い出になった。', 130, NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),
  ('d2222222-0006-0000-0000-000000000006', '22222222-2222-2222-2222-222222222222', 'f6113ded-5185-42e8-89fa-1a3f36414e66', '映画鑑賞', '家で映画を見た。感動する内容でとても良かった。', 95, NOW() - INTERVAL '12 days', NOW() - INTERVAL '12 days'),

-- Charlie Brownの日記（山田家メンバー）
  ('d5555555-0001-0000-0000-000000000001', '55555555-5555-5555-5555-555555555555', 'f6113ded-5185-42e8-89fa-1a3f36414e66', 'Learning Japanese', 'Started learning more Japanese today. It''s challenging but fun!', 150, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
  ('d5555555-0002-0000-0000-000000000002', '55555555-5555-5555-5555-555555555555', 'f6113ded-5185-42e8-89fa-1a3f36414e66', 'Family Dinner', 'Had a great dinner with the Yamada family. The food was delicious!', 120, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
  ('d5555555-0003-0000-0000-000000000003', '55555555-5555-5555-5555-555555555555', 'f6113ded-5185-42e8-89fa-1a3f36414e66', 'Weekend Activities', 'Went hiking with the family. The weather was perfect.', 100, NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days'),

-- Alice Smithの日記（Smith Family）
  ('d3333333-0001-0000-0000-000000000001', '33333333-3333-3333-3333-333333333333', 'f6113ded-5185-42e8-89fa-1a3f36414e66', 'Project Milestone', 'Reached an important milestone at work today. Team celebration tomorrow!', 200, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
  ('d3333333-0002-0000-0000-000000000002', '33333333-3333-3333-3333-333333333333', 'f6113ded-5185-42e8-89fa-1a3f36414e66', 'Morning Routine', 'Started a new morning routine with meditation. Feeling more focused.', 110, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),
  ('d3333333-0003-0000-0000-000000000003', '33333333-3333-3333-3333-333333333333', 'f6113ded-5185-42e8-89fa-1a3f36414e66', 'Cooking Experiment', 'Tried a new recipe from the cookbook. It turned out great!', 180, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days'),
  ('d3333333-0004-0000-0000-000000000004', '33333333-3333-3333-3333-333333333333', 'f6113ded-5185-42e8-89fa-1a3f36414e66', 'Book Club Meeting', 'Had a wonderful discussion at book club. Interesting perspectives shared.', 140, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days'),
  ('d3333333-0005-0000-0000-000000000005', '33333333-3333-3333-3333-333333333333', 'f6113ded-5185-42e8-89fa-1a3f36414e66', 'Spring Cleaning', 'Spent the day organizing the house. Very satisfying to see everything neat.', 90, NOW() - INTERVAL '9 days', NOW() - INTERVAL '9 days'),

-- Bob Johnsonの日記（Smith Family）
  ('d4444444-0001-0000-0000-000000000001', '44444444-4444-4444-4444-444444444444', 'f6113ded-5185-42e8-89fa-1a3f36414e66', 'Basketball Game', 'Played basketball with friends today. Great workout and fun time!', 130, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
  ('d4444444-0002-0000-0000-000000000002', '44444444-4444-4444-4444-444444444444', 'f6113ded-5185-42e8-89fa-1a3f36414e66', 'Tech Conference', 'Attended an interesting tech conference. Learned about new frameworks.', 220, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
  ('d4444444-0003-0000-0000-000000000003', '44444444-4444-4444-4444-444444444444', 'f6113ded-5185-42e8-89fa-1a3f36414e66', 'Guitar Practice', 'Practiced guitar for an hour. Making progress on that difficult song.', 170, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
  ('d4444444-0004-0000-0000-000000000004', '44444444-4444-4444-4444-444444444444', 'f6113ded-5185-42e8-89fa-1a3f36414e66', 'Family Game Night', 'Had a fun game night with the family. Lots of laughter!', 100, NOW() - INTERVAL '8 days', NOW() - INTERVAL '8 days'),
  ('d4444444-0005-0000-0000-000000000005', '44444444-4444-4444-4444-444444444444', 'f6113ded-5185-42e8-89fa-1a3f36414e66', 'Productive Day', 'Very productive day at work. Completed all tasks on my list.', 150, NOW() - INTERVAL '11 days', NOW() - INTERVAL '11 days');

-- ストリークデータの作成
-- 山田家のストリーク
INSERT INTO streaks (family_id, user_id,current_streak, last_post_date, created_at, updated_at) VALUES
  ('f6113ded-5185-42e8-89fa-1a3f36414e66', 'd48fa8d0-96a7-4a8c-82f4-2465f1aabbd8', 3, (NOW() - INTERVAL '1 day')::date, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Smith Familyのストリーク
INSERT INTO streaks (family_id, user_id, current_streak, last_post_date, created_at, updated_at) VALUES
  ('f6113ded-5185-42e8-89fa-1a3f36414e66', '33333333-3333-3333-3333-333333333333', 2, (NOW() - INTERVAL '1 day')::date, NOW(), NOW())
ON CONFLICT DO NOTHING;
