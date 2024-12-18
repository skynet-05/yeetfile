# syntax=docker/dockerfile:1

FROM alpine:latest AS builder

WORKDIR /app

RUN apk add --update go npm make
RUN npm install -g typescript

COPY go.mod go.sum ./
RUN go mod download

COPY backend/ ./backend
COPY utils/ ./utils
COPY web/ ./web
COPY shared/ ./shared
COPY tsconfig.json .

COPY Makefile .

RUN make backend

# Server image
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/yeetfile-server /app
RUN chmod +x /app/yeetfile-server
RUN apk add --update curl
EXPOSE 8090

CMD ["/app/yeetfile-server"]
