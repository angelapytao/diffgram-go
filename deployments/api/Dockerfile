FROM golang:1.25-alpine AS builder
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

FROM alpine:3.19
RUN apk add --no-cache curl ca-certificates
COPY --from=builder /app/api /api
ENTRYPOINT ["/api"]
