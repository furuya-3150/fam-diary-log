# fam-diary-log-api

家族日記アプリのバックエンド API サーバー群。  
サービスベース + イベント駆動アーキテクチャで構成される Go モノレポ。

## コンポーネント構成

| サービス         | ポート | 役割                                       |
| ---------------- | ------ | ------------------------------------------ |
| `diary-api`      | 8080   | 日記の作成・取得・ストリーク管理           |
| `diary-analysis` | 8081   | 日記分析結果の参照 API                     |
| `user-context`   | 8082   | 認証・ユーザー・家族管理                   |
| `diary-analyzer` | -      | 日記分析の非同期ワーカー（イベント駆動）   |
| `diary-mailer`   | -      | メール送信の非同期ワーカー（イベント駆動） |

## 技術スタック

- **言語**: Go 1.25
- **Web フレームワーク**: Echo v4
- **ORM**: GORM
- **DB**: PostgreSQL 16
- **メッセージブローカー**: RabbitMQ 3.12
- **認証**: JWT（HS256）/ Google OAuth2
- **CI/CD**: GitHub Actions
- **コンテナ**: Docker / Docker Hub

## ローカル開発

### 前提条件

- Docker / Docker Compose
- Go 1.25+
- [golang-migrate](https://github.com/golang-migrate/migrate)（DBマイグレーション用）

### セットアップ

```bash
# 1. 環境変数ファイルを作成（各サービスの .env.example を参考に）
cp cmd/diary-api/.env.example cmd/diary-api/.env
cp cmd/diary-analysis/.env.example cmd/diary-analysis/.env
cp cmd/user-context/.env.example cmd/user-context/.env
cp cmd/diary-analyzer/.env.example cmd/diary-analyzer/.env
cp cmd/diary-mailer/.env.example cmd/diary-mailer/.env

# 2. DB 初期化スクリプトをセットアップ
cp docker/db/init/01_init.sql.example docker/db/init/01_init.sql

# 3. Makefile のセットアップ
cp Makefile.example Makefile
# Makefile 内の DATABASE_URL を .env に合わせて編集

# 4. コンテナ起動
docker compose up -d --build

# 5. マイグレーション実行
make migrate-up
```

### マイグレーション

```bash
# マイグレーション適用
make migrate-up

# 1つ戻す
make migrate-down

# バージョン強制指定
make migrate-force v=<version>
```

## CI / CD

### CI（`diary-backend.yml`）

全ブランチへの push で以下を実行：

- テスト（カバレッジレポート付き）
- ビルド確認
- `go mod tidy` の差分チェック
- 脆弱性スキャン（`govulncheck`）

### CD（`docker-publish.yml`）

`main` ブランチへの push で、変更があったコンポーネントのみ実行：

```
変更検出（paths-filter）
    └── Docker Hub へイメージ push
            └── VPS へ SSH してコンテナを再起動
```

### GitHub Secrets

| Secret名              | 説明                        |
| --------------------- | --------------------------- |
| `DOCKERHUB_USERNAME`  | Docker Hub ユーザー名       |
| `DOCKERHUB_TOKEN`     | Docker Hub アクセストークン |
| `PROD_SERVER_HOST`    | 本番サーバー IP / ドメイン  |
| `PROD_SERVER_USER`    | 本番サーバー SSH ユーザー名 |
| `PROD_SERVER_SSH_KEY` | 本番サーバー SSH 秘密鍵     |
| `SLACK_BOT_TOKEN`     | Slack 通知用 Bot トークン   |

## 本番デプロイ

`main` ブランチへの push をトリガーに GitHub Actions が自動で実行する。

```
変更検出（paths-filter）
    └── Docker Hub へイメージ push（変更コンポーネントのみ）
            └── VPS へ SSH してコンテナを再起動
```

手動でのデプロイが必要な場合（スキーマ変更時のマイグレーション等）：

```bash
# マイグレーション実行
docker compose -f docker-compose.prod.yml --profile migrate run --rm migrate
```
