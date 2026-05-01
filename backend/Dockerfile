FROM docker.io/library/golang:1.26.2-alpine AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/backend ./cmd/server

FROM docker.io/library/alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /out/backend /usr/local/bin/backend
COPY migrations /app/migrations

ENV BACKEND_HTTP_PORT=8082
ENV GIN_MODE=release

EXPOSE 8082

ENTRYPOINT ["backend"]
