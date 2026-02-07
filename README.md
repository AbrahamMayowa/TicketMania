# TicketMania 🎟️

A robust, production-ready event ticketing REST API built with Go. Handles ticket sales, event management, and user authentication with built-in concurrency control, panic recovery, and graceful shutdown.

## Features

- 🎫 **Event Management** - Create and browse events with multiple ticket types
- 🔐 **Authentication & Authorization** - Secure JWT-based user authentication
- 💳 **Ticket Purchasing** - Buy tickets with support for multiple ticket types per order
- 🛡️ **Concurrency Control** - Race condition prevention for ticket inventory
- 🔄 **Panic Recovery** - Automatic recovery from runtime panics with detailed logging
- 🚦 **Graceful Shutdown** - Clean server shutdown with connection draining
- 📊 **Database Migrations** - Version-controlled schema management
- ✅ **Input Validation** - Request validation and sanitization
- 🏗️ **Clean Architecture** - Separation of concerns with middleware pattern

## Tech Stack

- **Language**: Go 1.25+
- **Router**: julienschmidt/httprouter
- **Database**: PostgreSQL
- **Migrations**: golang-migrate/migrate
- **Authentication**: JWT tokens

## Project Structure

```
ticketmania/
├── cmd/
│   └── api/          # Application entry point and HTTP handlers
├── internal/
│   ├── data/         # Database models and queries
│   ├── jsonlog/      # Structured JSON logging
│   └── validator/    # Input validation logic
├── migrations/       # Database migration files
├── vendor/          # Vendored dependencies
├── bin/             # Compiled binaries
├── Makefile         # Build and development commands
└── .env             # Environment configuration
```

## Getting Started

### Prerequisites

- Go 1.25 or higher
- PostgreSQL 12+
- golang-migrate CLI tool

Install golang-migrate:
```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Windows
choco install golang-migrate
```

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/AbrahamMayowa/ticketmania.git
cd ticketmania
```

2. **Set up environment variables**

Create a `.env` file in the root directory:

```env
DATABASE_URL=postgres://username:password@localhost/ticketmania?sslmode=disable
PORT=4000
JWT_SECRET=your-secret-key-here
ENV=development
HASH_SECRET_KEY="secret here"

MAILTRAP_TOKEN = 'token here'
MAILTRAP_HOST = live.smtp.mailtrap.io
MAILTRAP_PORT = 587
MAILTRAP_USERNAME = api
MAILTRAP_PASSWORD = "password here"
MAILTRAP_SENDER = "sender email here"
```

3. **Create the database**
```bash
createdb ticketmania
```

4. **Install dependencies**
```bash
make vendor
```

5. **Run database migrations**
```bash
make db/migrations/up
```

6. **Start the server**
```bash
make run/api
```

The API will be available at `http://localhost:4000`

## API Endpoints

### Authentication

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/v1/register` | Register a new user | ❌ |
| POST | `/v1/login` | Login and receive JWT token | ❌ |

### Events

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/v1/create-event` | Create a new event | ✅ |
| GET | `/v1/events` | List all events | ❌ |
| GET | `/v1/events/:id` | Get event details with ticket types | ❌ |

### Tickets

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/v1/buy-ticket` | Purchase tickets | ❌ |

## Database Schema

### Core Tables

**users**
- Stores user credentials and authentication data
- Password hashing with bcrypt
- Email validation

**events**
- Event details (title, description, location, datetime)
- Status tracking (draft, published, cancelled)
- Foreign key to user (event creator)

**ticket_types**
- Multiple ticket tiers per event (VIP, Regular, etc.)
- Price, currency, and inventory tracking
- Sold quantity management

**tickets**
- Individual ticket records
- Links to event, ticket type, and optionally user
- Status tracking (available, sold, used, cancelled)
- Guest purchase support (buyer_email, buyer_phone)

### Entity Relationships

```
users (1) ──────── (N) events
events (1) ──────── (N) ticket_types
ticket_types (1) ── (N) tickets
users (1) ──────── (N) tickets [optional - for registered users]
```

## Request Examples

### Register User
```bash
curl -X POST http://localhost:4000/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

**Response:**
```json
{
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2026-02-07T10:30:00Z"
  }
}
```

### Login
```bash
curl -X POST http://localhost:4000/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

### Create Event
```bash
curl -X POST http://localhost:4000/v1/create-event \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Tech Conference 2026",
    "description": "Annual technology conference",
    "location": "Lagos, Nigeria",
    "date": "2026-06-15",
    "start_time": "2026-06-15T09:00:00Z",
    "end_time": "2026-06-15T18:00:00Z",
    "ticket_types": [
      {
        "name": "VIP",
        "price": 50000,
        "currency": "NGN",
        "total_qty": 100
      },
      {
        "name": "Regular",
        "price": 25000,
        "currency": "NGN",
        "total_qty": 500
      }
    ]
  }'
```

**Response:**
```json
{
  "event": {
    "id": 1,
    "title": "Tech Conference 2026",
    "description": "Annual technology conference",
    "location": "Lagos, Nigeria",
    "date": "2026-06-15",
    "start_time": "2026-06-15T09:00:00Z",
    "end_time": "2026-06-15T18:00:00Z",
    "status": "published",
    "ticket_types": [
      {
        "id": 1,
        "name": "VIP",
        "price": 50000,
        "currency": "NGN",
        "total_qty": 100,
        "sold_qty": 0
      },
      {
        "id": 2,
        "name": "Regular",
        "price": 25000,
        "currency": "NGN",
        "total_qty": 500,
        "sold_qty": 0
      }
    ]
  }
}
```

### List Events
```bash
curl -X GET http://localhost:4000/v1/events
```

### Get Event Details
```bash
curl -X GET http://localhost:4000/v1/events/1
```

### Buy Tickets
```bash
curl -X POST http://localhost:4000/v1/buy-ticket \
  -H "Content-Type: application/json" \
  -d '[{
    "eventId": 1,
    "ticketTypes": [
      {
        "ticketTypeId": 1,
        "quantity": 2,
        "buyerEmail": "buyer@example.com",
        "buyerPhone": "+2348123456789"
      },
      {
        "ticketTypeId": 2,
        "quantity": 3,
        "buyerEmail": "buyer@example.com",
        "buyerPhone": "+2348123456789"
      }
    ]
  }]'
```

**Response:**
```json
{
  "tickets": [
    {
      "id": 1,
      "event_id": 1,
      "ticket_type_id": 1,
      "status": "sold",
      "buyer_email": "buyer@example.com",
      "buyer_phone": "+2348123456789"
    },
    {
      "id": 2,
      "event_id": 1,
      "ticket_type_id": 1,
      "status": "sold",
      "buyer_email": "buyer@example.com",
      "buyer_phone": "+2348123456789"
    }
  ],
  "total_amount": 100000,
  "currency": "NGN"
}
```

## Development

### Available Make Commands

```bash
# Development
make run/api              # Run the API server
make db/migrations/up     # Run all pending migrations
make db/migrations/new name=create_users_table  # Create new migration

# Quality Control
make audit               # Format, vet, and test code
make vendor             # Tidy and vendor dependencies

# Build
make dev/build/api      # Build for current OS
make prod/build/api     # Build for Linux AMD64 (production)
```

### Running Tests
```bash
make audit
```

This will:
1. Format all code with `go fmt`
2. Vet code for potential issues
3. Run all tests with race detection

### Creating Database Migrations
```bash
make db/migrations/new name=add_user_role
```

This creates two files in `migrations/`:
- `000XXX_add_user_role.up.sql` - Forward migration
- `000XXX_add_user_role.down.sql` - Rollback migration

### Building for Production
```bash
make prod/build/api
```

This creates a statically-linked binary in `bin/linux_amd64/api` optimized for production deployment.

## Architecture Highlights

### Middleware Chain

The application uses a layered middleware approach:

```
Request → recoverPanic → authenticate → router → handler
```

1. **recoverPanic**: Catches runtime panics, logs stack traces, returns 500 errors
2. **authenticate**: Extracts and validates JWT tokens, sets user context
3. **requireAuthentication**: Guards routes requiring authentication

Example middleware flow:
```go
return app.recoverPanic(app.authenticate(router))
```

### Concurrency Control

Ticket inventory is protected against race conditions:

- **Database transactions** for ticket purchases prevent overselling
- **Row-level locking** with `SELECT ... FOR UPDATE`
- **Atomic inventory updates** ensure consistency
- **Foreign key constraints** maintain referential integrity

Example transaction:
```go
tx.Begin()
// SELECT total_qty, sold_qty FROM ticket_types WHERE id = ? FOR UPDATE
// UPDATE ticket_types SET sold_qty = sold_qty + ? WHERE id = ?
// INSERT INTO tickets (...)
tx.Commit()
```

### Error Handling

Comprehensive error handling strategy:

- **Structured error responses** with consistent JSON format
- **Detailed error logging** with request context and stack traces
- **Foreign key validation** before database operations
- **Graceful handling** of `sql.ErrNoRows` and `sql.ErrNoRows`
- **Custom error types** (`ErrRecordNotFound`, `ErrEditConflict`)

Example error response:
```json
{
  "error": "the requested resource could not be found"
}
```

### Graceful Shutdown

The server implements graceful shutdown with:

- **30-second timeout** for existing connections
- **Signal handling** for SIGINT and SIGTERM
- **Connection draining** before shutdown
- **Resource cleanup** and logging

```go
srv.Shutdown(context.WithTimeout(context.Background(), 30*time.Second))
```

### Logging

Structured JSON logging for all events:

```json
{
  "level": "ERROR",
  "time": "2026-02-07T06:52:13Z",
  "message": "failed to insert ticket",
  "properties": {
    "request_method": "POST",
    "request_url": "/v1/buy-ticket"
  },
  "trace": "goroutine 49 [running]:\n..."
}
```

## Production Considerations

### Security
- ✅ JWT-based authentication
- ✅ Password hashing with bcrypt
- ✅ Input validation and sanitization
- ✅ SQL injection prevention (parameterized queries)
- ⚠️ TODO: Rate limiting
- ⚠️ TODO: CORS configuration
- ⚠️ TODO: HTTPS enforcement
- ⚠️ TODO: API key management

### Performance
- ✅ Connection pooling
- ✅ Efficient LEFT JOIN queries
- ✅ Prepared statement reuse
- ⚠️ TODO: Caching layer (Redis)
- ⚠️ TODO: Database indexing optimization
- ⚠️ TODO: Query optimization
- ⚠️ TODO: Load balancing

### Monitoring
- ✅ Structured JSON logging
- ✅ Request/error tracing with stack traces
- ⚠️ TODO: Metrics collection (Prometheus)
- ⚠️ TODO: Distributed tracing (Jaeger)
- ⚠️ TODO: Health check endpoints
- ⚠️ TODO: APM integration

### Deployment
- ✅ Cross-compilation support
- ✅ Vendored dependencies
- ⚠️ TODO: Docker containerization
- ⚠️ TODO: Kubernetes manifests
- ⚠️ TODO: CI/CD pipeline
- ⚠️ TODO: Database backup strategy

## Contact

Abraham Mayowa - [@AbrahamMayowa](https://github.com/AbrahamMayowa)

Project Link: [https://github.com/AbrahamMayowa/ticketmania](https://github.com/AbrahamMayowa/ticketmania)

---

**Built with ❤️ using Go**
