# Chirpy

A Twitter-like microblogging REST API built with Go and PostgreSQL.

## What It Does

Chirpy is a backend API for posting and managing short messages (chirps). It features:

- User authentication with JWT tokens
- Post chirps (max 140 characters) with profanity filtering
- Query, sort, and delete chirps
- Webhook integration for premium upgrades
- Token refresh and revocation

## Why Use It

- Clean, minimal Go implementation using only standard library + PostgreSQL
- Demonstrates modern API patterns: JWT auth, REST endpoints, SQL migrations
- Perfect starting point for learning Go web development or building a microblogging platform

## Installation & Setup

**Prerequisites:** Go 1.25+, PostgreSQL, [goose](https://github.com/pressly/goose) for migrations

1. Clone and install dependencies:
```bash
git clone https://github.com/x6Nenko/Chirpy.git
cd Chirpy
go mod download
```

2. Set up environment variables (create `.env`):
```
DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev
SECRET=your-jwt-secret
POLKA_KEY=your-webhook-key
```

3. Run database migrations:
```bash
goose -dir sql/schema postgres "$DB_URL" up
```

4. Start the server:
```bash
go run .
```

API runs on `http://localhost:8080`

## API Endpoints

**Users:**
- `POST /api/users` - Create user
- `POST /api/login` - Login
- `PUT /api/users` - Update user (authenticated)
- `POST /api/refresh` - Refresh access token
- `POST /api/revoke` - Revoke refresh token

**Chirps:**
- `POST /api/chirps` - Create chirp (authenticated)
- `GET /api/chirps` - Get all chirps (optional `?author_id=` and `?sort=desc`)
- `GET /api/chirps/{id}` - Get single chirp
- `DELETE /api/chirps/{id}` - Delete chirp (authenticated)

**Webhooks:**
- `POST /api/polka/webhooks` - Handle premium upgrade webhooks

**Health & Admin:**
- `GET /api/healthz` - Health check
- `GET /admin/metrics` - View metrics (dev only)
- `POST /admin/reset` - Reset database (dev only)
