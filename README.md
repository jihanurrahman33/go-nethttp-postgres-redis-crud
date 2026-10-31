# Go CRUD Backend (`net/http`, PostgreSQL, Redis)

A clean, production-ready Go CRUD backend service using standard library `net/http` (Go 1.22+ routing), PostgreSQL for persistent storage, and Redis for caching with the Cache-Aside pattern.

## Architecture

- **HTTP Routing**: Go standard library `net/http.ServeMux` with pattern matching (`POST /api/v1/products`, `GET /api/v1/products/{id}`, etc.)
- **Database**: PostgreSQL (`github.com/lib/pq`) with automated schema migrations on startup
- **Cache**: Redis (`github.com/redis/go-redis/v9`) implementing the Cache-Aside pattern
- **Clean Architecture**:
  - `cmd/api`: Application entry point and graceful shutdown
  - `internal/config`: Environment variable loading with fallback defaults
  - `internal/database`: Postgres connection pool & migration, Redis connection
  - `internal/models`: Domain entities and input validation
  - `internal/repository`: Database access layer (SQL queries)
  - `internal/service`: Business logic & cache-aside coordination
  - `internal/handler`: HTTP request/response handlers and middlewares (logging, CORS, panic recovery)

---

## Getting Started

### 1. Prerequisites
- Go 1.22+
- Docker & Docker Compose (or local Postgres & Redis servers)

### 2. Start PostgreSQL & Redis
```bash
docker compose up -d
```

### 3. Configure Environment (Optional)
Copy `.env.example` to `.env` or set environment variables:
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgrespassword
export DB_NAME=crud_db
export REDIS_ADDR=localhost:6379
```

### 4. Run the Backend
```bash
go run cmd/api/main.go
```
The server will start on `http://localhost:8080`.

---

## API Endpoints & Examples

### 1. Health Check
Checks connectivity for both PostgreSQL and Redis:
```bash
curl -i http://localhost:8080/health
```

### 2. Create Product (POST)
```bash
curl -i -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Mechanical Keyboard",
    "description": "Wireless RGB mechanical keyboard",
    "price": 89.99,
    "stock": 15
  }'
```

### 3. Get Product by ID (GET)
*First request fetches from Postgres and caches in Redis; subsequent requests read directly from Redis.*
```bash
curl -i http://localhost:8080/api/v1/products/1
```

### 4. List Products (GET)
Supports pagination via `limit` and `offset`:
```bash
curl -i "http://localhost:8080/api/v1/products?limit=10&offset=0"
```

### 5. Update Product (PUT)
*Updates Postgres record and invalidates the Redis cache key.*
```bash
curl -i -X PUT http://localhost:8080/api/v1/products/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Mechanical Keyboard Pro",
    "description": "Upgraded wireless RGB mechanical keyboard",
    "price": 109.99,
    "stock": 25
  }'
```

### 6. Delete Product (DELETE)
*Deletes Postgres record and evicts from Redis cache.*
```bash
curl -i -X DELETE http://localhost:8080/api/v1/products/1
```

---

## Testing

Run unit tests:
```bash
go test -v ./...
```
