# Bona Tournament End-to-End Setup

Use this checklist to get the fullstack app working with Supabase Auth, Supabase Postgres, the Go API, the React frontend, and AddisPay payments.

## Your Only Setup Work

You should not need to edit frontend or backend source code.

Your setup work is only:

1. Run the SQL migrations in Supabase.
2. Configure Supabase Auth redirect URLs.
3. Fill `backend/.env`.
4. Fill `frontend/.env`.
5. Put the AddisPay webhook URL and return URLs in the AddisPay dashboard.
6. Start backend and frontend.

If something requires editing `.go`, `.tsx`, `.ts`, or SQL files after this, treat it as an integration bug and fix it in code instead of changing random app files.

## 1. Supabase Project

1. Create a Supabase project.
2. In Supabase Auth, enable email/password sign-in.
3. Set Auth redirect URLs:
   - Local frontend: `http://localhost:5173`
   - Reset password: `http://localhost:5173/reset-password`
   - Production frontend: `https://your-frontend-domain.com`
   - Production reset password: `https://your-frontend-domain.com/reset-password`
4. Copy these values for env setup:
   - Project URL
   - Anon public key
   - Postgres connection string
   - Session Pooler connection string if your local network does not have IPv6

## 2. Run Database SQL

In Supabase SQL Editor, run the migration files in this exact order:

1. `backend/migrations/000001_create_tables.up.sql`
2. `backend/migrations/000002_profile_expansion.up.sql`
3. `backend/migrations/000003_notifications.up.sql`
4. `backend/migrations/000004_tournament_v2.up.sql`
5. `backend/migrations/000005_match_disputes.up.sql`
6. `backend/migrations/000006_payouts.up.sql`
7. `backend/migrations/000007_audit_log.up.sql`
8. `backend/migrations/000008_match_status_length.up.sql`

Do not run the `.down.sql` files unless you want to roll back.

## 3. Backend Environment

Create `backend/.env` from `backend/.env.example`:

```env
PORT=8080
DATABASE_URL=postgres://postgres.<project-ref>:<password>@aws-0-<region>.pooler.supabase.com:5432/postgres
SUPABASE_URL=https://<project-ref>.supabase.co
SUPABASE_ANON_KEY=<your-supabase-anon-key>
ADDISPAY_API_KEY=<your-addispay-api-key>
ADDISPAY_BASE_URL=https://uat.api.addispay.et
ADDISPAY_WEBHOOK_SECRET=<your-addispay-webhook-secret>
ADDISPAY_REDIRECT_URL=http://localhost:5173
ADDISPAY_CANCEL_URL=http://localhost:5173
ADDISPAY_SUCCESS_URL=http://localhost:5173
ADDISPAY_ERROR_URL=http://localhost:5173
ENVIRONMENT=development
ALLOWED_ORIGINS=http://localhost:5173
```

Use the Session Pooler URL from Supabase Dashboard -> Connect -> Session pooler for local development on IPv4-only networks. The direct URL, `postgresql://postgres:<password>@db.<project-ref>.supabase.co:5432/postgres`, requires IPv6 unless your Supabase project has the IPv4 add-on enabled.

For production, set:

```env
ENVIRONMENT=production
ALLOWED_ORIGINS=https://your-frontend-domain.com
ADDISPAY_BASE_URL=https://api.addispay.et
ADDISPAY_REDIRECT_URL=https://your-frontend-domain.com
ADDISPAY_CANCEL_URL=https://your-frontend-domain.com
ADDISPAY_SUCCESS_URL=https://your-frontend-domain.com
ADDISPAY_ERROR_URL=https://your-frontend-domain.com
```

The backend turns those frontend origins into per-payment return URLs like `/payments/return?payment_id=...&tournament_id=...&status=success&token=...`.

Never commit real `.env` files.

## 4. Frontend Environment

Create `frontend/.env` from `frontend/.env.example`:

```env
VITE_SUPABASE_URL=https://<project-ref>.supabase.co
VITE_SUPABASE_ANON_KEY=<your-supabase-anon-key>
VITE_API_BASE_URL=http://localhost:8080
```

For production:

```env
VITE_API_BASE_URL=https://your-backend-domain.com
```

## 5. AddisPay Setup

This project uses AddisPay Hosted Checkout. More detail is in `ADDISPAY_INTEGRATION.md`.

In AddisPay dashboard:

1. Add your backend webhook URL:
   - Local testing needs a public tunnel such as ngrok.
   - Production URL: `https://your-backend-domain.com/api/v1/payments/webhook`
2. Use the same webhook secret in AddisPay and `ADDISPAY_WEBHOOK_SECRET`.
3. Confirm AddisPay sends:
   - `reference`
   - `status`
   - `amount`
   - `currency`
   - `signature`
4. The backend verifies the webhook signature using:

```text
reference|status|amount
```

The backend only marks a registration as paid after the webhook signature is valid and the amount/currency match the saved payment.

## 6. Run Locally

Terminal 1:

```bash
cd backend
go mod tidy
go run ./cmd/server
```

Backend should start on:

```text
http://localhost:8080
```

Health check:

```bash
curl http://localhost:8080/health
```

Terminal 2:

```bash
cd frontend
npm install
npm run dev
```

Frontend should start on:

```text
http://localhost:5173
```

## 7. End-to-End Test Flow

Use this exact flow after env and SQL are ready:

1. Open `http://localhost:5173`.
2. Sign up with email/password.
3. Log in.
4. Create a tournament.
5. Open registration for the tournament.
6. Register as a player.
7. If the tournament has an entry fee, start payment.
8. Complete the AddisPay checkout.
9. Confirm AddisPay sends the webhook to the backend.
10. Check the participant now has `payment_status = paid`.
11. Close registration.
12. Generate bracket.
13. Submit and confirm match results.
14. Mark tournament completed.
15. Confirm payout row is created.
16. Organizer marks payout paid.

## 8. Verification Commands

Run before deploying:

```bash
cd frontend
npm run build
```

```bash
cd backend
GOCACHE=/tmp/bona-go-build-cache go test ./...
GOCACHE=/tmp/bona-go-build-cache go vet ./...
GOCACHE=/tmp/bona-go-build-cache go test -race ./...
```

Expected result: all commands pass.

## 9. Production Checklist

Before production launch:

1. Backend is deployed on HTTPS.
2. Frontend is deployed on HTTPS.
3. Supabase Auth redirect URLs include production frontend URLs.
4. `ALLOWED_ORIGINS` contains only the production frontend domain.
5. AddisPay webhook points to the production backend webhook URL.
6. AddisPay webhook secret matches backend env.
7. Database migrations are applied once in order.
8. `.env` values are stored in the hosting provider secrets/settings.
9. Run the end-to-end test flow with a small real payment.
10. Confirm notifications, brackets, payments, disputes, audit logs, and payouts all work.

## 10. Known Follow-Ups

These are not blockers for a local end-to-end test, but should be handled before serious production traffic:

1. Replace per-request Supabase Auth network verification with local JWT/JWKS verification.
2. Add more backend service tests around payments, registrations, brackets, and payouts.
3. Add frontend tests for critical user flows.
4. Split the frontend bundle with lazy route imports if bundle size becomes an issue.
5. Confirm AddisPay's real signature format matches `reference|status|amount`; update `backend/pkg/addispay/client.go` if their production docs differ.
