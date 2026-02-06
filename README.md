# Currency Exchange Rate Service

Go microservice that fetches ECB currency exchange rates from Bank.lv and exposes them via REST API.

## Quick Start (Docker)

```bash
docker compose up -d --build
```

This starts MariaDB, fetches initial rates for 10 currencies, and runs the API server on port 8080.

## API Endpoints

| Endpoint                               | Description                                                   |
|----------------------------------------|---------------------------------------------------------------|
| `GET /api/v1/rates/latest`             | Latest exchange rates for all currencies                      |
| `GET /api/v1/rates/history/{currency}` | Historical rates for a specific currency (e.g., `USD`, `GBP`) |

## CLI Commands

> [!Important]
> Requires MariaDB running

```bash
# Fetch rates (default: USD, GBP, JPY)
./currency-service fetch # or go run main.go fetch

# Fetch specific currencies
./currency-service fetch --currencies AUD,BRL,CAD,CHF,CNY,CZK,DKK,GBP,HKD,HUF # or go run main.go fetch --currencies AUD,BRL,CAD,CHF,CNY,CZK,DKK,GBP,HKD,HUF

# Start HTTP server
./currency-service serve --port 8080 # or go run main.go serve --port 8080
```
