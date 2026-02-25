FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 go build -o /app/server ./cmd/api
RUN CGO_ENABLED=0 go build -o /app/seed ./cmd/seed
RUN CGO_ENABLED=0 go build -o /app/fix-access-codes ./cmd/fix-access-codes
RUN CGO_ENABLED=0 go build -o /app/reset-db ./cmd/reset-db

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/seed .
COPY --from=builder /app/fix-access-codes .
COPY --from=builder /app/reset-db .
COPY --from=builder /app/openapi.json .
EXPOSE 8080
CMD ["./server"]