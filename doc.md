# Backend AddisPay Payment Documentation

This document explains only the backend payment integration for Bona Tournament. It is written so you can present the AddisPay work, defend the design, and answer technical questions from simple to advanced.

The key idea:

```text
The backend creates a local pending payment, sends an AddisPay Hosted Checkout order, redirects the user through the returned checkout URL, and only marks the payment paid after a signed AddisPay webhook is verified.
```

The frontend return page is not trusted as proof of payment. The verified backend webhook is the source of truth.

## 1. What AddisPay Is Used For

AddisPay is used only for tournament entry fee collection.

Used:

- creating a Hosted Checkout order
- redirecting the player to AddisPay's checkout page
- receiving AddisPay webhook events
- verifying the webhook signature
- updating local payment and registration status after verification

Not used:

- prize payouts to winners
- automatic refunds through AddisPay
- saved cards or saved wallets
- direct card handling inside this backend

Why Hosted Checkout matters:

```text
The app does not collect card, wallet, or payment credentials. AddisPay hosts the checkout. This reduces payment security responsibility inside our own app.
```

## 2. Backend Code Map

These are the backend files you must know.

```text
backend/cmd/server/main.go
```

Starts the server, loads config, initializes the AddisPay client when the API key exists, and passes payment config to the router.

```text
backend/internal/config/config.go
```

Reads environment variables such as `ADDISPAY_API_KEY`, `ADDISPAY_BASE_URL`, and `ADDISPAY_WEBHOOK_SECRET`.

```text
backend/internal/router/router.go
```

Mounts authenticated payment routes and the public webhook route.

```text
backend/internal/modules/payment/handler.go
```

HTTP layer for payments. It decodes requests, calls the payment service, verifies webhook signatures, and returns JSON responses.

```text
backend/internal/modules/payment/service.go
```

Business logic. It creates local payment rows, builds AddisPay checkout requests, validates returns, handles webhooks, updates registrations, and manages refund marking.

```text
backend/pkg/addispay/client.go
```

Low-level AddisPay HTTP client. It sends the create-order request, parses AddisPay responses, parses webhook payloads, and verifies HMAC signatures.

```text
backend/internal/repository/payment.go
backend/sqlc/queries/payment.sql
```

Payment database queries and repository methods.

```text
backend/migrations/000001_create_tables.up.sql
backend/migrations/000010_payment_reconciliation.up.sql
backend/migrations/000011_payment_refunds.up.sql
```

Database schema for payment records, webhook reconciliation fields, and manual refund tracking.

## 3. Environment Configuration

Backend payment config comes from `backend/internal/config/config.go`.

Required for AddisPay:

```env
ADDISPAY_API_KEY=<real AddisPay API key>
ADDISPAY_BASE_URL=https://uat.api.addispay.et
ADDISPAY_WEBHOOK_SECRET=<secret used to verify webhooks>
ADDISPAY_REDIRECT_URL=http://localhost:5173
ADDISPAY_CANCEL_URL=http://localhost:5173
ADDISPAY_SUCCESS_URL=http://localhost:5173
ADDISPAY_ERROR_URL=http://localhost:5173
```

What each value means:

| Variable | Backend meaning |
| --- | --- |
| `ADDISPAY_API_KEY` | Sent to AddisPay in the `Auth` header when creating checkout orders. |
| `ADDISPAY_BASE_URL` | Base API URL. Local default is the AddisPay UAT URL. |
| `ADDISPAY_WEBHOOK_SECRET` | Shared secret used for HMAC verification of webhooks and return tokens. |
| `ADDISPAY_REDIRECT_URL` | Fallback frontend base URL used when a specific success/cancel/error URL is missing. |
| `ADDISPAY_SUCCESS_URL` | Frontend base URL for successful checkout return. Backend rewrites path to `/payments/return`. |
| `ADDISPAY_CANCEL_URL` | Frontend base URL for cancelled checkout return. Backend rewrites path to `/payments/return`. |
| `ADDISPAY_ERROR_URL` | Frontend base URL for failed checkout return. Backend rewrites path to `/payments/return`. |

Important production rules:

- `ADDISPAY_API_KEY` must not be committed.
- `ADDISPAY_WEBHOOK_SECRET` must not be empty in production.
- `ADDISPAY_BASE_URL` must point to the correct AddisPay environment: UAT for testing, production URL for live payments.
- Return URLs should use HTTPS in production.
- The webhook URL configured in AddisPay must be public HTTPS.

## 4. Server Startup

File:

```text
backend/cmd/server/main.go
```

Startup flow:

1. `config.Load()` reads `.env` and environment variables.
2. The backend checks that `DATABASE_URL`, `SUPABASE_URL`, and `SUPABASE_ANON_KEY` exist.
3. Postgres pool is created.
4. Supabase auth client is created.
5. AddisPay client is created only if `ADDISPAY_API_KEY` is set and is not the placeholder.
6. Router receives `AddisPayClient`, webhook secret, and return URL config.

Important code behavior:

```text
If ADDISPAY_API_KEY is missing, the backend still starts, but payment features are disabled.
```

If a player tries to create a paid checkout while the client is missing, the service returns:

```text
payment gateway is not configured
```

This is intentional. It keeps the app usable for non-payment features while preventing fake paid payments.

## 5. Routes And Authentication

File:

```text
backend/internal/router/router.go
```

Authenticated payment routes:

```text
POST /api/v1/payments/create
POST /api/v1/payments/return
GET  /api/v1/payments/status/{id}
POST /api/v1/payments/{id}/mark-refunded
```

Public AddisPay webhook route:

```text
POST /api/v1/payments/webhook
```

This route is public because AddisPay cannot send a Supabase Bearer token. It is still protected by HMAC signature verification.

Important detail:

```text
The webhook is outside the auth middleware group, but it is still mounted under the /api/v1 prefix.
```

So the production webhook URL should look like:

```text
https://your-backend-domain.com/api/v1/payments/webhook
```

For local tunnel testing:

```text
https://your-ngrok-domain.ngrok-free.app/api/v1/payments/webhook
```

## 6. Database Payment Model

Main model:

```text
backend/internal/repository/models.go
```

Payment fields:

| Field | Meaning |
| --- | --- |
| `id` | Local payment UUID. This is also sent to AddisPay as `tx_ref` and `nonce`. |
| `user_id` | Player who is paying. |
| `tournament_id` | Tournament being paid for. |
| `amount` | Entry fee amount stored locally. |
| `currency` | Currency. Current payment flow uses `ETB`. |
| `status` | Local status: `pending`, `paid`, `failed`, or `refunded`. |
| `addispay_ref` | Reference used to find the payment from a webhook. In this code it equals the local payment ID. |
| `payment_url` | Hosted checkout URL returned by AddisPay. |
| `provider_status` | Raw status from AddisPay, such as `paid`, `success`, `failed`, etc. |
| `provider_payment_id` | AddisPay/provider-side payment ID if sent in webhook. |
| `verified_at` | Time the backend accepted a paid webhook. |
| `webhook_received_at` | Time any valid-format webhook was processed for this payment. |
| `failure_reason` | Why payment was marked failed or rejected. |
| `metadata` | JSON copy of important webhook fields for reconciliation. |
| `refund_status` | Manual refund tracking: `none`, `pending`, `refunded`, `failed`, or `not_required`. |
| `refund_reason` | Reason refund is needed. |
| `refund_requested_at` | Time refund was requested. |
| `refunded_at` | Time organizer marked refund paid. |
| `refunded_by` | Organizer/admin profile ID that marked refund paid. |

Why the backend stores both local and provider fields:

```text
Local fields decide app behavior. Provider fields help admin reconciliation when AddisPay says something different or a webhook arrives late.
```

## 7. Registration Before Payment

File:

```text
backend/internal/modules/registration/service.go
```

The player must register before payment creation.

For a free tournament:

```text
registrations.payment_status = paid
```

For a paid tournament:

```text
registrations.payment_status = pending
```

The payment service checks for this registration before creating a payment. That means payment is tied to an actual player/tournament registration, not just any random request.

Why this matters:

- prevents payment creation for tournaments the user did not join
- separates "reserved/pending participant" from "paid participant"
- allows the tournament to count only paid users when starting

The tournament start logic checks paid registrations, not just all registrations.

## 8. Create Payment Endpoint

Endpoint:

```text
POST /api/v1/payments/create
```

Auth:

```text
Required Supabase Bearer token
```

Request body:

```json
{
  "tournament_id": "tournament-id"
}
```

Handler:

```text
backend/internal/modules/payment/handler.go
CreatePayment()
```

Handler responsibility:

1. Read authenticated `userID` from request context.
2. Decode `tournament_id` from JSON body.
3. Call `service.CreatePayment(ctx, userID, tournamentID)`.
4. Return `201 Created` with payment data and checkout URL.
5. Return `400 Bad Request` for business errors.

Service:

```text
backend/internal/modules/payment/service.go
CreatePayment()
```

Service logic step by step:

1. Load tournament by ID.
2. Reject if tournament does not exist.
3. Reject if `entry_fee <= 0` because free tournaments do not need AddisPay.
4. Load existing registration for this user and tournament.
5. Reject if the user has not registered yet.
6. Reject if registration is already paid.
7. Reject if AddisPay client is not configured.
8. Load user profile.
9. Insert a local payment row with:

```text
status = pending
currency = ETB
amount = tournament.entry_fee
```

10. Build an AddisPay create-order request.
11. Call AddisPay.
12. If AddisPay fails, mark local payment failed and store the failure reason.
13. Extract hosted checkout URL from AddisPay response.
14. If AddisPay response has no usable URL, mark local payment failed.
15. Store `addispay_ref` and `payment_url` on the local payment.
16. Return the payment and `payment_url` to the caller.

Success response shape:

```json
{
  "payment": {
    "id": "local-payment-id",
    "user_id": "player-id",
    "tournament_id": "tournament-id",
    "amount": 100,
    "currency": "ETB",
    "status": "pending",
    "addispay_ref": "local-payment-id",
    "payment_url": "https://checkout..."
  },
  "payment_url": "https://checkout..."
}
```

Important design choice:

```text
The backend creates the local payment before calling AddisPay.
```

Why:

- we need a local ID to use as `tx_ref`
- we need an audit record even when AddisPay fails
- webhook reconciliation becomes direct because AddisPay returns the same reference

## 9. AddisPay Create-Order Request

File:

```text
backend/pkg/addispay/client.go
CreatePayment()
```

Request URL:

```text
POST {ADDISPAY_BASE_URL}/checkout-api/v1/create-order
```

Headers:

```text
Content-Type: application/json
Accept: application/json
Auth: <ADDISPAY_API_KEY>
```

Important fields sent to AddisPay:

```json
{
  "message": "Bona tournament entry payment",
  "data": {
    "redirect_url": "frontend return URL",
    "cancel_url": "frontend return URL",
    "success_url": "frontend return URL",
    "error_url": "frontend return URL",
    "order_reason": "Tournament entry: <title>",
    "currency": "ETB",
    "email": "player email",
    "first_name": "player first name",
    "last_name": "player last name",
    "nonce": "local payment id",
    "phone_number": "",
    "session_expired": "5000",
    "total_amount": "entry fee",
    "tx_ref": "local payment id",
    "order_detail": {
      "amount": 100,
      "description": "Tournament entry: <title>"
    }
  }
}
```

Most important field:

```text
tx_ref = local payment id
```

The code also sets:

```text
nonce = local payment id
```

Why:

```text
When AddisPay sends a webhook, the backend can find the local payment directly by reference.
```

Reference mapping:

```text
AddisPay tx_ref/reference -> payments.addispay_ref -> payments.id
```

Customer name handling:

```text
customerName(profile)
```

This function tries to produce safe first and last names from:

1. display name
2. username
3. email local-part
4. fallback `Bona Player`

It strips numeric and short invalid tokens because payment gateways often reject bad names.

## 10. AddisPay Response Parsing

File:

```text
backend/pkg/addispay/client.go
PaymentResponse
HostedCheckoutURL()
UnmarshalJSON()
```

The client accepts multiple possible AddisPay response shapes.

Direct:

```json
{
  "checkout_url": "https://checkout.addispay.et/pay",
  "uuid": "order-uuid"
}
```

Wrapped:

```json
{
  "data": {
    "checkout_url": "https://checkout.addispay.et/pay",
    "uuid": "order-uuid"
  }
}
```

Hosted URL:

```json
{
  "data": {
    "hosted_checkout_url": "https://checkout.addispay.et/pay/order-uuid"
  }
}
```

The backend supports:

- `hosted_checkout_url`
- `payment_url`
- `checkout_url`
- `checkout_url + uuid`

Why:

```text
Payment providers sometimes return slightly different response field names between sandbox, production, or API versions. The parser is defensive so checkout still works if the URL is present.
```

If no valid checkout URL exists, the backend marks the payment failed with:

```text
AddisPay did not return a checkout URL
```

## 11. Return URL Token

The backend sends AddisPay return URLs that look like this:

```text
https://frontend-domain.com/payments/return?payment_id=<id>&tournament_id=<id>&status=success&token=<hmac>
```

Function:

```text
paymentReturnURL()
paymentReturnToken()
```

Token input:

```text
payment.id|payment.user_id|payment.tournament_id|payment.amount|status
```

Algorithm:

```text
HMAC-SHA256(ADDISPAY_WEBHOOK_SECRET, token input)
```

Why:

```text
The return URL is visible in the browser. The token prevents someone from editing payment_id/status in the URL and tricking the backend return endpoint.
```

Very important:

```text
The return token does not prove the user paid.
```

It only proves the return URL was generated by our backend. Payment is still confirmed only by webhook.

## 12. Confirm Return Endpoint

Endpoint:

```text
POST /api/v1/payments/return
```

Auth:

```text
Required Supabase Bearer token
```

Request body:

```json
{
  "payment_id": "local-payment-id",
  "status": "success",
  "token": "hmac-token-from-return-url"
}
```

Handler:

```text
ConfirmReturn()
```

Service:

```text
ConfirmReturn()
```

Service logic:

1. Reject if `payment_id` is missing.
2. Load payment by ID.
3. Reject if payment does not belong to authenticated user.
4. Validate return token using `ADDISPAY_WEBHOOK_SECRET`.
5. Return the current stored payment status.

Response:

```json
{
  "payment": {
    "id": "local-payment-id",
    "status": "pending"
  },
  "tournament_id": "tournament-id",
  "payment_status": "pending"
}
```

Important:

```text
This endpoint does not update payment.status to paid.
```

Why:

```text
Browser redirects can be forged, replayed, or opened before AddisPay fully settles the transaction. The backend waits for the signed server-to-server webhook.
```

## 13. Webhook Endpoint

Endpoint:

```text
POST /api/v1/payments/webhook
```

Auth:

```text
No Supabase auth. Protected by HMAC signature.
```

Handler:

```text
backend/internal/modules/payment/handler.go
Webhook()
```

Expected payload shape:

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

| Meaning | Accepted fields |
| --- | --- |
| Reference | `reference`, `tx_ref`, `txRef`, `transaction_reference`, `transactionReference`, `nonce`, `order_id`, `orderId` |
| Status | `status`, `payment_status`, `paymentStatus`, `state` |
| Amount | `amount`, `total_amount`, `totalAmount` |
| Provider payment ID | `payment_id`, `paymentId`, `uuid`, `id` |

It also accepts fields inside:

```json
{
  "data": {
    "tx_ref": "...",
    "payment_status": "..."
  }
}
```

Why:

```text
This makes webhook parsing tolerant of wrapped payloads and small naming differences.
```

## 14. Webhook Signature Verification

File:

```text
backend/pkg/addispay/client.go
VerifyWebhookSignature()
```

Current signed string:

```text
reference|status|amount
```

Algorithm:

```text
HMAC-SHA256(ADDISPAY_WEBHOOK_SECRET, reference + "|" + status + "|" + amount)
```

Signature format:

```text
hex encoded
```

The handler accepts the signature from:

1. JSON field `signature`
2. `X-AddisPay-Signature`
3. `X-Addispay-Signature`
4. `X-Signature`
5. `Signature`

If verification fails:

```text
HTTP 401 Invalid signature
```

Production note:

```text
Before going live, confirm AddisPay's exact webhook signature base string with their current documentation or integration team. If AddisPay signs a different string or the raw body, update only VerifyWebhookSignature() and the related tests.
```

## 15. Webhook Business Logic

Service:

```text
backend/internal/modules/payment/service.go
HandleWebhook()
```

Webhook handling step by step:

1. Reject if `payload.Reference` is missing.
2. Find payment by `addispay_ref`.
3. If not found, try finding payment directly by local payment `id`.
4. Compare webhook amount with local `payments.amount`.
5. Reject and record failure if amount does not match.
6. Compare webhook currency with local `payments.currency`.
7. Reject and record failure if currency does not match.
8. Convert provider status to local status.
9. Reject and record failure if status is unsupported.
10. Complete the payment from webhook.

Status mapping:

| AddisPay/provider status | Local status |
| --- | --- |
| `completed` | `paid` |
| `complete` | `paid` |
| `success` | `paid` |
| `successful` | `paid` |
| `paid` | `paid` |
| `approved` | `paid` |
| `captured` | `paid` |
| `failed` | `failed` |
| `failure` | `failed` |
| `declined` | `failed` |
| `cancelled` | `failed` |
| `canceled` | `failed` |
| `expired` | `failed` |
| `error` | `failed` |

Unsupported statuses are rejected because the backend should not guess payment meaning.

## 16. Atomic Payment Completion

Function:

```text
completePaymentFromWebhook()
```

This function uses a database transaction.

When status is `paid`, the transaction does:

```text
payments.status = paid
payments.provider_status = <raw AddisPay status>
payments.provider_payment_id = <provider ID if present>
payments.verified_at = NOW()
payments.webhook_received_at = NOW()
registrations.payment_status = paid
```

Why a transaction:

```text
Payment and registration must change together. We do not want payment.status = paid while registration.payment_status is still pending, or the opposite.
```

If updating registration fails, the payment update rolls back.

After commit:

- paid notification is sent once when payment first becomes paid
- failed notification is sent when a pending payment becomes failed

Notifications are intentionally after commit because payment state must be saved first.

## 17. Idempotency And Duplicate Webhooks

Repository:

```text
backend/internal/repository/payment.go
CompletePaymentFromWebhook()
```

Important SQL behavior:

```sql
status = CASE
    WHEN status = 'paid' THEN status
    ELSE $2
END
```

Meaning:

```text
If the payment is already paid, a later webhook cannot downgrade it to failed.
```

Why this matters:

- payment providers can retry webhooks
- webhooks can arrive out of order
- a user may pay successfully and then AddisPay may send a late duplicate event
- the backend should treat already-paid as final for entry confirmation

Also:

```sql
verified_at = CASE
    WHEN $2 = 'paid' THEN COALESCE(verified_at, NOW())
    ELSE verified_at
END
```

Meaning:

```text
The first successful verification time is preserved.
```

## 18. Mismatch Handling

Function:

```text
recordWebhookMismatch()
```

Mismatch examples:

- amount mismatch
- currency mismatch
- unsupported status

When mismatch happens, backend records:

```text
payments.status = failed
payments.provider_status = webhook status
payments.provider_payment_id = webhook provider ID if present
payments.failure_reason = reason
payments.metadata = webhook summary JSON
payments.webhook_received_at = NOW()
```

Why:

```text
Even rejected webhooks should leave a trail. Admins need to know whether the webhook arrived and why it was not accepted.
```

Important safety:

```text
The SQL does not downgrade an already-paid payment.
```

So if a payment is already paid, mismatch recording cannot change `status` from `paid` to `failed`.

## 19. Payment Status State Machine

Local payment statuses:

```text
pending
paid
failed
refunded
```

Normal payment transitions:

```text
pending -> paid
pending -> failed
paid -> refunded
```

Blocked/avoided transitions:

```text
paid -> failed
failed -> paid by browser return
pending -> paid without webhook
```

Registration statuses:

```text
pending
paid
failed
refunded
```

For paid tournaments:

```text
registration starts pending
registration becomes paid only after payment webhook is verified
```

## 20. Refund Tracking

This backend has manual refund tracking, not AddisPay automatic refunds.

When a paid tournament is cancelled:

```text
payments.refund_status = pending
payments.refund_reason = Tournament cancelled by organizer
payments.refund_requested_at = NOW()
```

Route:

```text
POST /api/v1/payments/{id}/mark-refunded
```

Rules:

1. User must be authenticated.
2. User must be the tournament organizer.
3. Tournament must be cancelled.
4. Payment must have `refund_status = pending`.
5. Backend marks:

```text
payments.status = refunded
payments.refund_status = refunded
payments.refunded_at = NOW()
payments.refunded_by = organizerID
registrations.payment_status = refunded
```

Important presentation point:

```text
Refund money movement is not performed through AddisPay in this code. This is only backend tracking after the organizer handles the refund externally.
```

## 21. Admin And Reconciliation

Admin route:

```text
GET /api/v1/admin/payments
```

Handler:

```text
backend/internal/modules/admin/handler.go
ListAllPayments()
```

Admin payment records expose:

- local payment ID
- user ID
- tournament ID
- amount and currency
- local status
- AddisPay reference
- provider status
- provider payment ID
- `webhook_received_at`
- `verified_at`
- `failure_reason`
- refund fields

This supports these operational questions:

| Question | Field to check |
| --- | --- |
| Did checkout creation happen? | `payment_url`, `addispay_ref` |
| Did AddisPay webhook arrive? | `webhook_received_at` |
| Did backend accept the payment as paid? | `status`, `verified_at` |
| What did AddisPay call the payment? | `provider_status`, `provider_payment_id` |
| Why was it rejected? | `failure_reason`, `metadata` |
| Is refund pending? | `refund_status`, `refund_requested_at` |

## 22. Local End-To-End Backend Test

Start backend:

```bash
cd backend
go run ./cmd/server
```

Backend health:

```text
GET http://localhost:8080/health
```

Expose backend to AddisPay:

```bash
ngrok http 8080
```

If ngrok gives:

```text
https://abc123.ngrok-free.app
```

Configure AddisPay webhook URL as:

```text
https://abc123.ngrok-free.app/api/v1/payments/webhook
```

Expected backend flow:

1. Player registers for a paid tournament.
2. Backend registration row is `pending`.
3. `POST /api/v1/payments/create` creates a local `pending` payment.
4. Backend calls AddisPay create-order.
5. Backend stores `payment_url` and `addispay_ref`.
6. Player pays on AddisPay Hosted Checkout.
7. AddisPay sends webhook to backend.
8. Backend verifies signature.
9. Backend validates amount and currency.
10. Backend updates payment and registration in one transaction.
11. Admin can see `status = paid`, `verified_at`, and `webhook_received_at`.

## 23. Manual Webhook Test

For a local manual test, first get a real local payment row with:

```text
payments.id = <payment-id>
amount = <amount>
currency = ETB
```

Generate signature using the same string the backend verifies:

```bash
printf '%s' '<payment-id>|paid|<amount>' \
  | openssl dgst -sha256 -hmac '<ADDISPAY_WEBHOOK_SECRET>' -hex
```

The output looks like:

```text
SHA2-256(stdin)= <signature>
```

Send webhook:

```bash
curl -X POST http://localhost:8080/api/v1/payments/webhook \
  -H 'Content-Type: application/json' \
  -d '{
    "reference": "<payment-id>",
    "status": "paid",
    "amount": <amount>,
    "currency": "ETB",
    "signature": "<signature>",
    "payment_id": "manual-test-provider-id"
  }'
```

Expected response:

```json
{
  "status": "received"
}
```

Expected database result:

```text
payments.status = paid
payments.verified_at is not null
payments.webhook_received_at is not null
registrations.payment_status = paid
```

## 24. Automated Tests

Payment-related tests:

```text
backend/pkg/addispay/client_test.go
backend/internal/modules/payment/service_test.go
backend/internal/repository/payment_test.go
backend/internal/router/router_test.go
```

What they cover:

- hosted checkout URL construction
- wrapped AddisPay response parsing
- direct hosted checkout URL parsing
- webhook payload alias parsing
- webhook HMAC signature accept/reject behavior
- signed payment return token validation
- payment return URL generation
- provider status mapping to local `paid` / `failed`
- webhook metadata capture for reconciliation
- SQL idempotency guard that prevents `paid` payments from being downgraded
- customer name sanitization
- public webhook route mounting at `/api/v1/payments/webhook`
- public webhook rejection before database logic when signature is invalid

Run:

```bash
cd backend
go test ./...
```

If Go cache permissions are restricted:

```bash
cd backend
GOCACHE=/tmp/bona-go-build-cache go test ./...
```

## 25. Common Backend Failures

Payment gateway is not configured:

```text
ADDISPAY_API_KEY is missing or still set to placeholder.
```

Checkout creation fails:

```text
Check ADDISPAY_API_KEY, ADDISPAY_BASE_URL, AddisPay response body in backend error, and payment.failure_reason.
```

Payment stays pending:

```text
Webhook did not arrive, webhook URL is wrong, tunnel is down, secret mismatch caused 401, or AddisPay has not sent the final event yet.
```

Webhook returns 401:

```text
Signature is missing or invalid. Check ADDISPAY_WEBHOOK_SECRET and the signature base string.
```

Webhook returns amount mismatch:

```text
payload.amount does not equal payments.amount.
```

Webhook returns currency mismatch:

```text
payload.currency is not ETB or does not match payments.currency.
```

Webhook returns unsupported status:

```text
AddisPay sent a status not listed in gatewayPaymentStatus().
```

Return endpoint says invalid token:

```text
ADDISPAY_WEBHOOK_SECRET changed after payment creation, token was edited, status value changed, or secret is empty.
```

## 26. Production Readiness Checklist

Before presenting as production-ready, verify these:

- `ADDISPAY_API_KEY` is configured in backend secret storage.
- `ADDISPAY_WEBHOOK_SECRET` is strong, private, and matches AddisPay configuration.
- `ADDISPAY_BASE_URL` is the correct live AddisPay URL for production.
- Frontend return URLs are HTTPS production URLs.
- AddisPay webhook URL is `https://backend-domain/api/v1/payments/webhook`.
- Database migrations through `000011_payment_refunds.up.sql` are applied.
- Backend logs are collected for webhook failures.
- Admin users can access `/api/v1/admin/payments`.
- AddisPay signature format has been confirmed with AddisPay's current documentation/team.
- A real UAT payment has been tested end to end.
- Duplicate webhook test confirms paid payments are not downgraded.
- Amount mismatch test records `failure_reason`.
- Currency mismatch test records `failure_reason`.
- Pending payment operational process exists: check `webhook_received_at`, AddisPay dashboard, and backend logs.

## 27. How To Explain The Integration In An Interview

Short answer:

```text
We integrated AddisPay Hosted Checkout for tournament entry fees. The backend creates a pending local payment, sends AddisPay a create-order request using our payment ID as tx_ref, returns the hosted checkout URL, and finalizes the payment only when AddisPay sends a valid signed webhook. The webhook is verified using HMAC, then amount and currency are checked, and payment plus registration are updated in one database transaction.
```

If asked why not trust the return URL:

```text
The return URL is browser-controlled, so it can be edited or replayed. It is useful for user experience, but not for financial truth. The server-to-server signed webhook is the authority.
```

If asked how reconciliation works:

```text
We store the local payment ID as AddisPay tx_ref/addispay_ref. When webhook comes back, we use that reference to find the payment. We also store provider_status, provider_payment_id, webhook_received_at, verified_at, failure_reason, and metadata so admins can investigate payment state.
```

If asked about security:

```text
Paid status requires a valid HMAC webhook signature, matching amount, matching currency, and a supported provider status. The webhook route is public because AddisPay cannot authenticate with Supabase, but it is protected by signature verification.
```

If asked about consistency:

```text
The backend updates payment.status and registrations.payment_status inside one transaction. If either update fails, the transaction rolls back.
```

If asked about duplicate webhooks:

```text
The SQL is idempotent. Once a payment is paid, later webhooks cannot downgrade it to failed, and verified_at keeps the first verification time.
```

If asked about refunds:

```text
The code tracks manual refunds after tournament cancellation. It does not call AddisPay to refund money automatically.
```

## 28. Deep Technical Q&A

Question: Why do we create the payment before AddisPay?

Answer:

```text
Because we need a stable local payment ID to send as tx_ref/nonce. That ID is later used to match AddisPay webhooks back to the exact local payment.
```

Question: What happens if AddisPay create-order fails?

Answer:

```text
The backend marks the local payment failed and stores the error in failure_reason. The user can retry payment, which creates a new payment attempt.
```

Question: Can a user pay without registering?

Answer:

```text
No. CreatePayment checks GetRegistrationByUserAndTournament first. If there is no registration, it returns "you must register before paying".
```

Question: Can someone pay for another user's payment?

Answer:

```text
The create endpoint uses the authenticated user ID. The return endpoint also verifies the payment belongs to the authenticated user. The webhook itself is provider-to-server and is verified by HMAC and payment reference.
```

Question: What if webhook amount is smaller than expected?

Answer:

```text
The backend rejects it, records failure_reason = "payment amount mismatch", and does not mark registration paid.
```

Question: What if AddisPay sends `success` instead of `paid`?

Answer:

```text
gatewayPaymentStatus() maps success-like statuses such as success, successful, completed, approved, and captured to local paid.
```

Question: What if AddisPay sends an unknown status?

Answer:

```text
The backend rejects it as unsupported instead of guessing. This protects against accidentally accepting ambiguous states.
```

Question: Why store provider_status if status already exists?

Answer:

```text
status is our app's normalized state. provider_status is the raw AddisPay state for audit and debugging.
```

Question: Why is the webhook route public?

Answer:

```text
External payment providers cannot send our users' Supabase tokens. So the route must be public, but protected by cryptographic signature verification.
```

Question: What must be confirmed with AddisPay before production?

Answer:

```text
The live base URL, the API key format, the webhook URL configured in AddisPay, and the exact signature algorithm/base string.
```

## 29. One-Minute Backend Walkthrough

Use this if you need to explain the full backend flow quickly:

```text
The user first joins a paid tournament, so registration is created as pending. Then POST /api/v1/payments/create creates a pending payment row and sends AddisPay a create-order request. The local payment ID is used as tx_ref and nonce, which makes webhook reconciliation simple. AddisPay returns a hosted checkout URL, and the backend stores it on the payment row.

When the user returns from AddisPay, POST /api/v1/payments/return only validates that our backend created that return URL and then shows the current stored status. It never marks the payment paid.

The real confirmation is POST /api/v1/payments/webhook. The backend parses the payload, verifies the HMAC signature, checks amount and currency, maps provider status to local status, and then updates payments and registrations in one transaction. Admin reconciliation fields store provider status, provider payment ID, webhook time, verified time, failure reason, and metadata.
```
