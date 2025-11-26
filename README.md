# QISUR

API de catálogo e identidad con REST + WebSockets.

## Requisitos
- Go 1.21+
- Docker y docker-compose (para base y migraciones)
- PostgreSQL (si corres sin Docker)

## Instalación y ejecución
1. Clonar el repo y configurar el entorno:
   ```bash
   cp .env.example .env  # ajusta credenciales si es necesario
   ```
2. Levantar base y aplicar migraciones:
   ```bash
   docker-compose up -d db
   psql "postgres://catalog:catalog@localhost:55432/catalog?sslmode=disable" -f migrations/001_init.sql
   psql "postgres://catalog:catalog@localhost:55432/catalog?sslmode=disable" -f migrations/002_seed_roles.sql
   psql "postgres://catalog:catalog@localhost:55432/catalog?sslmode=disable" -f migrations/003_update_products.sql
   psql "postgres://catalog:catalog@localhost:55432/catalog?sslmode=disable" -f migrations/004_verification_codes.sql
   psql "postgres://catalog:catalog@localhost:55432/catalog?sslmode=disable" -f migrations/005_seed_sample_data.sql
   ```
   (O bien con migrate: `docker-compose run --rm -e DATABASE_URL=postgres://catalog:catalog@db:5432/catalog?sslmode=disable migrate -path /migrations -database $env:DATABASE_URL up`)
3. Ejecutar la API:
   ```bash
   go run cmd/catalog-api/main.go
   ```
   Swagger disponible en `http://localhost:8080/docs` (spec en `/swagger/doc.json`).

## Configuración del entorno
Variables clave (ver `.env.example`):
- `DATABASE_URL`: URL de Postgres (por defecto `postgres://catalog:catalog@db:5432/catalog?sslmode=disable` en Docker).
- `HTTP_PORT`: Puerto HTTP (8080 por defecto).
- SMTP (`SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_FROM`) para envío de verificación.
- JWT (`JWT_SECRET`, `JWT_ISSUER`, `JWT_TTL`).

Plantilla `.env.example`:
```
HTTP_PORT=8080
POSTGRES_USER=catalog
POSTGRES_PASSWORD=catalog
POSTGRES_DB=catalog
POSTGRES_HOST=localhost
POSTGRES_PORT=55432
DATABASE_URL=postgres://catalog:catalog@localhost:55432/catalog?sslmode=disable

ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=changeme
ADMIN_FULL_NAME=Catalog Admin

JWT_SECRET=changeme
JWT_ISSUER=catalog-api
JWT_TTL=15m

SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=you@example.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=you@example.com
SMTP_TLS_SKIP_VERIFY=false
```

## WebSockets
- Servidor Socket.IO en `GET /socket.io/*any`.
- Eventos emitidos (namespace `/`):
  - `category.created|updated|deleted`
  - `product.created|updated|deleted`
  - `product.category_assigned`
- Catálogo de eventos: `GET /api/v1/events`.

Ejemplo de cliente (JavaScript):
```js
const socket = io("http://localhost:8080");
socket.on("product.updated", (data) => {
  console.log("Producto actualizado:", data);
});
```

## Ejemplos de uso (REST)
- Login:
  ```bash
  curl -X POST http://localhost:8080/api/v1/identity/login \
    -H "Content-Type: application/json" \
    -d '{"email":"user1@example.com","password":"password123"}'
  ```
- Crear categoría (requiere Bearer token con rol admin):
  ```bash
  curl -X POST http://localhost:8080/api/v1/categories \
    -H "Authorization: Bearer <TOKEN>" \
    -H "Content-Type: application/json" \
    -d '{"name":"Garden","description":"Outdoor and garden"}'
  ```
- Asignar categoría a producto:
  ```bash
  curl -X POST http://localhost:8080/api/v1/products/<product_id>/categories/<category_id> \
    -H "Authorization: Bearer <TOKEN>"
  ```
- Historial de producto (filtrado):
  ```bash
  curl "http://localhost:8080/api/v1/products/<product_id>/history?start=2025-01-01T00:00:00Z&end=2025-12-31T23:59:59Z"
  ```

