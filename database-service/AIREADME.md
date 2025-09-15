# Autovibe Database Service

HTTP REST API service boilerplate using Go, Gin framework, and Ent ORM.

## Features

- ✅ **HTTP REST API** with Gin framework
- ✅ **Ent ORM** for type-safe database operations
- ✅ **Pure JSON responses**
- ✅ **Environment-based configuration**
- ✅ **Graceful shutdown handling**
- ✅ **PostgreSQL support**

## Quick Start

```bash
# Build
make build

# Run
make run

# Generate Ent code (after schema changes)
make generate

# Test
make test
```

## API Endpoints

Basic example endpoints (customize for your needs):
- `GET /health` - Health check
- `GET /api/v1/examples` - List examples
- `POST /api/v1/examples` - Create example
- `GET /api/v1/examples/{id}` - Get example

## Configuration

Environment variables:
- `SERVICE_NAME` (default: example-service)
- `SERVER_PORT` (default: 8080)
- `DB_HOST` (default: localhost)
- `DB_PASSWORD` (required for production)
- `ENVIRONMENT` (default: dev)
- `DEBUG` (set to "true" for debug logging)

## Database Schema

Basic example schema included. Customize the `ent/schema/` directory for your specific needs.

## Development

1. Modify schemas in `ent/schema/`
2. Run `make generate` to generate Ent code
3. Update handlers in `internal/server/` as needed
4. Test with `make test`