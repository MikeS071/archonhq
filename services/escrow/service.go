package escrow

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
	StatusPendingLock = "pending_lock"
	StatusLocked      = "locked"
	StatusReleased    = "released"
	StatusRefunded    = "refunded"
	StatusDisputed    = "disputed"
)

const (
	TransferLock    = "lock"
	TransferRelease = "release"
	TransferRefund  = "refund"
	TransferFee     = "fee"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrAlreadyExists  = errors.New("already exists")
	ErrInvalidRequest = errors.New("invalid request")
)

type Escrow struct {
	EscrowID         string         `json:"escrow_id"`
	TenantID         string         `json:"tenant_id"`
	ListingID        string         `json:"listing_id"`
	FundingAccountID string         `json:"funding_account_id"`
	Currency         string         `json:"currency"`
	TotalLocked      float64        `json:"total_locked"`
	ReleasedAmount   float64        `json:"released_amount"`
	RefundedAmount   float64        `json:"refunded_amount"`
	Status           string         `json:"status"`
	Metadata         map[string]any `json:"metadata,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type Transfer struct {
	TransferID   string         `json:"transfer_id"`
	EscrowID     string         `json:"escrow_id"`
	TenantID     string         `json:"tenant_id"`
	TransferType string         `json:"transfer_type"`
	Amount       float64        `json:"amount"`
	Currency     string         `json:"currency"`
	Status       string         `json:"status"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

type EnsureEscrowRequest struct {
	TenantID         string
	EscrowID         string
	ListingID        string
	FundingAccountID string
	Currency         string
	Metadata         map[string]any
}

type AdjustEscrowRequest struct {
	TenantID     string
	EscrowID     string
	Amount       float64
	TransferType string
	Metadata     map[string]any
}

type Service struct {
	mu sync.RWMutex

	escrows           map[string]Escrow
	escrowByListing   map[string]string
	transfersByEscrow map[string][]Transfer
	seq               uint64
}

func New() *Service {
	return &Service{
		escrows:           map[string]Escrow{},
		escrowByListing:   map[string]string{},
		transfersByEscrow: map[string][]Transfer{},
	}
}

func (s *Service) EnsureEscrow(_ context.Context, req EnsureEscrowRequest) (Escrow, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.ListingID) == "" || strings.TrimSpace(req.FundingAccountID) == "" || strings.TrimSpace(req.Currency) == "" {
		return Escrow{}, fmt.Errorf("%w: tenant_id, listing_id, funding_account_id, and currency are required", ErrInvalidRequest)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	listingKey := tenantScoped(req.TenantID, req.ListingID)
	if escrowID, ok := s.escrowByListing[listingKey]; ok {
		existing, ok := s.escrows[tenantScoped(req.TenantID, escrowID)]
		if ok {
			return existing, nil
		}
	}

	escrowID := strings.TrimSpace(req.EscrowID)
	if escrowID == "" {
		escrowID = s.nextIDLocked("escrow")
	}
	key := tenantScoped(req.TenantID, escrowID)
	if _, ok := s.escrows[key]; ok {
		return Escrow{}, fmt.Errorf("%w: escrow_id", ErrAlreadyExists)
	}

	now := time.Now().UTC()
	escrow := Escrow{
		EscrowID:         escrowID,
		TenantID:         strings.TrimSpace(req.TenantID),
		ListingID:        strings.TrimSpace(req.ListingID),
		FundingAccountID: strings.TrimSpace(req.FundingAccountID),
		Currency:         strings.ToUpper(strings.TrimSpace(req.Currency)),
		Status:           StatusPendingLock,
		Metadata:         copyMap(req.Metadata),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	s.escrows[key] = escrow
	s.escrowByListing[listingKey] = escrow.EscrowID
	return escrow, nil
}

func (s *Service) GetEscrow(_ context.Context, tenantID, escrowID string) (Escrow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	escrow, ok := s.escrows[tenantScoped(tenantID, escrowID)]
	if !ok {
		return Escrow{}, ErrNotFound
	}
	return escrow, nil
}

func (s *Service) GetEscrowByListing(_ context.Context, tenantID, listingID string) (Escrow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	escrowID, ok := s.escrowByListing[tenantScoped(tenantID, listingID)]
	if !ok {
		return Escrow{}, ErrNotFound
	}
	escrow, ok := s.escrows[tenantScoped(tenantID, escrowID)]
	if !ok {
		return Escrow{}, ErrNotFound
	}
	return escrow, nil
}

func (s *Service) Lock(_ context.Context, req AdjustEscrowRequest) (Escrow, Transfer, error) {
	return s.adjustEscrow(req, TransferLock)
}

func (s *Service) Release(_ context.Context, req AdjustEscrowRequest) (Escrow, Transfer, error) {
	return s.adjustEscrow(req, TransferRelease)
}

func (s *Service) Refund(_ context.Context, req AdjustEscrowRequest) (Escrow, Transfer, error) {
	return s.adjustEscrow(req, TransferRefund)
}

func (s *Service) ListTransfers(_ context.Context, tenantID, escrowID string) ([]Transfer, error) {
	if _, err := s.GetEscrow(context.Background(), tenantID, escrowID); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	items := append([]Transfer{}, s.transfersByEscrow[tenantScoped(tenantID, escrowID)]...)
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})
	return items, nil
}

func (s *Service) ListEscrows(_ context.Context, tenantID string) []Escrow {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Escrow, 0)
	for _, escrow := range s.escrows {
		if escrow.TenantID != tenantID {
			continue
		}
		items = append(items, escrow)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items
}

func (s *Service) adjustEscrow(req AdjustEscrowRequest, expectedType string) (Escrow, Transfer, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.EscrowID) == "" {
		return Escrow{}, Transfer{}, fmt.Errorf("%w: tenant_id and escrow_id are required", ErrInvalidRequest)
	}
	if req.Amount <= 0 {
		return Escrow{}, Transfer{}, fmt.Errorf("%w: amount must be > 0", ErrInvalidRequest)
	}
	if strings.TrimSpace(req.TransferType) == "" {
		req.TransferType = expectedType
	}
	if req.TransferType != expectedType {
		return Escrow{}, Transfer{}, fmt.Errorf("%w: transfer_type mismatch", ErrInvalidRequest)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := tenantScoped(req.TenantID, req.EscrowID)
	escrow, ok := s.escrows[key]
	if !ok {
		return Escrow{}, Transfer{}, ErrNotFound
	}

	switch req.TransferType {
	case TransferLock:
		escrow.TotalLocked += req.Amount
		escrow.Status = StatusLocked
	case TransferRelease:
		if escrow.TotalLocked-(escrow.ReleasedAmount+escrow.RefundedAmount) < req.Amount {
			return Escrow{}, Transfer{}, fmt.Errorf("%w: release exceeds locked balance", ErrInvalidRequest)
		}
		escrow.ReleasedAmount += req.Amount
		escrow.Status = StatusReleased
	case TransferRefund:
		if escrow.TotalLocked-(escrow.ReleasedAmount+escrow.RefundedAmount) < req.Amount {
			return Escrow{}, Transfer{}, fmt.Errorf("%w: refund exceeds locked balance", ErrInvalidRequest)
		}
		escrow.RefundedAmount += req.Amount
		escrow.Status = StatusRefunded
	default:
		return Escrow{}, Transfer{}, fmt.Errorf("%w: unsupported transfer_type", ErrInvalidRequest)
	}

	escrow.UpdatedAt = time.Now().UTC()
	s.escrows[key] = escrow

	transfer := Transfer{
		TransferID:   s.nextIDLocked("et"),
		EscrowID:     escrow.EscrowID,
		TenantID:     escrow.TenantID,
		TransferType: req.TransferType,
		Amount:       req.Amount,
		Currency:     escrow.Currency,
		Status:       "posted",
		Metadata:     copyMap(req.Metadata),
		CreatedAt:    time.Now().UTC(),
	}
	transfersKey := tenantScoped(req.TenantID, req.EscrowID)
	s.transfersByEscrow[transfersKey] = append(s.transfersByEscrow[transfersKey], transfer)

	return escrow, transfer, nil
}

func tenantScoped(tenantID, id string) string {
	return strings.TrimSpace(tenantID) + "::" + strings.TrimSpace(id)
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

func (s *Service) nextIDLocked(prefix string) string {
	s.seq++
	return fmt.Sprintf("%s_%06d", prefix, s.seq)
}
