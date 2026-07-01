package repository

import (
	"encoding/json"
	"time"
)

type Profile struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	DisplayName *string   `json:"display_name,omitempty"`
	Email       *string   `json:"email,omitempty"`
	AvatarUrl   *string   `json:"avatar_url,omitempty"`
	Role        string    `json:"role"`
	Bio         *string   `json:"bio,omitempty"`
	Country     *string   `json:"country,omitempty"`
	CountryCode *string   `json:"country_code,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProfileStats struct {
	TournamentsPlayed int64 `json:"tournaments_played"`
	TournamentsHosted int64 `json:"tournaments_hosted"`
	Wins              int64 `json:"wins"`
}

type ProfileSearchResult struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name,omitempty"`
	AvatarUrl   *string `json:"avatar_url,omitempty"`
	Role        string  `json:"role"`
}

type ProfileWithStats struct {
	Profile
	Stats ProfileStats `json:"stats"`
}

type Tournament struct {
	ID                   string     `json:"id"`
	Title                string     `json:"title"`
	Game                 string     `json:"game"`
	Description          *string    `json:"description,omitempty"`
	Rules                *string    `json:"rules,omitempty"`
	Format               string     `json:"format"`
	Status               string     `json:"status"`
	MaxParticipants      int32      `json:"max_participants"`
	MinParticipants      int32      `json:"min_participants"`
	TeamSize             int32      `json:"team_size"`
	EntryFee             int32      `json:"entry_fee"`
	PrizePool            int32      `json:"prize_pool"`
	Currency             string     `json:"currency"`
	Location             string     `json:"location"`
	BestOf               int32      `json:"best_of"`
	PlatformFeePct       int32      `json:"platform_fee_pct"`
	OrganizerFeePct      int32      `json:"organizer_fee_pct"`
	StartDate            time.Time  `json:"start_date"`
	EndDate              *time.Time `json:"end_date,omitempty"`
	RegistrationDeadline *time.Time `json:"registration_deadline,omitempty"`
	RegistrationCloseAt  *time.Time `json:"registration_close_at,omitempty"`
	OrganizerID          string     `json:"organizer_id"`
	BannerUrl            *string    `json:"banner_url,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type Registration struct {
	ID            string    `json:"id"`
	TournamentID  string    `json:"tournament_id"`
	UserID        string    `json:"user_id"`
	PaymentStatus string    `json:"payment_status"`
	Seed          *int32    `json:"seed,omitempty"`
	RegisteredAt  time.Time `json:"registered_at"`
}

type Match struct {
	ID                    string     `json:"id"`
	TournamentID          string     `json:"tournament_id"`
	Round                 int32      `json:"round"`
	Position              int32      `json:"position"`
	PlayerAID             *string    `json:"player_a_id,omitempty"`
	PlayerBID             *string    `json:"player_b_id,omitempty"`
	WinnerID              *string    `json:"winner_id,omitempty"`
	Score                 *string    `json:"score,omitempty"`
	Status                string     `json:"status"`
	ScheduledAt           *time.Time `json:"scheduled_at,omitempty"`
	CompletedAt           *time.Time `json:"completed_at,omitempty"`
	EvidenceScreenshotURL *string    `json:"evidence_screenshot_url,omitempty"`
	EvidenceVideoURL      *string    `json:"evidence_video_url,omitempty"`
	EvidenceNotes         *string    `json:"evidence_notes,omitempty"`
	ResultSubmittedBy     *string    `json:"result_submitted_by,omitempty"`
	ResultConfirmedAt     *time.Time `json:"result_confirmed_at,omitempty"`
	DisputeStatus         string     `json:"dispute_status"`
	DisputeReason         *string    `json:"dispute_reason,omitempty"`
	DisputeOpenedBy       *string    `json:"dispute_opened_by,omitempty"`
	DisputeOpenedAt       *time.Time `json:"dispute_opened_at,omitempty"`
	DisputeResolvedAt     *time.Time `json:"dispute_resolved_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type AuditLogEntry struct {
	ID         string          `json:"id"`
	ActorID    *string         `json:"actor_id,omitempty"`
	ActorRole  *string         `json:"actor_role,omitempty"`
	Action     string          `json:"action"`
	EntityType string          `json:"entity_type"`
	EntityID   *string         `json:"entity_id,omitempty"`
	Details    json.RawMessage `json:"details,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type Payout struct {
	ID                       string     `json:"id"`
	TournamentID             string     `json:"tournament_id"`
	WinnerID                 string     `json:"winner_id"`
	Amount                   int32      `json:"amount"`
	Currency                 string     `json:"currency"`
	Status                   string     `json:"status"`
	PayoutMethod             *string    `json:"payout_method,omitempty"`
	PhoneNumber              *string    `json:"phone_number,omitempty"`
	TelebirrNumber           *string    `json:"telebirr_number,omitempty"`
	BankName                 *string    `json:"bank_name,omitempty"`
	BankAccountName          *string    `json:"bank_account_name,omitempty"`
	BankAccountNumber        *string    `json:"bank_account_number,omitempty"`
	PayoutDetailsSubmittedAt *time.Time `json:"payout_details_submitted_at,omitempty"`
	PaidAt                   *time.Time `json:"paid_at,omitempty"`
	PaidBy                   *string    `json:"paid_by,omitempty"`
	Note                     *string    `json:"note,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

type Notification struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Type      string          `json:"type"`
	Title     string          `json:"title"`
	Message   string          `json:"message"`
	Link      *string         `json:"link,omitempty"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	ReadAt    *time.Time      `json:"read_at,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type Payment struct {
	ID                string          `json:"id"`
	UserID            string          `json:"user_id"`
	TournamentID      string          `json:"tournament_id"`
	Amount            int32           `json:"amount"`
	Currency          string          `json:"currency"`
	Status            string          `json:"status"`
	AddispayRef       *string         `json:"addispay_ref,omitempty"`
	PaymentURL        *string         `json:"payment_url,omitempty"`
	ProviderStatus    *string         `json:"provider_status,omitempty"`
	ProviderPaymentID *string         `json:"provider_payment_id,omitempty"`
	VerifiedAt        *time.Time      `json:"verified_at,omitempty"`
	WebhookReceivedAt *time.Time      `json:"webhook_received_at,omitempty"`
	FailureReason     *string         `json:"failure_reason,omitempty"`
	RefundStatus      string          `json:"refund_status"`
	RefundReason      *string         `json:"refund_reason,omitempty"`
	RefundRequestedAt *time.Time      `json:"refund_requested_at,omitempty"`
	RefundedAt        *time.Time      `json:"refunded_at,omitempty"`
	RefundedBy        *string         `json:"refunded_by,omitempty"`
	Metadata          json.RawMessage `json:"metadata,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type GetTournamentRow = Tournament
type CreateTournamentRow = Tournament
type ListTournamentsRow = Tournament
type ListTournamentsByOrganizerRow = Tournament
type UpdateTournamentRow = Tournament
type UpdateTournamentStatusRow = Tournament

type GetMatchRow = Match

type CreateRegistrationRow = Registration

type ListRegistrationsByTournamentRow struct {
	Registration
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name,omitempty"`
	AvatarUrl   *string `json:"avatar_url,omitempty"`
}

type ListRegistrationsByUserRow struct {
	Registration
	Title     string    `json:"title"`
	Game      string    `json:"game"`
	Status    string    `json:"status"`
	StartDate time.Time `json:"start_date"`
	PrizePool int32     `json:"prize_pool"`
}
