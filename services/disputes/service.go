package disputes

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
	StatusOpen     = "open"
	StatusResolved = "resolved"
	StatusAppealed = "appealed"
)

const (
	TypeNonDelivery        = "non_delivery"
	TypeAcceptanceDisagree = "acceptance_disagreement"
	TypeSpecDrift          = "spec_drift"
	TypeRequesterDefault   = "requester_default"
	TypeExecutorMisconduct = "executor_misconduct"
	TypeSealedInputMisuse  = "sealed_input_misuse"
)

const (
	EscrowActionNone    = "none"
	EscrowActionRelease = "release"
	EscrowActionRefund  = "refund"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrAlreadyExists  = errors.New("already exists")
	ErrInvalidRequest = errors.New("invalid request")
)

type Dispute struct {
	DisputeID            string         `json:"dispute_id"`
	TenantID             string         `json:"tenant_id"`
	ListingID            string         `json:"listing_id"`
	EscrowID             string         `json:"escrow_id,omitempty"`
	ClaimID              string         `json:"claim_id,omitempty"`
	DisputeType          string         `json:"dispute_type"`
	Reason               string         `json:"reason"`
	Status               string         `json:"status"`
	OpenedBy             string         `json:"opened_by"`
	Decision             string         `json:"decision,omitempty"`
	FeeShift             float64        `json:"fee_shift,omitempty"`
	EscrowReleaseAction  string         `json:"escrow_release_action,omitempty"`
	ReputationAdjustment map[string]any `json:"reputation_adjustment,omitempty"`
	AppealAllowed        bool           `json:"appeal_allowed"`
	AppealReason         string         `json:"appeal_reason,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

type OpenDisputeRequest struct {
	TenantID    string
	DisputeID   string
	ListingID   string
	EscrowID    string
	ClaimID     string
	DisputeType string
	Reason      string
	OpenedBy    string
}

type ResolveDisputeRequest struct {
	TenantID             string
	DisputeID            string
	Decision             string
	FeeShift             float64
	EscrowReleaseAction  string
	ReputationAdjustment map[string]any
	AppealAllowed        bool
}

type AppealDisputeRequest struct {
	TenantID  string
	DisputeID string
	AppealBy  string
	Reason    string
}

type Service struct {
	mu sync.RWMutex

	disputes map[string]Dispute
	seq      uint64
}

func New() *Service {
	return &Service{
		disputes: map[string]Dispute{},
	}
}

func (s *Service) OpenDispute(_ context.Context, req OpenDisputeRequest) (Dispute, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.ListingID) == "" || strings.TrimSpace(req.DisputeType) == "" || strings.TrimSpace(req.Reason) == "" {
		return Dispute{}, fmt.Errorf("%w: tenant_id, listing_id, dispute_type, and reason are required", ErrInvalidRequest)
	}
	if !isDisputeType(req.DisputeType) {
		return Dispute{}, fmt.Errorf("%w: invalid dispute_type", ErrInvalidRequest)
	}

	disputeID := strings.TrimSpace(req.DisputeID)
	if disputeID == "" {
		disputeID = s.nextID("disp")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := tenantScoped(req.TenantID, disputeID)
	if _, ok := s.disputes[key]; ok {
		return Dispute{}, fmt.Errorf("%w: dispute_id", ErrAlreadyExists)
	}

	now := time.Now().UTC()
	dispute := Dispute{
		DisputeID:   disputeID,
		TenantID:    req.TenantID,
		ListingID:   req.ListingID,
		EscrowID:    strings.TrimSpace(req.EscrowID),
		ClaimID:     strings.TrimSpace(req.ClaimID),
		DisputeType: req.DisputeType,
		Reason:      req.Reason,
		Status:      StatusOpen,
		OpenedBy:    strings.TrimSpace(req.OpenedBy),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.disputes[key] = dispute
	return dispute, nil
}

func (s *Service) GetDispute(_ context.Context, tenantID, disputeID string) (Dispute, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dispute, ok := s.disputes[tenantScoped(tenantID, disputeID)]
	if !ok {
		return Dispute{}, ErrNotFound
	}
	return dispute, nil
}

func (s *Service) ResolveDispute(_ context.Context, req ResolveDisputeRequest) (Dispute, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.DisputeID) == "" || strings.TrimSpace(req.Decision) == "" {
		return Dispute{}, fmt.Errorf("%w: tenant_id, dispute_id, and decision are required", ErrInvalidRequest)
	}
	action := strings.TrimSpace(req.EscrowReleaseAction)
	if action == "" {
		action = EscrowActionNone
	}
	if !isEscrowAction(action) {
		return Dispute{}, fmt.Errorf("%w: invalid escrow_release_action", ErrInvalidRequest)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := tenantScoped(req.TenantID, req.DisputeID)
	dispute, ok := s.disputes[key]
	if !ok {
		return Dispute{}, ErrNotFound
	}
	if dispute.Status != StatusOpen && dispute.Status != StatusAppealed {
		return Dispute{}, fmt.Errorf("%w: dispute must be open or appealed to resolve", ErrInvalidRequest)
	}

	dispute.Status = StatusResolved
	dispute.Decision = strings.TrimSpace(req.Decision)
	dispute.FeeShift = req.FeeShift
	dispute.EscrowReleaseAction = action
	dispute.ReputationAdjustment = copyMap(req.ReputationAdjustment)
	dispute.AppealAllowed = req.AppealAllowed
	dispute.UpdatedAt = time.Now().UTC()
	s.disputes[key] = dispute
	return dispute, nil
}

func (s *Service) AppealDispute(_ context.Context, req AppealDisputeRequest) (Dispute, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.DisputeID) == "" || strings.TrimSpace(req.Reason) == "" {
		return Dispute{}, fmt.Errorf("%w: tenant_id, dispute_id, and reason are required", ErrInvalidRequest)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := tenantScoped(req.TenantID, req.DisputeID)
	dispute, ok := s.disputes[key]
	if !ok {
		return Dispute{}, ErrNotFound
	}
	if dispute.Status != StatusResolved {
		return Dispute{}, fmt.Errorf("%w: dispute must be resolved before appeal", ErrInvalidRequest)
	}
	if !dispute.AppealAllowed {
		return Dispute{}, fmt.Errorf("%w: appeal not allowed", ErrInvalidRequest)
	}

	dispute.Status = StatusAppealed
	dispute.AppealReason = strings.TrimSpace(req.Reason)
	dispute.UpdatedAt = time.Now().UTC()
	s.disputes[key] = dispute
	return dispute, nil
}

func (s *Service) ListDisputes(_ context.Context, tenantID string) []Dispute {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Dispute, 0)
	for _, dispute := range s.disputes {
		if dispute.TenantID != tenantID {
			continue
		}
		items = append(items, dispute)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items
}

func isDisputeType(v string) bool {
	switch strings.TrimSpace(v) {
	case TypeNonDelivery, TypeAcceptanceDisagree, TypeSpecDrift, TypeRequesterDefault, TypeExecutorMisconduct, TypeSealedInputMisuse:
		return true
	default:
		return false
	}
}

func isEscrowAction(v string) bool {
	switch strings.TrimSpace(v) {
	case EscrowActionNone, EscrowActionRelease, EscrowActionRefund:
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

func (s *Service) nextID(prefix string) string {
	s.seq++
	return fmt.Sprintf("%s_%06d", prefix, s.seq)
}
