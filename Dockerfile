FROM golang:1.25.7-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /currency-service .

FROM gcr.io/distroless/static-debian12:nonroot AS runner

WORKDIR /app

COPY --from=builder /currency-service .

EXPOSE 8080

ENTRYPOINT ["./currency-service"]
