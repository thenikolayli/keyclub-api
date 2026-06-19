# Key Club API

Backend API for JHS Key Club. Handles user authentication via magic-link email login and cookie-based sessions.

## Modules

### auth

User authentication and authorization. Manages users, pending logins, and server-side sessions. Exposes HTTP handlers for starting a login, waiting for email verification, verifying a magic link, and logging out. Session cookies hold opaque tokens; only hashed tokens are stored in the database.

### email

Sends transactional HTML email over SMTP. Loads rendered templates from disk and fills them with request-specific data (for example, a pending-login magic link).

### internal

Application wiring: config loading, SQLite database connection and migrations, HTTP server setup, and structured logging. Also includes small CLI utilities such as `adduser` for creating users directly in the database.

### maizzle

HTML email templates built with [Maizzle](https://maizzle.com) and Tailwind CSS. Source templates live here; run the Maizzle build to produce the static HTML files that the `email` package reads at runtime (`maizzle/build_production`).

## Environment variables

The app loads a `.env` file at startup (via `godotenv`). All variables below are required unless noted otherwise.

| Variable | Description |
| --- | --- |
| `SMTP_PASSWORD` | App password for the Gmail account used to send transactional email. |
| `DB_SQLITE_PATH` | Filesystem path to the SQLite database file (for example, `data/keyclub.db`). |
| `DB_MIGRATIONS_PATH` | Path to the SQL migration files directory (for example, `migrations`). |
| `FRONTEND_URL` | Base URL of the frontend app. Used to build magic-login links in email (for example, `https://jhskeyclub.com`). |
| `API_URL` | Base URL of this API. Loaded into config; not used by handlers yet. |

## Running

```bash
go run .
```

Create a user from the command line:

```bash
go run ./internal/cmd/adduser --email <email> --first <first> --last <last> [--role member]
```

Build email templates:

```bash
cd maizzle && npm install && npm run build
```
