# Payments and AddisPay Integration

This file explains what payment functionality is integrated, where it lives in the codebase, and where the user actually interacts with AddisPay.

## Current Payment Integration

The project has AddisPay Hosted Checkout integrated for tournament entry fees.

Integrated:

- Players can register for paid tournaments.
- The frontend asks the backend to create a payment.
- The backend creates a local `payments` row.
- The backend sends a hosted checkout order to AddisPay.
- The frontend redirects the player to AddisPay's hosted checkout page.
- AddisPay redirects the player back to the app after payment, cancellation, or failure.
- The backend can update payment and registration status from the return URL.
- The backend can also receive AddisPay webhooks and update payment and registration status.
- Paid registrations count toward tournament paid spots and prize pool calculations.

Not integrated:

- No card/mobile-money form is built inside this app. Payment happens on AddisPay's hosted page.
- No refund flow is implemented.
- No saved payment methods are implemented.
- No server-side "verify transaction by reference" API call is implemented after return.
- Winner payouts are not paid through AddisPay. Payouts are tracked manually by the organizer in the app.

## Where The User Interacts With AddisPay

The user touches AddisPay from the tournament details page.

Frontend file:

```text
frontend/src/pages/TournamentDetails.tsx
```

Important functions:

- `handleRegister()`
- `handlePay()`

The user sees one of these buttons in the registration card:

- `Register for {entry_fee} {currency}`
- `Pay {entry_fee} {currency}`

When clicked, the frontend calls:

```text
POST /api/v1/payments/create
```

If the backend returns `payment_url`, the frontend does:

```ts
window.location.href = pay.payment_url
```

That is the exact handoff to AddisPay. The browser leaves Bona and opens the AddisPay hosted checkout URL.

## End-To-End Flow

1. Player opens a tournament page.
2. Player clicks `Register for ...` on a paid tournament.
3. Frontend calls:

```text
POST /api/v1/tournaments/{id}/join
```

4. Backend creates a registration with `payment_status = pending`.
5. Frontend calls:

```text
POST /api/v1/payments/create
```

with:

```json
{
  "tournament_id": "tournament-id"
}
```

6. Backend creates a local `payments` row with `status = pending`.
7. Backend calls AddisPay:

```text
POST {ADDISPAY_BASE_URL}/checkout-api/v1/create-order
```

8. AddisPay returns checkout information.
9. Backend returns the hosted checkout URL to the frontend:

```json
{
  "payment": {
    "id": "local-payment-id",
    "status": "pending"
  },
  "payment_url": "https://addispay-hosted-checkout-url"
}
```

10. Frontend redirects the browser to `payment_url`.
11. Player completes, cancels, or fails payment on AddisPay.
12. AddisPay redirects the player back to:

```text
/payments/return?payment_id=...&tournament_id=...&status=...&token=...
```

13. Frontend page `PaymentReturn.tsx` calls:

```text
POST /api/v1/payments/return
```

14. Backend validates the return token, then marks the payment `paid` or `failed`.
15. Separately, AddisPay can also call the webhook:

```text
POST /api/v1/payments/webhook
```

16. Webhook validation can also mark the payment `paid` or `failed`.
17. When a payment is paid, backend sets:

```text
payments.status = paid
registrations.payment_status = paid
```

## Main Files

### Frontend

```text
frontend/src/pages/TournamentDetails.tsx
```

Starts the payment flow. It calls `/api/v1/payments/create` and redirects to `payment_url`.

```text
frontend/src/pages/PaymentReturn.tsx
```

Handles the user returning from AddisPay. It reads `payment_id`, `tournament_id`, `status`, and `token` from the URL, then calls `/api/v1/payments/return`.

```text
frontend/src/App.tsx
```

Registers the frontend return route:

```text
/payments/return
```

### Backend

```text
backend/cmd/server/main.go
```

Initializes the AddisPay client when `ADDISPAY_API_KEY` is set and is not the placeholder value.

```text
backend/internal/config/config.go
```

Loads AddisPay environment variables:

- `ADDISPAY_API_KEY`
- `ADDISPAY_BASE_URL`
- `ADDISPAY_WEBHOOK_SECRET`
- `ADDISPAY_REDIRECT_URL`
- `ADDISPAY_CANCEL_URL`
- `ADDISPAY_SUCCESS_URL`
- `ADDISPAY_ERROR_URL`

```text
backend/internal/router/router.go
```

Registers payment routes.

Authenticated routes:

```text
POST /api/v1/payments/create
POST /api/v1/payments/return
GET  /api/v1/payments/status/{id}
```

Public route for AddisPay:

```text
POST /api/v1/payments/webhook
```

```text
backend/internal/modules/payment/handler.go
```

HTTP layer for creating payments, confirming return, checking status, and receiving webhooks.

```text
backend/internal/modules/payment/service.go
```

Main payment business logic:

- Validates tournament and registration.
- Creates local payment row.
- Builds AddisPay checkout request.
- Saves `addispay_ref` and `payment_url`.
- Confirms return token.
- Handles webhook payload.
- Marks payment paid or failed.
- Updates registration payment status.
- Emits payment notifications.

```text
backend/pkg/addispay/client.go
```

Low-level AddisPay client:

- Sends checkout order request to AddisPay.
- Adds the AddisPay `Auth` header.
- Parses AddisPay checkout responses.
- Builds final hosted checkout URL.
- Parses webhook payloads.
- Verifies webhook signatures.

```text
backend/internal/repository/payment.go
```

Database access for `payments`.

```text
backend/migrations/000001_create_tables.up.sql
```

Defines the `payments` table and `registrations.payment_status`.

## AddisPay Request

The backend sends this request from:

```text
backend/pkg/addispay/client.go
```

Endpoint:

```text
POST {ADDISPAY_BASE_URL}/checkout-api/v1/create-order
```

Headers:

```text
Content-Type: application/json
Accept: application/json
Auth: {ADDISPAY_API_KEY}
```

Request body shape:

```json
{
  "data": {
    "redirect_url": "http://localhost:5173/payments/return?payment_id=...&tournament_id=...&status=success&token=...",
    "cancel_url": "http://localhost:5173/payments/return?payment_id=...&tournament_id=...&status=cancelled&token=...",
    "success_url": "http://localhost:5173/payments/return?payment_id=...&tournament_id=...&status=success&token=...",
    "error_url": "http://localhost:5173/payments/return?payment_id=...&tournament_id=...&status=failed&token=...",
    "order_reason": "Tournament entry: Tournament Title",
    "currency": "ETB",
    "email": "player@example.com",
    "first_name": "Player",
    "last_name": "Name",
    "nonce": "local-payment-id",
    "order_detail": {
      "amount": 100,
      "description": "Tournament entry: Tournament Title"
    },
    "phone_number": "",
    "session_expired": "5000",
    "total_amount": "100",
    "tx_ref": "local-payment-id"
  },
  "message": "Bona tournament entry payment"
}
```

Important mapping:

- `tx_ref` is the local `payments.id`.
- `nonce` is also the local `payments.id`.
- `addispay_ref` saved in the database is currently the local `payments.id`.
- `total_amount` and `order_detail.amount` come from `tournaments.entry_fee`.
- Currency is currently set to `ETB`.

## AddisPay Response Handling

The code accepts several response shapes from AddisPay.

It supports top-level fields:

```json
{
  "checkout_url": "https://checkout.addispay.et/pay",
  "uuid": "order-uuid"
}
```

It also supports wrapped `data` fields:

```json
{
  "status": "success",
  "data": {
    "checkout_url": "https://checkout.addispay.et/pay",
    "uuid": "order-uuid"
  }
}
```

It also supports direct hosted checkout URLs:

```json
{
  "data": {
    "hosted_checkout_url": "https://checkout.addispay.et/pay/order-uuid"
  }
}
```

The final URL is chosen by `PaymentResponse.HostedCheckoutURL()` in:

```text
backend/pkg/addispay/client.go
```

## Return URL Handling

The return page is:

```text
frontend/src/pages/PaymentReturn.tsx
```

Registered at:

```text
/payments/return
```

The backend creates return URLs in:

```text
backend/internal/modules/payment/service.go
```

Function:

```text
paymentReturnURL(...)
```

The URL includes:

- `payment_id`
- `tournament_id`
- `status`
- `token`

The token is HMAC-SHA256 using `ADDISPAY_WEBHOOK_SECRET`.

Signed values:

```text
payment_id|user_id|tournament_id|amount|status
```

The frontend posts those values to:

```text
POST /api/v1/payments/return
```

If the token is valid and the status maps to success, the backend marks the payment as paid.

Success statuses accepted by code:

```text
completed, complete, success, successful, paid, approved, captured
```

Failure statuses accepted by code:

```text
failed, failure, declined, cancelled, canceled, expired, error
```

## Webhook Handling

AddisPay webhook URL:

```text
https://your-backend-domain.com/api/v1/payments/webhook
```

For local development with backend on port `8080`, expose it with a tunnel and give AddisPay:

```text
https://your-tunnel-domain/api/v1/payments/webhook
```

Webhook route:

```text
backend/internal/router/router.go
```

Webhook handler:

```text
backend/internal/modules/payment/handler.go
```

Webhook service logic:

```text
backend/internal/modules/payment/service.go
```

Webhook payload parser and signature verifier:

```text
backend/pkg/addispay/client.go
```

Expected payload can be direct:

```json
{
  "reference": "local-payment-id",
  "status": "success",
  "amount": 100,
  "currency": "ETB",
  "signature": "hex_hmac_signature",
  "payment_id": "gateway-payment-id"
}
```

Or wrapped:

```json
{
  "data": {
    "tx_ref": "local-payment-id",
    "payment_status": "PAID",
    "total_amount": "100",
    "currency": "ETB",
    "signature": "hex_hmac_signature",
    "uuid": "gateway-payment-id"
  }
}
```

The parser accepts these reference keys:

```text
reference, tx_ref, txRef, transaction_reference, transactionReference, nonce, order_id, orderId
```

The parser accepts these status keys:

```text
status, payment_status, paymentStatus, state
```

The parser accepts these amount keys:

```text
amount, total_amount, totalAmount
```

Signature can be in the JSON body or one of these headers:

```text
X-AddisPay-Signature
X-Addispay-Signature
X-Signature
Signature
```

Current webhook signature check:

```text
HMAC-SHA256(reference|status|amount, ADDISPAY_WEBHOOK_SECRET)
```

If AddisPay signs a different string in your account or docs, update:

```text
backend/pkg/addispay/client.go
```

Function:

```text
VerifyWebhookSignature(...)
```

## Database State

Payment rows are stored in:

```text
payments
```

Important columns:

- `id`
- `user_id`
- `tournament_id`
- `amount`
- `currency`
- `status`
- `addispay_ref`
- `payment_url`
- `metadata`
- `created_at`
- `updated_at`

Registration payment status is stored in:

```text
registrations.payment_status
```

Statuses used:

```text
pending, paid, failed, refunded
```

When a payment succeeds:

```text
payments.status = paid
registrations.payment_status = paid
```

When a payment fails:

```text
payments.status = failed
```

The current code does not update `registrations.payment_status` to `failed` when payment fails.

## Environment Variables

Development/UAT:

```env
ADDISPAY_API_KEY=your_addispay_api_key
ADDISPAY_BASE_URL=https://uat.api.addispay.et
ADDISPAY_WEBHOOK_SECRET=your_addispay_webhook_secret
ADDISPAY_REDIRECT_URL=http://localhost:5173
ADDISPAY_CANCEL_URL=http://localhost:5173
ADDISPAY_SUCCESS_URL=http://localhost:5173
ADDISPAY_ERROR_URL=http://localhost:5173
```

Production:

```env
ADDISPAY_API_KEY=your_production_addispay_api_key
ADDISPAY_BASE_URL=https://api.addispay.et
ADDISPAY_WEBHOOK_SECRET=your_production_webhook_secret
ADDISPAY_REDIRECT_URL=https://your-frontend-domain.com
ADDISPAY_CANCEL_URL=https://your-frontend-domain.com
ADDISPAY_SUCCESS_URL=https://your-frontend-domain.com
ADDISPAY_ERROR_URL=https://your-frontend-domain.com
```

Backend startup behavior:

- If `ADDISPAY_API_KEY` is empty, payment gateway features are disabled.
- If `ADDISPAY_API_KEY` is `your_addispay_api_key`, payment gateway features are disabled.
- Otherwise `backend/cmd/server/main.go` creates an AddisPay client.

## Important Security Note

The current implementation can mark payment as paid from either:

- A valid AddisPay webhook.
- A valid signed return URL handled by `/api/v1/payments/return`.

The return URL token prevents simple user tampering because it is signed by the backend using `ADDISPAY_WEBHOOK_SECRET`. However, the return path does not call AddisPay server-to-server to verify the transaction status.

For stricter production payment confirmation, use webhook-only confirmation or add a backend verification call to AddisPay before marking a payment as paid from the return page.

## Quick Manual Test

1. Start backend:

```bash
cd backend
go run ./cmd/server
```

2. Start frontend:

```bash
cd frontend
npm run dev
```

3. Log in.
4. Create or open a paid tournament.
5. Click `Register for ...`.
6. Confirm the browser redirects to AddisPay.
7. Complete or cancel the payment in AddisPay UAT.
8. Confirm the browser returns to `/payments/return`.
9. Confirm `payments.status` changes.
10. Confirm `registrations.payment_status` becomes `paid` for successful payments.
11. Confirm paid players count toward tournament paid spots.

## Files To Check When Something Breaks

If clicking pay does nothing:

```text
frontend/src/pages/TournamentDetails.tsx
backend/internal/modules/payment/handler.go
backend/internal/modules/payment/service.go
```

If AddisPay does not open:

```text
backend/pkg/addispay/client.go
ADDISPAY_API_KEY
ADDISPAY_BASE_URL
```

If return page fails:

```text
frontend/src/pages/PaymentReturn.tsx
backend/internal/modules/payment/service.go
ADDISPAY_WEBHOOK_SECRET
```

If webhook fails:

```text
backend/internal/modules/payment/handler.go
backend/pkg/addispay/client.go
ADDISPAY_WEBHOOK_SECRET
```

If registration stays pending:

```text
backend/internal/modules/payment/service.go
backend/internal/repository/payment.go
backend/internal/repository/registration.go
registrations.payment_status
```
