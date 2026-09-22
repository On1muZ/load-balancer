FROM golang:1.26.7-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/app ./cmd/lb
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/backend ./test/backend.go

FROM alpine:3.20

WORKDIR /app
COPY nodes.json .
COPY --from=builder /app/bin/app .
COPY --from=builder /app/bin/backend .
EXPOSE 9000

CMD ["./app"]
