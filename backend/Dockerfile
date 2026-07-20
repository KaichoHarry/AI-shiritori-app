# TODO: バックエンド実装ステップ(go.mod / cmd/server/main.go 作成)後に、
#       実際の依存関係・ビルドタグ・使用フレームワークに合わせて内容を見直す。

FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/bin/server ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/bin/server ./server
COPY migrations ./migrations
EXPOSE 8080
CMD ["./server"]
