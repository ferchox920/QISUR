FROM golang:1.25 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o catalog-api ./cmd/catalog-api

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/catalog-api /app/catalog-api
COPY docs/swagger /app/docs/swagger
COPY migrations /app/migrations

EXPOSE 8080
ENTRYPOINT ["/app/catalog-api"]
