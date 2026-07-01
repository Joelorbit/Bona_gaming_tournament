# Bona Tournament Technical Documentation

This document explains how the app works locally and exactly where AddisPay is used. It is written for technical review, including questions about what was integrated, where it was integrated, and how payment verification works.

## 1. System Summary

Bona Tournament is a full-stack tournament platform.

Frontend:

- React 19
- Vite
- Tailwind CSS
- Supabase JS

Backend:

- Go
- chi router
- pgx/Postgres
- Supabase Auth verification
- AddisPay Hosted Checkout client

Database:

- Supabase Postgres or normal Postgres
- SQL migrations in `backend/migrations`

Core app features:

- Supabase email/password login
- user profiles
- tournament creation
- tournament registration
- paid entry fees through AddisPay Hosted Checkout
- webhook-first payment confirmation
- brackets, matches, disputes, notifications
- winner payout detail collection
- organizer payout marking
- admin panel for payments, payouts, tournaments, and audit log

## 2. Local URLs

Frontend:

```text
http://localhost:5173
```

Backend:

```text
http://localhost:8080
```

Backend health:

```text
http://localhost:8080/health
```

API base:

```text
http://localhost:8080/api/v1
```

AddisPay webhook route:

```text
POST http://localhost:8080/payments/webhook
```

The webhook route is outside `/api/v1` because AddisPay calls it without a Supabase Bearer token.

## 3. Required Local Setup

You need:

- Go
- Node.js and npm
- Supabase project or Postgres database
- `golang-migrate` if using `make migrate-up`
- AddisPay sandbox/test credentials
- ngrok or another tunnel if AddisPay must call your local backend

AddisPay cannot call `localhost` directly. For real local webhook testing, expose port `8080` through a public tunnel.

## 4. Database Migrations

Run migrations in this order:

```text
backend/migrations/000001_create_tables.up.sql
backend/migrations/000002_profile_expansion.up.sql
backend/migrations/000003_notifications.up.sql
backend/migrations/000004_tournament_v2.up.sql
backend/migrations/000005_match_disputes.up.sql
backend/migrations/000006_payouts.up.sql
backend/migrations/000007_audit_log.up.sql
backend/migrations/000008_match_status_length.up.sql
backend/migrations/000009_payout_details.up.sql
backend/migrations/000010_payment_reconciliation.up.sql
backend/migrations/000011_payment_refunds.up.sql
```

Do not run `.down.sql` files unless rolling back.

Migration `000010_payment_reconciliation` adds:

- `provider_status`
- `provider_payment_id`
- `verified_at`
- `webhook_received_at`
- `failure_reason`

These fields support admin reconciliation of AddisPay payment events.

Migration `000011_payment_refunds` adds manual refund tracking for cancelled tournaments:

- `refund_status`
- `refund_reason`
- `refund_requested_at`
- `refunded_at`
- `refunded_by`

## 5. Backend Environment

Create `backend/.env`:

```bash
cd backend
cp .env.example .env
```

Example local backend env:

```env
PORT=8080
DATABASE_URL=postgres://postgres.<project-ref>:<password>@aws-0-<region>.pooler.supabase.com:5432/postgres
SUPABASE_URL=https://<project-ref>.supabase.co
SUPABASE_ANON_KEY=<your-supabase-anon-key>

ADDISPAY_API_KEY=<your-addispay-test-api-key>
ADDISPAY_BASE_URL=https://uat.api.addispay.et
ADDISPAY_WEBHOOK_SECRET=<same-secret-used-by-addispay>

ADDISPAY_REDIRECT_URL=http://localhost:5173
ADDISPAY_CANCEL_URL=http://localhost:5173
ADDISPAY_SUCCESS_URL=http://localhost:5173
ADDISPAY_ERROR_URL=http://localhost:5173

ENVIRONMENT=development
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
```

Notes:

- Use the Supabase Session Pooler URL if your network cannot reach Supabase over IPv6.
- `ADDISPAY_API_KEY` must be real test/sandbox credentials.
- `ADDISPAY_WEBHOOK_SECRET` must match the secret AddisPay uses to sign webhook payloads.
- Do not commit `.env` files.

## 6. Frontend Environment

Create `frontend/.env`:

```bash
cd frontend
cp .env.example .env
```

Example local frontend env:

```env
VITE_SUPABASE_URL=https://<project-ref>.supabase.co
VITE_SUPABASE_ANON_KEY=<your-supabase-anon-key>
VITE_API_BASE_URL=http://localhost:8080
```

## 7. Run Locally

Backend:

```bash
cd backend
go mod tidy
go run ./cmd/server
```

Frontend:

```bash
cd frontend
npm install
npm run dev
```

Open:

```text
http://localhost:5173
```

## 8. AddisPay Integration Summary

The app uses AddisPay Hosted Checkout for paid tournament entry fees.

AddisPay is used for:

- creating hosted checkout orders
- redirecting players to AddisPay checkout
- receiving signed payment webhooks
- confirming local payment and registration status after verification

AddisPay is not used for:

- winner prize payouts
- refunds
- saved payment methods
- in-app card or wallet credential collection

Winner payouts are manual. The winner submits Telebirr or bank details, the organizer pays outside the app, then marks the payout paid.

## 9. AddisPay Code Locations

Frontend payment start:

```text
frontend/src/pages/TournamentDetails.tsx
```

Relevant functions:

- `handleRegister()`
- `handlePay()`

The frontend calls:

```text
POST /api/v1/payments/create
```

If the backend returns `payment_url`, the browser is redirected:

```ts
window.location.href = pay.payment_url
```

Frontend return page:

```text
frontend/src/pages/PaymentReturn.tsx
```

This reads:

- `payment_id`
- `tournament_id`
- `status`
- `token`

Then it calls:

```text
POST /api/v1/payments/return
```

Important: the return page does not mark payments paid. It only validates the return token and shows the current stored payment status. Final payment confirmation comes from the verified webhook.

Backend payment routes:

```text
backend/internal/router/router.go
```

Authenticated routes:

```text
POST /api/v1/payments/create
POST /api/v1/payments/return
GET  /api/v1/payments/status/{id}
```

Public AddisPay webhook:

```text
POST /payments/webhook
```

Backend payment handler:

```text
backend/internal/modules/payment/handler.go
```

Backend payment business logic:

```text
backend/internal/modules/payment/service.go
```

AddisPay HTTP client and webhook signature verification:

```text
backend/pkg/addispay/client.go
```

Payment repository:

```text
backend/internal/repository/payment.go
```

Payment SQL reference:

```text
backend/sqlc/queries/payment.sql
```

## 10. AddisPay Checkout Request

The backend sends checkout creation requests to:

```text
POST {ADDISPAY_BASE_URL}/checkout-api/v1/create-order
```

Headers:

```text
Auth: <ADDISPAY_API_KEY>
Content-Type: application/json
Accept: application/json
```

The request contains the local payment ID as both `nonce` and `tx_ref`.

Important fields:

```json
{
  "message": "Bona tournament entry payment",
  "data": {
    "redirect_url": "http://localhost:5173/payments/return?...",
    "cancel_url": "http://localhost:5173/payments/return?...",
    "success_url": "http://localhost:5173/payments/return?...",
    "error_url": "http://localhost:5173/payments/return?...",
    "order_reason": "Tournament entry: <tournament title>",
    "currency": "ETB",
    "email": "<player email>",
    "first_name": "<first name>",
    "last_name": "<last name>",
    "nonce": "<local payment id>",
    "total_amount": "<entry fee>",
    "tx_ref": "<local payment id>",
    "order_detail": {
      "amount": "<entry fee>",
      "description": "Tournament entry: <tournament title>"
    }
  }
}
```

Using the local payment ID as the AddisPay reference makes reconciliation direct:

```text
AddisPay reference -> payments.addispay_ref -> payments.id
```

## 11. Payment Flow

1. Player opens a paid tournament.
2. Player clicks register/pay.
3. Frontend calls:

```text
POST /api/v1/tournaments/{id}/join
```

4. Backend creates registration:

```text
registrations.payment_status = pending
```

5. Frontend calls:

```text
POST /api/v1/payments/create
```

Body:

```json
{
  "tournament_id": "tournament-id"
}
```

6. Backend creates local payment:

```text
payments.status = pending
```

7. Backend calls AddisPay create-order.
8. AddisPay returns hosted checkout URL.
9. Backend returns `payment_url`.
10. Frontend redirects player to AddisPay.
11. Player completes/cancels/fails payment on AddisPay.
12. AddisPay redirects player to `/payments/return`.
13. Frontend calls `/api/v1/payments/return`.
14. Backend validates return token and returns current payment status.
15. AddisPay sends signed webhook to `/payments/webhook`.
16. Backend verifies signature.
17. Backend validates amount and currency.
18. Backend marks payment paid and registration paid in one transaction.
19. Admin sees payment as verified in `/admin`.

## 12. Payment State Machine

Local payment statuses:

```text
pending
paid
failed
refunded
```

Normal transitions:

```text
pending -> paid
pending -> failed
```

Safety rules:

- frontend return URL cannot mark payment paid
- only verified webhook can mark payment paid
- duplicate webhooks are safe
- a `paid` payment is not downgraded by later failed webhook data
- payment and registration confirmation happen in one database transaction

When webhook confirms payment:

```text
payments.status = paid
payments.verified_at = NOW()
payments.webhook_received_at = NOW()
registrations.payment_status = paid
```

When webhook fails validation:

```text
payments.status = failed
payments.failure_reason = <reason>
payments.webhook_received_at = NOW()
```

## 13. Webhook Verification

Webhook route:

```text
POST /payments/webhook
```

Expected webhook fields:

```json
{
  "reference": "local-payment-id",
  "status": "paid",
  "amount": 100,
  "currency": "ETB",
  "signature": "hex-hmac-signature",
  "payment_id": "provider-payment-id"
}
```

The parser also accepts aliases:

- `tx_ref`
- `payment_status`
- `total_amount`
- `uuid`

Current signature base string:

```text
reference|status|amount
```

Algorithm:

```text
HMAC-SHA256(ADDISPAY_WEBHOOK_SECRET, reference + "|" + status + "|" + amount)
```

Signature must be hex encoded.

If AddisPay's internal or current documentation uses a different signature base string, update:

```text
backend/pkg/addispay/client.go
```

Function:

```text
VerifyWebhookSignature
```

## 14. Local Webhook Testing

AddisPay cannot call `localhost`. Use a tunnel.

Example:

```bash
ngrok http 8080
```

If ngrok gives:

```text
https://abc123.ngrok-free.app
```

Set AddisPay webhook URL to:

```text
https://abc123.ngrok-free.app/payments/webhook
```

Keep these local:

```text
Frontend: http://localhost:5173
Backend:  http://localhost:8080
Webhook:  https://abc123.ngrok-free.app/payments/webhook
```

The player still returns to the local frontend because the backend sends local return URLs to AddisPay.

## 15. Admin Panel

Admin page:

```text
http://localhost:5173/admin
```

There is no separate admin login. Log in normally and set your profile role to `admin`:

```sql
UPDATE profiles
SET role = 'admin'
WHERE email = 'your-email@example.com';
```

Then sign out and sign in again.

Admin routes:

```text
GET /api/v1/admin/stats
GET /api/v1/admin/tournaments
GET /api/v1/admin/payments
GET /api/v1/admin/payouts
GET /api/v1/admin/audit
```

Admin payments panel shows:

- local payment ID
- player ID
- tournament ID
- amount/currency
- local status
- AddisPay reference
- provider status
- provider payment ID
- webhook received time
- verified time
- failure reason
- stale pending warning after 30 minutes

Admin payout panel intentionally does not show bank account or Telebirr details. Sensitive winner payout details are visible only to the hosting organizer in `/dashboard`.

Use this panel to answer:

- Did checkout creation happen?
- Did webhook arrive?
- Did amount/currency match?
- Was payment verified?
- Why did a payment fail?

## 16. Winner Payout Flow

Winner payouts are manual and not sent through AddisPay.

Flow:

1. Tournament completes.
2. Backend creates payout row.
3. Winner opens `/me/payouts`.
4. Winner submits Telebirr or bank payout details.
5. Organizer opens `/dashboard`.
6. Organizer sees payout details.
7. Organizer pays winner outside the app.
8. Organizer clicks `Mark paid`.
9. Winner receives notification.
10. Admin can audit payout records in `/admin`.

Telebirr fields:

- phone number
- Telebirr number

Bank fields:

- phone number
- bank name
- bank account name
- bank account number

## 17. Important API Routes

Payments:

```text
POST /api/v1/payments/create
POST /api/v1/payments/return
GET  /api/v1/payments/status/{id}
POST /payments/webhook
```

Payouts:

```text
GET  /api/v1/me/payouts
GET  /api/v1/me/organizer-payouts
POST /api/v1/payouts/{id}/details
POST /api/v1/payouts/{id}/mark-paid
```

Admin:

```text
GET /api/v1/admin/stats
GET /api/v1/admin/tournaments
GET /api/v1/admin/payments
GET /api/v1/admin/payouts
GET /api/v1/admin/audit
```

## 18. End-To-End Local Test

1. Start backend on `http://localhost:8080`.
2. Start frontend on `http://localhost:5173`.
3. Start tunnel to backend port `8080`.
4. Configure AddisPay webhook:

```text
https://your-tunnel-domain/payments/webhook
```

5. Sign up or log in.
6. Create a paid tournament.
7. Open registration.
8. Register as a player.
9. Click payment.
10. Confirm browser redirects to AddisPay Hosted Checkout.
11. Complete checkout.
12. Confirm browser returns to `/payments/return`.
13. Confirm webhook arrives in backend logs.
14. Open `/admin`.
15. Confirm payment shows:

```text
status = paid
verified
webhook_received_at exists
verified_at exists
```

16. Open tournament participants.
17. Confirm player is `paid`.
18. Complete tournament.
19. Winner opens `/me/payouts`.
20. Winner submits payout details.
21. Organizer opens `/dashboard`.
22. Organizer sees payout details and marks payout paid.

## 19. Common Problems

Payment stays pending:

- webhook did not reach backend
- tunnel URL is wrong
- webhook path should be `/payments/webhook`, not `/api/v1/payments/webhook`
- webhook secret does not match
- AddisPay did not send expected reference/status/amount

Payment failed with amount mismatch:

- webhook amount differs from local `payments.amount`
- check `/admin` payment `failure_reason`

Payment failed with currency mismatch:

- webhook currency differs from local `payments.currency`
- expected currency is `ETB`

Frontend cannot call backend:

- check `VITE_API_BASE_URL`
- check `ALLOWED_ORIGINS`
- check backend is running

Admin page says access required:

- `profiles.role` is not `admin`
- update SQL role and sign in again

Checkout creation fails:

- check `ADDISPAY_API_KEY`
- check `ADDISPAY_BASE_URL`
- check backend logs
- check admin payment `failure_reason`

## 20. Verification Commands

Backend:

```bash
cd backend
go test ./...
```

If Go cache is restricted:

```bash
cd backend
GOCACHE=/tmp/bona-go-build-cache go test ./...
```

Frontend:

```bash
cd frontend
npm run build
```

## 21. AddisPay Review Summary

Use this explanation when asked where and how AddisPay is used:

Bona integrates AddisPay as a Hosted Checkout provider for tournament entry fees. The app creates a local pending payment, sends an AddisPay create-order request with the local payment ID as `tx_ref` and `nonce`, redirects the player to AddisPay's hosted checkout URL, and waits for AddisPay's signed webhook to finalize payment. The frontend return URL is not trusted as proof of payment. The backend verifies webhook signature, amount, and currency before marking the payment paid. Payment confirmation and registration confirmation happen in one database transaction. Admins can review provider status, webhook timestamps, verification timestamps, and failure reasons from the admin panel.

Code locations:

- Checkout request: `backend/internal/modules/payment/service.go`
- AddisPay client: `backend/pkg/addispay/client.go`
- Webhook handler: `backend/internal/modules/payment/handler.go`
- Payment repository: `backend/internal/repository/payment.go`
- Frontend payment start: `frontend/src/pages/TournamentDetails.tsx`
- Frontend return page: `frontend/src/pages/PaymentReturn.tsx`
- Admin payment review: `frontend/src/pages/Admin.tsx`

Reliability choices:

- Hosted Checkout avoids direct handling of payment credentials.
- Webhook-first confirmation prevents forged return URLs from marking registrations paid.
- HMAC verification protects webhook authenticity.
- Amount/currency validation protects against mismatched events.
- Idempotent webhook handling protects against duplicate callbacks.
- Database transaction keeps payment and registration state consistent.
