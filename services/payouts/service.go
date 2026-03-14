package payouts

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	AccountStatusPending   = "pending_verification"
	AccountStatusActive    = "active"
	AccountStatusSuspended = "suspended"
	AccountStatusClosed    = "closed"
)

const (
	PayoutStatusRequested  = "requested"
	PayoutStatusProcessing = "processing"
	PayoutStatusCompleted  = "completed"
	PayoutStatusFailed     = "failed"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrAlreadyExists  = errors.New("already exists")
	ErrInvalidRequest = errors.New("invalid request")
)

type Account struct {
	TenantID           string         `json:"tenant_id"`
	PayoutAccountID    string         `json:"payout_account_id"`
	OwnerProfileID     string         `json:"owner_profile_id"`
	Provider           string         `json:"provider"`
	ProviderAccountRef string         `json:"provider_account_ref"`
	Jurisdiction       string         `json:"jurisdiction"`
	Status             string         `json:"status"`
	Metadata           map[string]any `json:"metadata,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

type Payout struct {
	TenantID        string         `json:"tenant_id"`
	PayoutID        string         `json:"payout_id"`
	PayoutAccountID string         `json:"payout_account_id"`
	EscrowID        string         `json:"escrow_id,omitempty"`
	Amount          float64        `json:"amount"`
	Currency        string         `json:"currency"`
	Status          string         `json:"status"`
	FailureReason   string         `json:"failure_reason,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type CreateAccountRequest struct {
	TenantID           string
	PayoutAccountID    string
	OwnerProfileID     string
	Provider           string
	ProviderAccountRef string
	Jurisdiction       string
	Status             string
	Metadata           map[string]any
}

type RequestPayoutRequest struct {
	TenantID        string
	PayoutID        string
	PayoutAccountID string
	EscrowID        string
	Amount          float64
	Currency        string
	AutoComplete    bool
	Metadata        map[string]any
}

type Service struct {
	mu sync.RWMutex

	accounts             map[string]Account
	payouts              map[string]Payout
	allowedJurisdictions map[string]struct{}
	seq                  uint64
}

func New() *Service {
	return &Service{
		accounts: map[string]Account{},
		payouts:  map[string]Payout{},
		allowedJurisdictions: map[string]struct{}{
			"AU": {},
			"US": {},
			"GB": {},
			"CA": {},
		},
	}
}

func (s *Service) CreateAccount(_ context.Context, req CreateAccountRequest) (Account, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.OwnerProfileID) == "" || strings.TrimSpace(req.Provider) == "" || strings.TrimSpace(req.ProviderAccountRef) == "" || strings.TrimSpace(req.Jurisdiction) == "" {
		return Account{}, fmt.Errorf("%w: tenant_id, owner_profile_id, provider, provider_account_ref, and jurisdiction are required", ErrInvalidRequest)
	}
	accountID := strings.TrimSpace(req.PayoutAccountID)
	if accountID == "" {
		accountID = "payoutacct_" + fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = AccountStatusActive
	}
	if !isAccountStatus(status) {
		return Account{}, fmt.Errorf("%w: invalid account status", ErrInvalidRequest)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := tenantScoped(req.TenantID, accountID)
	if _, ok := s.accounts[key]; ok {
		return Account{}, fmt.Errorf("%w: payout account", ErrAlreadyExists)
	}
	now := time.Now().UTC()
	account := Account{
		TenantID:           req.TenantID,
		PayoutAccountID:    accountID,
		OwnerProfileID:     req.OwnerProfileID,
		Provider:           strings.TrimSpace(req.Provider),
		ProviderAccountRef: strings.TrimSpace(req.ProviderAccountRef),
		Jurisdiction:       strings.ToUpper(strings.TrimSpace(req.Jurisdiction)),
		Status:             status,
		Metadata:           copyMap(req.Metadata),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	s.accounts[key] = account
	return account, nil
}

func (s *Service) GetAccount(_ context.Context, tenantID, payoutAccountID string) (Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	account, ok := s.accounts[tenantScoped(tenantID, payoutAccountID)]
	if !ok {
		return Account{}, ErrNotFound
	}
	return account, nil
}

func (s *Service) ListAccounts(_ context.Context, tenantID string) []Account {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Account, 0)
	for _, account := range s.accounts {
		if account.TenantID != tenantID {
			continue
		}
		items = append(items, account)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items
}

func (s *Service) RequestPayout(_ context.Context, req RequestPayoutRequest) (Payout, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.PayoutAccountID) == "" || req.Amount <= 0 || strings.TrimSpace(req.Currency) == "" {
		return Payout{}, fmt.Errorf("%w: tenant_id, payout_account_id, amount, and currency are required", ErrInvalidRequest)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	accountKey := tenantScoped(req.TenantID, req.PayoutAccountID)
	account, ok := s.accounts[accountKey]
	if !ok {
		return Payout{}, ErrNotFound
	}
	if account.Status != AccountStatusActive {
		return Payout{}, fmt.Errorf("%w: payout account must be active", ErrInvalidRequest)
	}
	if _, ok := s.allowedJurisdictions[account.Jurisdiction]; !ok {
		return Payout{}, fmt.Errorf("%w: payout jurisdiction not allowed", ErrInvalidRequest)
	}

	payoutID := strings.TrimSpace(req.PayoutID)
	if payoutID == "" {
		payoutID = s.nextIDLocked("payout")
	}
	payoutKey := tenantScoped(req.TenantID, payoutID)
	if _, ok := s.payouts[payoutKey]; ok {
		return Payout{}, fmt.Errorf("%w: payout_id", ErrAlreadyExists)
	}

	status := PayoutStatusRequested
	if req.AutoComplete {
		status = PayoutStatusCompleted
	}
	now := time.Now().UTC()
	payout := Payout{
		TenantID:        req.TenantID,
		PayoutID:        payoutID,
		PayoutAccountID: req.PayoutAccountID,
		EscrowID:        strings.TrimSpace(req.EscrowID),
		Amount:          req.Amount,
		Currency:        strings.ToUpper(strings.TrimSpace(req.Currency)),
		Status:          status,
		Metadata:        copyMap(req.Metadata),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	s.payouts[payoutKey] = payout
	return payout, nil
}

func (s *Service) GetPayout(_ context.Context, tenantID, payoutID string) (Payout, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	payout, ok := s.payouts[tenantScoped(tenantID, payoutID)]
	if !ok {
		return Payout{}, ErrNotFound
	}
	return payout, nil
}

func (s *Service) ListPayouts(_ context.Context, tenantID string) []Payout {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Payout, 0)
	for _, payout := range s.payouts {
		if payout.TenantID != tenantID {
			continue
		}
		items = append(items, payout)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items
}

func isAccountStatus(v string) bool {
	switch strings.TrimSpace(v) {
	case AccountStatusPending, AccountStatusActive, AccountStatusSuspended, AccountStatusClosed:
		return true
	default:
		return false
	}
}

func copyMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	out := map[string]any{}
	for k, v := range in {
		out[k] = v
	}
	return out
}

func tenantScoped(tenantID, id string) string {
	return strings.TrimSpace(tenantID) + "::" + strings.TrimSpace(id)
}

func (s *Service) nextIDLocked(prefix string) string {
	s.seq++
	return fmt.Sprintf("%s_%06d", prefix, s.seq)
}
