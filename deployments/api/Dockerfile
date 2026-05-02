FROM golang:1.25-alpine AS builder
WORKDIR /app
ENV GOPROXY=https://proxy.golang.org,direct
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

# 非 root 用户运行，与 builder 使用相同 alpine 版本
FROM alpine:3.21
RUN apk add --no-cache curl ca-certificates && \
    addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser
COPY --from=builder /app/api /api
ENTRYPOINT ["/api"]
