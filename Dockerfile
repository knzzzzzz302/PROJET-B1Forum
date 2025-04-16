# Étape de build
FROM golang:1.23.0 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app main.go

# Étape de prod
FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/app .
COPY ./certs ./certs
COPY ./public ./public

EXPOSE 3030

CMD ["./app", "--https"]
