# ビルド環境
# ------------------------------------------------
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

# モジュールダウンロード
RUN go mod download

# ルートディレクトリ以下を/appにコピー
COPY . .

# コンパイル
# 開発環境
RUN go build -o /app/cmd/diary-api/main /app/cmd/diary-api/main.go
# RUN go build -o migrate /app/db/migrate/migrate.go
# RUN go build -o seeder /app/db/seeders/seeder.go

# 本番環境
# ------------------------------------------------
FROM alpine:latest AS prod

WORKDIR /app

# ビルド環境でコンパイルされたバイナリファイルを/appにコピー
# COPY --from=builder /app/main ./
# COPY --from=builder /app/.env.prod ./.env
# COPY --from=builder /app/config/config.toml ./config/
# COPY --from=builder /app/migrate ./db/
# COPY --from=builder /app/seeder ./db/

EXPOSE 8080

# バイナリを実行
CMD ["/app/main"]


# 開発環境
# ------------------------------------------------
FROM golang:1.25-alpine AS dev

WORKDIR /app

COPY . .

RUN go install github.com/air-verse/air@latest && \
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest && \
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest && \
  go install gotest.tools/gotestsum@latest

RUN apk add --no-cache make git


CMD ["air", "-c", ".air-diary-api.toml"]