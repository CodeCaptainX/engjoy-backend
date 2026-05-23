# Engjoy Backend (Sentence Miner)

This project is the backend for Engjoy, a language learning application focused on sentence mining and AI-driven analysis.

## Tech Stack
- **Language**: Go 1.25+
- **Web Framework**: [Gofiber/fiber](https://gofiber.io/)
- **Database**: PostgreSQL (pgx/v5, sqlx)
- **Migrations**: Goose
- **AI Integration**: Google Gemini API (for sentence analysis and TTS)

## Project Structure
- `config/`: Application configuration and database migrations.
- `handler/`: Global route registration.
- `internal/`: Core business logic organized by module.
    - `auth/`: Authentication logic (including Google OAuth2).
    - `user/`: User management.
    - `sentences/`: Sentence mining, AI analysis, and repository logic.
- `pkg/`: Shared utilities, constants, and HTTP helpers.
- `routers/`: Fiber application initialization and middleware.

## Conventions
- **Module Structure**: Every module MUST have the following layers:
    - `handler.go`: Responsible for extracting parameters from the request and returning the final data/response.
    - `service.go`: Responsible for business logic, external API calls, and orchestrating repositories.
    - `repository.go`: Responsible strictly for database interactions (using standard `db` methods, no `context.Context`).
    - `model.go`: Responsible for struct definitions and receiver functions (e.g., for parameter extraction/validation).
- **File Organization**: 
    - For simple modules, use single files (e.g., `handler.go`).
    - For complex modules with many functions, create directories (e.g., `handler/`, `service/`) and split logic into separate files (e.g., `handler/user_handler.go`, `handler/admin_handler.go`).
- **Naming Convention**: Use the following names for standard CRUD-like operations:
    - `show`: List multiple records.
    - `showOne`: Fetch a single record.
    - `update`: Modify an existing record.
    - `delete`: Remove a record.
- **Errors**: Standardized JSON error responses via `pkg/http/response`.
- **Logging**: Uses `rs/zerolog`.
- **Auth**: Currently uses placeholder 32-byte hex tokens.
- **Context Usage**: `context.Context` should ONLY be used for Redis operations or External API calls (like Gemini). Do NOT use it for database queries.

## Database Rules
- **Table Naming**: Every table name MUST start with the `tbl_` prefix.
- **Schema Changes**: NEVER use `ALTER TABLE` directly. Ask the user first.
- **Mandatory Columns**: Every table MUST include: `id`, `uuid`, `status_id`, `order`, `created_by`, `created_at`, `updated_by`, `updated_at`, `deleted_by`, `deleted_at`.

## Authentication Implementation
- **Email/Password**: standard bcrypt hashing.
- **Google OAuth2**: Integration via `golang.org/x/oauth2`.
- **Session Management**: Uses **JWT (JSON Web Tokens)** for stateless authentication. 
- **Session Tracking**: The `tbl_users` table includes a `login_session` field to track the latest active session.
- **Frontend**: Tokens should be stored in `localStorage` on the client side.

## Security Rules
- **Public APIs**: APIs that show sentences are "free" but restricted to **website-only** access. Implement this using CORS (AllowOrigins) and/or Origin validation to prevent unauthorized third-party usage.
