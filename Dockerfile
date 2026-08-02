# multi-stage build
# 与 go.mod 的 go 1.22 对齐；镜像可略新，但不要随意拉 latest
FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM alpine:3.19
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
COPY --from=builder /out/server /app/server
COPY configs /app/configs
COPY docs /app/docs
COPY web /app/web
WORKDIR /app
EXPOSE 8091
ENTRYPOINT ["/app/server"]
CMD ["-config", "/app/configs/config-docker.yaml"]
