FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/hitalent ./cmd/hitalent-app

FROM golang:1.26-alpine AS migrate

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.27.1

WORKDIR /app

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /bin/hitalent /app/hitalent

EXPOSE 8080

ENTRYPOINT ["/app/hitalent"]
