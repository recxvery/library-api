FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o exe ./cmd/server

FROM alpine
WORKDIR /app

COPY --from=builder /app/exe .
COPY --from=builder /app/migrations ./migrations
ENTRYPOINT [ "./exe" ]