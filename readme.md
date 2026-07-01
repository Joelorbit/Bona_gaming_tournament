# Bona Tournament

Tournament platform with prize pools. Go backend + React frontend.

## Stack

- **Backend**: Go 1.26, chi router, pgx (Postgres), Supabase auth, AddisPay
- **Frontend**: React 19, Vite, Tailwind, Supabase JS

## Layout

```
backend/
  cmd/server/         # main entrypoint
  internal/
    config/           # env loading
    db/               # postgres pool
    middleware/       # auth (Supabase JWT), CORS, logging
    modules/          # auth, user, tournament, registration, bracket, match, payment
    repository/       # hand-written query layer (sqlc-equivalent)
    router/           # chi routes
    utils/            # JSON helpers
  migrations/         # golang-migrate SQL
  pkg/
    addispay/         # AddisPay Hosted Checkout client
    supabase/         # Supabase auth client
  sqlc/queries/       # source-of-truth SQL queries (kept in sync with repository/)
  .env.example

frontend/
  src/
    components/{layout,ui}/
    contexts/AuthContext.tsx
    lib/{api,supabase,utils}.ts
    pages/
  .env.example
```

## Setup

### Backend

```bash
cd backend
cp .env.example .env   # fill in DATABASE_URL, SUPABASE_*, ADDISPAY_*
go mod tidy

# Apply migrations (requires golang-migrate)
make migrate-up

make run               # http://localhost:8080
```

### Frontend

```bash
cd frontend
cp .env.example .env   # fill in VITE_SUPABASE_*
npm install
npm run dev            # http://localhost:5173
```

## Notes

- The `pkg/supabase` client verifies user JWTs by calling `auth.getUser` on each
  request — fine for low traffic, switch to local JWT verification (JWKS) before
  production scale.
- AddisPay uses Hosted Checkout (`/checkout-api/v1/create-order`). See
  `ADDISPAY_INTEGRATION.md` for env setup, redirect URLs, webhook expectations,
  and the exact test flow.
- The `internal/repository/` layer is hand-written and mirrors `sqlc/queries/`.
  Keep them in sync if you add new queries — or wire up sqlc and regenerate.
- All `/api/v1` routes require a Supabase Bearer token. The frontend `api`
  helper attaches it automatically.
