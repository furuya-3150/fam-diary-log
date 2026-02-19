-- Diary Analysis Seeder
-- このファイルは開発環境用の日記分析サンプルデータを投入します
-- 注意: このSQLは02_diary_seed.sqlの後に実行する必要があります
-- 注意: diary_analysesとdiariesは別DBのため、diary_idは02_diary_seed.sqlで定義された固定IDを使用します
-- 山田太郎の日記分析（過去2週間分）
INSERT INTO
  diary_analyses (
    id,
    diary_id,
    user_id,
    family_id,
    char_count,
    sentence_count,
    accuracy_score,
    writing_time_seconds,
    created_at,
    updated_at
  )
VALUES
  (
    gen_random_uuid (),
    'd1111111-0001-0000-0000-000000000001',
    'd48fa8d0-96a7-4a8c-82f4-2465f1aabbd8',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    46,
    3,
    85,
    180,
    NOW (),
    NOW ()
  ),
  (
    gen_random_uuid (),
    'd1111111-0002-0000-0000-000000000002',
    'd48fa8d0-96a7-4a8c-82f4-2465f1aabbd8',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    42,
    2,
    78,
    120,
    NOW () - INTERVAL '1 days',
    NOW () - INTERVAL '1 days'
  ),
  (
    gen_random_uuid (),
    'd1111111-0003-0000-0000-000000000003',
    'd48fa8d0-96a7-4a8c-82f4-2465f1aabbd8',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    35,
    2,
    82,
    90,
    NOW () - INTERVAL '2 days',
    NOW () - INTERVAL '2 days'
  ),
  (
    gen_random_uuid (),
    'd1111111-0004-0000-0000-000000000004',
    'd48fa8d0-96a7-4a8c-82f4-2465f1aabbd8',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    38,
    2,
    88,
    150,
    NOW () - INTERVAL '3 days',
    NOW () - INTERVAL '3 days'
  ),
  (
    gen_random_uuid (),
    'd1111111-0005-0000-0000-000000000005',
    'd48fa8d0-96a7-4a8c-82f4-2465f1aabbd8',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    44,
    2,
    75,
    200,
    NOW () - INTERVAL '4 days',
    NOW () - INTERVAL '4 days'
  ),
  (
    gen_random_uuid (),
    'd1111111-0006-0000-0000-000000000006',
    'd48fa8d0-96a7-4a8c-82f4-2465f1aabbd8',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    36,
    2,
    92,
    100,
    NOW () - INTERVAL '5 days',
    NOW () - INTERVAL '5 days'
  ),
  (
    gen_random_uuid (),
    'd1111111-0007-0000-0000-000000000007',
    'd48fa8d0-96a7-4a8c-82f4-2465f1aabbd8',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    40,
    2,
    86,
    140,
    NOW () - INTERVAL '6 days',
    NOW () - INTERVAL '6 days'
  ),
  -- 山田花子の日記分析
  (
    gen_random_uuid (),
    'd2222222-0001-0000-0000-000000000001',
    '22222222-2222-2222-2222-222222222222',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    38,
    3,
    90,
    160,
    NOW () - INTERVAL '7 days',
    NOW () - INTERVAL '7 days'
  ),
  (
    gen_random_uuid (),
    'd2222222-0002-0000-0000-000000000002',
    '22222222-2222-2222-2222-222222222222',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    42,
    2,
    88,
    240,
    NOW () - INTERVAL '2 days',
    NOW () - INTERVAL '2 days'
  ),
  (
    gen_random_uuid (),
    'd2222222-0003-0000-0000-000000000003',
    '22222222-2222-2222-2222-222222222222',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    34,
    2,
    85,
    110,
    NOW () - INTERVAL '4 days',
    NOW () - INTERVAL '4 days'
  ),
  (
    gen_random_uuid (),
    'd2222222-0004-0000-0000-000000000004',
    '22222222-2222-2222-2222-222222222222',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    36,
    2,
    92,
    180,
    NOW () - INTERVAL '6 days',
    NOW () - INTERVAL '6 days'
  ),
  (
    gen_random_uuid (),
    'd2222222-0005-0000-0000-000000000005',
    '22222222-2222-2222-2222-222222222222',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    45,
    2,
    87,
    130,
    NOW () - INTERVAL '8 days',
    NOW () - INTERVAL '8 days'
  ),
  (
    gen_random_uuid (),
    'd2222222-0006-0000-0000-000000000006',
    '22222222-2222-2222-2222-222222222222',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    30,
    2,
    83,
    95,
    NOW () - INTERVAL '12 days',
    NOW () - INTERVAL '12 days'
  ),
  -- Charlie Brownの日記分析
  (
    gen_random_uuid (),
    'd5555555-0001-0000-0000-000000000001',
    '55555555-5555-5555-5555-555555555555',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    62,
    2,
    70,
    150,
    NOW () - INTERVAL '1 day',
    NOW () - INTERVAL '1 day'
  ),
  (
    gen_random_uuid (),
    'd5555555-0002-0000-0000-000000000002',
    '55555555-5555-5555-5555-555555555555',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    66,
    2,
    75,
    120,
    NOW () - INTERVAL '3 days',
    NOW () - INTERVAL '3 days'
  ),
  (
    gen_random_uuid (),
    'd5555555-0003-0000-0000-000000000003',
    '55555555-5555-5555-5555-555555555555',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    54,
    2,
    68,
    100,
    NOW () - INTERVAL '7 days',
    NOW () - INTERVAL '7 days'
  ),
  -- Alice Smithの日記分析
  (
    gen_random_uuid (),
    'd3333333-0001-0000-0000-000000000001',
    '33333333-3333-3333-3333-333333333333',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    72,
    2,
    95,
    200,
    NOW () - INTERVAL '1 day',
    NOW () - INTERVAL '1 day'
  ),
  (
    gen_random_uuid (),
    'd3333333-0002-0000-0000-000000000002',
    '33333333-3333-3333-3333-333333333333',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    67,
    2,
    88,
    110,
    NOW () - INTERVAL '2 days',
    NOW () - INTERVAL '2 days'
  ),
  (
    gen_random_uuid (),
    'd3333333-0003-0000-0000-000000000003',
    '33333333-3333-3333-3333-333333333333',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    58,
    2,
    92,
    180,
    NOW () - INTERVAL '4 days',
    NOW () - INTERVAL '4 days'
  ),
  (
    gen_random_uuid (),
    'd3333333-0004-0000-0000-000000000004',
    '33333333-3333-3333-3333-333333333333',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    71,
    2,
    90,
    140,
    NOW () - INTERVAL '6 days',
    NOW () - INTERVAL '6 days'
  ),
  (
    gen_random_uuid (),
    'd3333333-0005-0000-0000-000000000005',
    '33333333-3333-3333-3333-333333333333',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    74,
    2,
    87,
    90,
    NOW () - INTERVAL '9 days',
    NOW () - INTERVAL '9 days'
  ),
  -- Bob Johnsonの日記分析
  (
    gen_random_uuid (),
    'd4444444-0001-0000-0000-000000000001',
    '44444444-4444-4444-4444-444444444444',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    63,
    2,
    80,
    130,
    NOW () - INTERVAL '1 day',
    NOW () - INTERVAL '1 day'
  ),
  (
    gen_random_uuid (),
    'd4444444-0002-0000-0000-000000000002',
    '44444444-4444-4444-4444-444444444444',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    70,
    2,
    85,
    220,
    NOW () - INTERVAL '3 days',
    NOW () - INTERVAL '3 days'
  ),
  (
    gen_random_uuid (),
    'd4444444-0003-0000-0000-000000000003',
    '44444444-4444-4444-4444-444444444444',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    69,
    2,
    78,
    170,
    NOW () - INTERVAL '5 days',
    NOW () - INTERVAL '5 days'
  ),
  (
    gen_random_uuid (),
    'd4444444-0004-0000-0000-000000000004',
    '44444444-4444-4444-4444-444444444444',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    56,
    2,
    82,
    100,
    NOW () - INTERVAL '8 days',
    NOW () - INTERVAL '8 days'
  ),
  (
    gen_random_uuid (),
    'd4444444-0005-0000-0000-000000000005',
    '44444444-4444-4444-4444-444444444444',
    'f6113ded-5185-42e8-89fa-1a3f36414e66',
    60,
    2,
    88,
    150,
    NOW () - INTERVAL '11 days',
    NOW () - INTERVAL '11 days'
  ) ON CONFLICT DO NOTHING;