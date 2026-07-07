# ビルドステージ
FROM golang:alpine AS builder

WORKDIR /app

# 依存ファイルのコピーとダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピーとビルド
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# 実行ステージ
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /

COPY --from=builder /server /server

EXPOSE 8080

CMD ["/server"]
