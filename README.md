# 📦 QISUR Catalog API

API RESTful y WebSocket de alto rendimiento para la gestión de catálogos e identidad, construida con **Go (Golang)** siguiendo principios de **Arquitectura Limpia (Clean Architecture)**.

Este proyecto implementa un sistema robusto para manejar productos, categorías y usuarios, con notificaciones en tiempo real, seguridad avanzada y documentación automática.

---

## 🚀 Características Principales

### 🛒 Catálogo & Productos
- **CRUD Completo:** Gestión de Categorías y Productos.
- **Búsqueda Avanzada:** Filtrado por texto, paginación y ordenamiento dinámico.
- **Historial de Precios:** Auditoría automática de cambios en precio y stock (`ProductHistory`).
- **Relaciones:** Asignación de productos a múltiples categorías.
- **Tabla de relación:** `product_category` implementa la relación muchos-a-muchos entre productos y categorías.

### 🔐 Identidad & Seguridad
- **Autenticación JWT:** Tokens firmados para acceso seguro.
- **Roles y Permisos:** Sistema RBAC (Admin, User, Client).
- **Verificación de Email:** Flujo seguro de registro con códigos OTP (con soporte SMTP).
- **Rate Limiting:** Protección contra ataques DDoS y fuerza bruta (con limpieza de memoria).
- **Mitigación de Ataques:** Protección contra Timing Attacks en el login.
- **Security Headers:** Middleware para cabeceras defensivas HTTP.

### ⚡ Real-time (WebSockets)
- Notificaciones instantáneas para clientes conectados cuando ocurren cambios en el catálogo.
- Gestión eficiente de conexiones con canales y limpieza de recursos.
- **Eventos:** `product.created`, `product.updated`, `category.deleted`, etc.

### 🛠 Ingeniería & Infraestructura
- **Base de Datos:** PostgreSQL con `pgx/v5` y pool de conexiones optimizado.
- **Arquitectura:** Diseño hexagonal (Ports & Adapters) para desacoplar dominio de infraestructura.
- **Graceful Shutdown:** Manejo correcto de señales del sistema para apagado seguro.
- **Docker:** Contenerización completa para desarrollo y producción.

---

## 🛠️ Stack Tecnológico

- **Lenguaje:** Go 1.24
- **Framework Web:** [Gin Gonic](https://github.com/gin-gonic/gin)
- **Base de Datos:** PostgreSQL
- **Driver SQL:** [pgx/v5](https://github.com/jackc/pgx)
- **WebSockets:** [Gorilla WebSocket](https://github.com/gorilla/websocket)
- **Documentación:** [Swagger (Swaggo)](https://github.com/swaggo/swag)
- **Autenticación:** [Golang-JWT](https://github.com/golang-jwt/jwt)
- **Fechas:** Los parámetros `start`/`end` del historial usan formato RFC3339 (ej. `2023-01-01T00:00:00Z`).

---

## ⚙️ Configuración

El proyecto utiliza un archivo `.env` para la configuración. Copia el ejemplo para empezar:

```bash
cp .env.example .env
```

### Variables de Entorno

| Variable | Descripción | Valor por Defecto |
|---|---|---|
| `HTTP_PORT` | Puerto del servidor | `8080` |
| `DATABASE_URL` | String de conexión a Postgres | `postgres://...` |
| `POSTGRES_SSLMODE` | Modo SSL de Postgres | `disable` |
| `JWT_SECRET` | **Requerido**. Clave para firmar tokens | - |
| `JWT_ISSUER` | Emisor del token | `catalog-api` |
| `JWT_TTL` | Duración del token | `15m` |
| `WS_ALLOWED_ORIGINS`| Orígenes permitidos para WS (CORS) | `*` |
| `SMTP_HOST` | Host del servidor de correo | - |
| `SMTP_PORT` | Puerto SMTP | `587` |
| `SMTP_USERNAME` | Usuario SMTP | - |
| `SMTP_PASSWORD` | Password SMTP | - |
| `SMTP_FROM` | Remitente de correos | - |
| `SMTP_TLS_SKIP_VERIFY` | Saltar verificación TLS (solo dev) | `false` |
| `ADMIN_EMAIL` | Email para crear admin inicial | - |
| `ADMIN_PASSWORD` | Password del admin inicial | - |
| `ADMIN_FULL_NAME` | Nombre del admin inicial | `Catalog Admin` |
| `WS_ALLOWED_ORIGINS` | Lista de orígenes permitidos WS (coma) | `http://localhost:8080` |
| `SHUTDOWN_TIMEOUT` | Timeout de apagado elegante | `10s` |

---

## 🏃‍♂️ Cómo Ejecutar

### Opción A: Usando Docker (Recomendado)

Levanta la base de datos y la API automáticamente:

```bash
docker-compose up --build
```

### Opción B: Ejecución Local

1) **Levantar Base de Datos:**

```bash
docker run --name qisur-db -e POSTGRES_PASSWORD=catalog -e POSTGRES_DB=catalog -p 55432:5432 -d postgres:15-alpine
```

2) **Instalar Dependencias y Correr:**

```bash
go mod download
go run cmd/catalog-api/main.go
```

---

## 📖 Documentación de la API

- **Swagger UI:** `http://localhost:8080/docs/index.html`
- **Diagrama ER:** `http://localhost:8080/db-schema.puml`
- **Eventos WebSocket:** `ws://localhost:8080/ws?token=TU_JWT_TOKEN`

### Mensaje de ejemplo WS

```json
{
  "event": "product.updated",
  "data": {
    "id": "uuid...",
    "name": "Nuevo Nombre",
    "price": 1500
  }
}
```

---

## 📂 Estructura del Proyecto

```text
.
├── cmd/
│   └── catalog-api/    # Punto de entrada (Main)
├── internal/
│   ├── catalog/        # Dominio: productos/categorías
│   ├── identity/       # Dominio: usuarios y auth
│   ├── http/           # Transporte HTTP: handlers, middleware, router
│   ├── storage/        # Persistencia: repositorios Postgres
│   └── ws/             # Transporte WebSocket: Hub
├── pkg/
│   ├── config/         # Carga y validación de configuración
│   ├── crypto/         # JWT, hashing
│   ├── logger/         # Logs estructurados
│   └── mailer/         # Cliente SMTP
├── migrations/         # SQL de inicialización
└── docs/               # Swagger generado
```

---

## 🧪 Testing

```bash
go test ./... -v
```

---
