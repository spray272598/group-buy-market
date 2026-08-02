# multi-stage build
FROM golang:1.24-alpine AS builder
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
EXPOSE 8091
ENTRYPOINT ["/app/server"]
CMD ["-config", "/app/configs/config-docker.yaml"]
