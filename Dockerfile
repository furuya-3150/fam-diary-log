# ビルド環境（全サービス共通）
# ------------------------------------------------
FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

# モジュールダウンロード
RUN go mod download

# ソースコードをコピー
COPY . .

# 全サービスをコンパイル
RUN go build -o /bin/diary-api      ./cmd/diary-api/main.go
RUN go build -o /bin/diary-analyzer ./cmd/diary-analyzer/main.go
RUN go build -o /bin/diary-analysis ./cmd/diary-analysis/main.go
RUN go build -o /bin/user-context   ./cmd/user-context/main.go
RUN go build -o /bin/diary-mailer   ./cmd/diary-mailer/main.go


# 本番環境: diary-api
# ------------------------------------------------
FROM alpine:latest AS prod-diary-api

RUN apk add --no-cache tzdata
ENV TZ=Asia/Tokyo
WORKDIR /app
COPY --from=builder /bin/diary-api ./diary-api
EXPOSE 8080
CMD ["./diary-api"]


# 本番環境: diary-analyzer
# ------------------------------------------------
FROM alpine:latest AS prod-diary-analyzer

RUN apk add --no-cache tzdata
ENV TZ=Asia/Tokyo
WORKDIR /app
COPY --from=builder /bin/diary-analyzer ./diary-analyzer
CMD ["./diary-analyzer"]


# 本番環境: diary-analysis
# ------------------------------------------------
FROM alpine:latest AS prod-diary-analysis

RUN apk add --no-cache tzdata
ENV TZ=Asia/Tokyo
WORKDIR /app
COPY --from=builder /bin/diary-analysis ./diary-analysis
EXPOSE 8081
CMD ["./diary-analysis"]


# 本番環境: user-context
# ------------------------------------------------
FROM alpine:latest AS prod-user-context

RUN apk add --no-cache tzdata
ENV TZ=Asia/Tokyo
WORKDIR /app
COPY --from=builder /bin/user-context ./user-context
EXPOSE 8082
CMD ["./user-context"]


# 本番環境: diary-mailer
# ------------------------------------------------
FROM alpine:latest AS prod-diary-mailer

RUN apk add --no-cache tzdata
ENV TZ=Asia/Tokyo
WORKDIR /app
COPY --from=builder /bin/diary-mailer ./diary-mailer
CMD ["./diary-mailer"]


# マイグレーション
# ------------------------------------------------
FROM golang:1.25-alpine AS prod-migrate

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

WORKDIR /app
COPY setup/db/migrations ./setup/db/migrations
COPY docker/migrate-entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

CMD ["./entrypoint.sh"]


# 開発環境（全サービス共通）
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