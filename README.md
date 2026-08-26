# SportCabinet REST API

Go backend for **SportCabinet**, a web application for viewing and analyzing sports statistics across seasons and territories.

SportCabinet provides tools for exploring sports data, comparing opponents, viewing summary tables, and analyzing team and competition statistics.

**Live application:** https://www.sportcabinet.ru/

## Tech Stack

- Go
- chi
- REST API
- SQLite / SQLCipher
- JWT authentication
- bcrypt
- JSON
- structured logging (`slog`)
- Docker / Docker Compose
- Nginx
- HTTPS

## Architecture

The backend is organized into several layers:

- `cmd/server` — application entry point and HTTP server setup
- `internal/http-server/handlers` — REST API handlers
- `internal/http-server/middleware` — authentication, authorization and logging middleware
- `internal/storage` — storage abstraction
- `internal/storage/sqlite` — SQLite implementation
- `internal/jwt` — JWT token handling
- `internal/config` — application configuration
- `internal/lib` — shared API response and logging utilities

## Features

- REST API for sports statistics
- Statistics by season and territory
- Team and competition data
- Opponent comparison
- Summary tables
- JWT authentication
- User permissions
- Password verification with bcrypt
- SQLite data storage
- SQLCipher database encryption support
- Request logging and log rotation
- CORS configuration
- Docker deployment

## Authentication

The project includes JWT-based authentication and authorization.

The API can issue access tokens after successful login. Middleware is implemented for JWT validation and permission checks.

Authentication support is currently prepared for further use in the application; most public SportCabinet functionality does not currently require authentication.

## Configuration

Local configuration is stored outside the repository.

An example configuration is available at:

`config/local.example.yaml`

Create `config/local.yaml` based on the example and provide your own settings and JWT secret.

## Docker

The backend can be built and run with Docker Compose:

    docker compose up -d --build

The service listens on port `8082` and is bound to localhost on the host machine.

In production, the application runs behind Nginx with HTTPS.

## Frontend

The web frontend is implemented with React and TypeScript.

Frontend repository:
https://github.com/Alf73x/cabinet-react

## Project Structure

    cmd/
      dbcrypt/
      server/

    config/

    internal/
      config/
      http-server/
        handlers/
        middleware/
      jwt/
      lib/
      storage/
        sqlite/

    storage/

## Status

SportCabinet is under active development. New statistics, analysis tools, and backend functionality are being added.