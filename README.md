# QISUR

API de catalogo e identidad con REST + WebSockets.

## Requisitos
- Go 1.21+
- Docker y docker-compose (para base y migraciones)
- PostgreSQL (si corres sin Docker)

## Instalacion y ejecucion
1. Clonar el repo y configurar el entorno:
   ```bash
   cp .env.example .env
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
   Swagger UI en `http://localhost:8080/docs` (spec en `http://localhost:8080/swagger/doc.json`).
   UI de eventos en `http://localhost:8080/events-ui`.

## Configuracion del entorno
Variables clave (ver `.env.example`):
- `DATABASE_URL`: URL de Postgres (por defecto `postgres://catalog:catalog@db:5432/catalog?sslmode=disable` en Docker).
- `HTTP_PORT`: Puerto HTTP (8080 por defecto).
- `LOG_LEVEL`: Nivel de log (`debug|info|warn|error`, por defecto `info`).
- `LOG_FORMAT`: Formato de log (`json` por defecto, `text` para desarrollo).
- SMTP (`SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_FROM`) para envio de verificacion via gomail (TLS oportunista, con `SMTP_TLS_SKIP_VERIFY=true` solo para desarrollo).
- JWT (`JWT_SECRET`, `JWT_ISSUER`, `JWT_TTL`).

## Verificacion por email
- El envio de codigos usa `gopkg.in/gomail.v2` (SMTP). Si faltan `SMTP_HOST` o `SMTP_FROM`, el sender queda deshabilitado y se usa un noop.
- Para desarrollo sin TLS estricto, habilita `SMTP_TLS_SKIP_VERIFY=true`; en produccion dejar en `false` y usar credenciales reales.
- El contenido del correo es texto plano con el codigo de verificacion.

## WebSockets
- Endpoint WebSocket nativo en `GET /ws`.
- Mensajes JSON con forma `{"event": "<nombre>", "data": <payload>}`.
- Eventos emitidos:
  - `category.created|updated|deleted`
  - `product.created|updated|deleted`
  - `product.category_assigned`
- Catalogo de eventos: `GET /api/v1/events`.

Ejemplo de cliente (JavaScript):
```js
const ws = new WebSocket("ws://localhost:8080/ws");
ws.onmessage = (evt) => {
  const { event, data } = JSON.parse(evt.data);
  if (event === "product.updated") {
    console.log("Producto actualizado:", data);
  }
};
```

## Ejemplos de uso (REST)
Registro y autenticacion:
```bash
curl -X POST http://localhost:8080/api/v1/identity/users/client \
  -H "Content-Type: application/json" \
  -d '{"email":"client@example.com","password":"password123","full_name":"Cliente Uno"}'

curl -X POST http://localhost:8080/api/v1/identity/users \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123","full_name":"Usuario Uno"}'

curl -X POST http://localhost:8080/api/v1/identity/verify \
  -H "Content-Type: application/json" \
  -d '{"user_id":"<USER_ID>","code":"<CODE>"}'

curl -X POST http://localhost:8080/api/v1/identity/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user1@example.com","password":"password123"}'
```

Categorias:
```bash
curl http://localhost:8080/api/v1/categories

curl -X POST http://localhost:8080/api/v1/categories \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Garden","description":"Outdoor and garden"}'

curl -X PUT http://localhost:8080/api/v1/categories/<id> \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Garden 2","description":"Updated"}'

curl -X DELETE http://localhost:8080/api/v1/categories/<id> \
  -H "Authorization: Bearer <TOKEN>"
```

Productos:
```bash
curl "http://localhost:8080/api/v1/products?limit=20&offset=0"

curl -X POST http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Laptop","description":"Ligera","price":125000,"stock":30}'

curl -X PUT http://localhost:8080/api/v1/products/<id> \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Laptop","description":"Actualizada","price":129000,"stock":25}'

curl -X DELETE http://localhost:8080/api/v1/products/<id> \
  -H "Authorization: Bearer <TOKEN>"

curl -X POST http://localhost:8080/api/v1/products/<product_id>/categories/<category_id> \
  -H "Authorization: Bearer <TOKEN>"

curl "http://localhost:8080/api/v1/search?type=product&q=laptop&limit=5"

curl "http://localhost:8080/api/v1/products/<product_id>/history?start=2025-01-01T00:00:00Z&end=2025-12-31T23:59:59Z"
```

## Pruebas
- Suite de handlers HTTP (catalogo e identidad) con `gin` + `httptest` para validar codigos de estado, payloads y middleware de auth/roles.
- Repositorio de catalogo probado con `pgxmock` (paginas, historial de precios/stock, guardrails sin pool) y verificacion de conexion DSN.
- Ejecutar todo: `go test ./...`

## Sistema de relaciones (DB)
- `roles` 1:N `users` (cada usuario referencia un rol).
- `users` 1:1 `verification_codes` (PK compartida, se borra en cascada).
- `categories` N:M `products` via `product_category` (llaves foraneas con ON DELETE CASCADE).
- `products` 1:N `product_history` (historial de precio/stock por producto, cascada en borrado).
- Semillas: `roles` basicos, categorias y productos de ejemplo, usuarios de prueba (password `password123`).

Diagrama ER en PlantUML: `docs/db-schema.puml` (servido en `GET /db-schema.puml`).


