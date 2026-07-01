package addispay

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	APIKey  string
	BaseURL string
	http    *http.Client
}

type PaymentRequest struct {
	Data    OrderData `json:"data"`
	Message string    `json:"message"`
}

type OrderData struct {
	RedirectURL    string      `json:"redirect_url"`
	CancelURL      string      `json:"cancel_url"`
	SuccessURL     string      `json:"success_url"`
	ErrorURL       string      `json:"error_url"`
	OrderReason    string      `json:"order_reason"`
	Currency       string      `json:"currency"`
	Email          string      `json:"email"`
	FirstName      string      `json:"first_name"`
	LastName       string      `json:"last_name"`
	Nonce          string      `json:"nonce"`
	OrderDetail    OrderDetail `json:"order_detail"`
	PhoneNumber    string      `json:"phone_number"`
	SessionExpired string      `json:"session_expired"`
	TotalAmount    string      `json:"total_amount"`
	TxRef          string      `json:"tx_ref"`
}

type OrderDetail struct {
	Amount      int    `json:"amount"`
	Description string `json:"description"`
}

type PaymentResponse struct {
	CheckoutURL        string `json:"checkout_url"`
	HostedCheckoutLink string `json:"hosted_checkout_url"`
	PaymentURL         string `json:"payment_url"`
	UUID               string `json:"uuid"`
	Status             string `json:"status"`
}

type WebhookPayload struct {
	Reference string `json:"reference"`
	Status    string `json:"status"`
	Amount    int    `json:"amount"`
	Currency  string `json:"currency"`
	Signature string `json:"signature"`
	PaymentID string `json:"payment_id"`
}

func NewClient(apiKey, baseURL string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: baseURL,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) CreatePayment(req PaymentRequest) (*PaymentResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal payment request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", strings.TrimRight(c.BaseURL, "/")+"/checkout-api/v1/create-order", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Auth", c.APIKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send payment request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("addispay error (%d): %s", resp.StatusCode, string(respBody))
	}

	var paymentResp PaymentResponse
	if err := json.Unmarshal(respBody, &paymentResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &paymentResp, nil
}

func (r *PaymentResponse) UnmarshalJSON(data []byte) error {
	type paymentResponse PaymentResponse
	var top paymentResponse
	if err := json.Unmarshal(data, &top); err != nil {
		return err
	}

	var wrapped struct {
		Data *paymentResponse `json:"data"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return err
	}
	if wrapped.Data != nil {
		fillEmpty(&top.CheckoutURL, wrapped.Data.CheckoutURL)
		fillEmpty(&top.HostedCheckoutLink, wrapped.Data.HostedCheckoutLink)
		fillEmpty(&top.PaymentURL, wrapped.Data.PaymentURL)
		fillEmpty(&top.UUID, wrapped.Data.UUID)
		fillEmpty(&top.Status, wrapped.Data.Status)
	}

	*r = PaymentResponse(top)
	return nil
}

func (r *PaymentResponse) HostedCheckoutURL() string {
	if r == nil {
		return ""
	}
	for _, candidate := range []string{r.HostedCheckoutLink, r.PaymentURL} {
		if validURL(candidate) {
			return strings.TrimSpace(candidate)
		}
	}
	if r.CheckoutURL == "" {
		return ""
	}
	if r.UUID == "" {
		if validURL(r.CheckoutURL) {
			return strings.TrimSpace(r.CheckoutURL)
		}
		return ""
	}
	return strings.TrimRight(strings.TrimSpace(r.CheckoutURL), "/") + "/" + strings.TrimLeft(strings.TrimSpace(r.UUID), "/")
}

func (p *WebhookPayload) UnmarshalJSON(data []byte) error {
	var decoded WebhookPayload
	if err := decodeWebhookFields(data, &decoded); err != nil {
		return err
	}

	var wrapped struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return err
	}
	if len(wrapped.Data) > 0 && string(wrapped.Data) != "null" {
		if err := decodeWebhookFields(wrapped.Data, &decoded); err != nil {
			return err
		}
	}

	*p = decoded
	return nil
}

func decodeWebhookFields(data []byte, payload *WebhookPayload) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	fillStringFromFields(&payload.Reference, fields,
		"reference",
		"tx_ref",
		"txRef",
		"transaction_reference",
		"transactionReference",
		"nonce",
		"order_id",
		"orderId",
	)
	fillStringFromFields(&payload.Status, fields,
		"status",
		"payment_status",
		"paymentStatus",
		"state",
	)
	fillIntFromFields(&payload.Amount, fields,
		"amount",
		"total_amount",
		"totalAmount",
	)
	fillStringFromFields(&payload.Currency, fields, "currency")
	fillStringFromFields(&payload.Signature, fields, "signature")
	fillStringFromFields(&payload.PaymentID, fields,
		"payment_id",
		"paymentId",
		"uuid",
		"id",
	)

	return nil
}

func fillEmpty(target *string, value string) {
	if strings.TrimSpace(*target) == "" {
		*target = strings.TrimSpace(value)
	}
}

func fillStringFromFields(target *string, fields map[string]json.RawMessage, keys ...string) {
	if strings.TrimSpace(*target) != "" {
		return
	}
	for _, key := range keys {
		raw, ok := fields[key]
		if !ok {
			continue
		}
		var value string
		if err := json.Unmarshal(raw, &value); err == nil && strings.TrimSpace(value) != "" {
			*target = strings.TrimSpace(value)
			return
		}
	}
}

func fillIntFromFields(target *int, fields map[string]json.RawMessage, keys ...string) {
	if *target != 0 {
		return
	}
	for _, key := range keys {
		raw, ok := fields[key]
		if !ok {
			continue
		}
		var intValue int
		if err := json.Unmarshal(raw, &intValue); err == nil {
			*target = intValue
			return
		}
		var floatValue float64
		if err := json.Unmarshal(raw, &floatValue); err == nil {
			*target = int(floatValue)
			return
		}
		var stringValue string
		if err := json.Unmarshal(raw, &stringValue); err != nil {
			continue
		}
		parsed, err := strconv.ParseFloat(strings.TrimSpace(stringValue), 64)
		if err == nil {
			*target = int(parsed)
			return
		}
	}
}

func validURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	u, err := url.Parse(raw)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// VerifyWebhookSignature checks the HMAC-SHA256 signature on a webhook payload.
// The signed string is reference|status|amount, matching what the gateway signs.
// Returns false if secret is empty (forces explicit configuration).
func (c *Client) VerifyWebhookSignature(payload WebhookPayload, secret string) bool {
	if secret == "" {
		return false
	}
	signed := fmt.Sprintf("%s|%s|%d", payload.Reference, payload.Status, payload.Amount)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signed))
	expected := hex.EncodeToString(mac.Sum(nil))
	provided, err := hex.DecodeString(payload.Signature)
	if err != nil {
		return false
	}
	expectedBytes, _ := hex.DecodeString(expected)
	return hmac.Equal(expectedBytes, provided)
}
