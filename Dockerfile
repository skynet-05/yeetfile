# syntax=docker/dockerfile:1

FROM golang:1.20 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY web/ ./web
COPY shared/ ./shared

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -tags yeetfile-web -o /yeetfile-web ./web

# Server image
FROM alpine:latest
COPY --from=builder /yeetfile-web /
EXPOSE 8090

CMD ["/yeetfile-web"]
