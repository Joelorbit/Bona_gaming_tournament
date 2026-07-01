# Bona Tournament

Full-stack tournament platform with Supabase auth, Go backend, React frontend, AddisPay Hosted Checkout for entry fees, admin reconciliation, and manual winner payout tracking.

## Quick Local Start

Backend:

```bash
cd backend
cp .env.example .env
go mod tidy
make migrate-up
go run ./cmd/server
```

Frontend:

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

Default URLs:

- Frontend: `http://localhost:5173`
- Backend: `http://localhost:8080`
- Health: `http://localhost:8080/health`

## Documentation

Read [doc.md](doc.md) for full setup, local AddisPay webhook testing, payment flow, admin access, payout flow, and review notes.

## Verification

```bash
cd backend
go test ./...
```

```bash
cd frontend
npm run build
```
