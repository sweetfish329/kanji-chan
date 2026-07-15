# 1. フロントエンドのビルドステージ
FROM oven/bun:alpine AS frontend-builder
WORKDIR /app
COPY frontend/package.json frontend/bun.lock ./
RUN bun install
COPY frontend/ .
RUN bun run build

# 2. バックエンドのビルドステージ
FROM golang:alpine AS backend-builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# ソースコードをコピー
COPY backend/ .

# フロントエンドのビルド結果を Go の embed 対象ディレクトリにコピー
COPY --from=frontend-builder /app/build ./cmd/server/dist

# Go バイナリをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# 3. 実行ステージ
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
# SQLiteのDBファイル保存用ディレクトリを作成
RUN mkdir -p /data
ENV DB_PATH=/data/kanji.db

WORKDIR /
COPY --from=backend-builder /server /server

EXPOSE 8080
CMD ["/server"]
