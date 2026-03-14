package marketplace

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	ProfileTypeRequester = "requester"
	ProfileTypeExecutor  = "executor"
	ProfileTypeHybrid    = "hybrid"
)

const (
	VerificationPending  = "pending"
	VerificationVerified = "verified"
	VerificationRejected = "rejected"
)

const (
	WorkClassPublicOpen       = "public_open"
	WorkClassPublicSealed     = "public_sealed"
	WorkClassRestrictedMarket = "restricted_market"
	WorkClassPrivateTenant    = "private_tenant_only"
)

const (
	ListingModeFixedOpenClaim   = "fixed_price_open_claim"
	ListingModeFixedBidSelect   = "fixed_price_bid_select"
	ListingModeReserveAuction   = "reserve_price_auction"
	ListingModeRedundantCompete = "redundant_competition"
	ListingModeShardMarket      = "decomposed_shard_market"
)

const (
	ListingStatusDraft     = "draft"
	ListingStatusPublished = "published"
	ListingStatusCancelled = "cancelled"
)

const (
	ClaimTypeWholeTask = "whole_task"
	ClaimTypeShard     = "shard"
	ClaimTypeVerifier  = "verifier"
	ClaimTypeReducer   = "reducer"
	ClaimTypeRedundant = "redundant_competitor"
)

const (
	ClaimStatusSubmitted = "submitted"
	ClaimStatusWithdrawn = "withdrawn"
	ClaimStatusAwarded   = "awarded"
	ClaimStatusRejected  = "rejected"
)

const (
	BidStatusSubmitted = "submitted"
	BidStatusAccepted  = "accepted"
	BidStatusRejected  = "rejected"
)

const (
	FundingStatusActive = "active"
)

const (
	maxPublishedListingsPerRequester = 5
	pendingRequesterBudgetCap        = 200.0
)

var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInvalidRequest    = errors.New("invalid request")
	ErrInsufficientFunds = errors.New("insufficient funded reserve")
	ErrClaimHoarding     = errors.New("claim hoarding limit reached")
	ErrAntiSpamControl   = errors.New("anti-spam control violation")
	ErrSybilControl      = errors.New("anti-sybil control violation")
)

type Profile struct {
	TenantID           string    `json:"tenant_id"`
	ProfileID          string    `json:"profile_id"`
	ProfileType        string    `json:"profile_type"`
	DisplayName        string    `json:"display_name"`
	VerificationStatus string    `json:"verification_status"`
	ExecutorTier       string    `json:"executor_tier,omitempty"`
	Status             string    `json:"status"`
	CapabilityTags     []string  `json:"capability_tags,omitempty"`
	RegionAllowlist    []string  `json:"region_allowlist,omitempty"`
	WorkClassAllowlist []string  `json:"work_class_allowlist,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type ReputationSnapshot struct {
	TenantID            string    `json:"tenant_id"`
	ProfileID           string    `json:"profile_id"`
	RejectionRatio      float64   `json:"rejection_ratio"`
	ClaimCompletionRate float64   `json:"claim_completion_rate"`
	DisputeLossRate     float64   `json:"dispute_loss_rate"`
	PayoutSuccessRate   float64   `json:"payout_success_rate"`
	Score               float64   `json:"score"`
	AsOf                time.Time `json:"as_of"`
}

type FundingAccount struct {
	TenantID         string    `json:"tenant_id"`
	AccountID        string    `json:"account_id"`
	OwnerProfileID   string    `json:"owner_profile_id"`
	Currency         string    `json:"currency"`
	AvailableBalance float64   `json:"available_balance"`
	ReservedBalance  float64   `json:"reserved_balance"`
	ReservePolicyID  string    `json:"reserve_policy_id,omitempty"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Listing struct {
	TenantID           string         `json:"tenant_id"`
	ListingID          string         `json:"listing_id"`
	TaskID             string         `json:"task_id"`
	RequesterProfileID string         `json:"requester_profile_id"`
	WorkClass          string         `json:"work_class"`
	ListingMode        string         `json:"listing_mode"`
	BudgetTotal        float64        `json:"budget_total"`
	BudgetPerShard     float64        `json:"budget_per_shard,omitempty"`
	Currency           string         `json:"currency"`
	FundingAccountID   string         `json:"funding_account_id"`
	Status             string         `json:"status"`
	PublishReason      string         `json:"publish_reason,omitempty"`
	CancelReason       string         `json:"cancel_reason,omitempty"`
	ContractSnapshot   map[string]any `json:"contract_snapshot,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	PublishedAt        time.Time      `json:"published_at,omitempty"`
	CancelledAt        time.Time      `json:"cancelled_at,omitempty"`
}

type Claim struct {
	TenantID           string         `json:"tenant_id"`
	ClaimID            string         `json:"claim_id"`
	ListingID          string         `json:"listing_id"`
	ExecutorProfileID  string         `json:"executor_profile_id"`
	ClaimType          string         `json:"claim_type"`
	BondAmount         float64        `json:"bond_amount,omitempty"`
	Status             string         `json:"status"`
	PolicyChecksPassed bool           `json:"policy_checks_passed"`
	Metadata           map[string]any `json:"metadata,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

type Bid struct {
	TenantID          string         `json:"tenant_id"`
	BidID             string         `json:"bid_id"`
	ListingID         string         `json:"listing_id"`
	ExecutorProfileID string         `json:"executor_profile_id"`
	Amount            float64        `json:"amount"`
	Currency          string         `json:"currency"`
	Status            string         `json:"status"`
	Metadata          map[string]any `json:"metadata,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

type CreateProfileRequest struct {
	TenantID           string
	ProfileID          string
	ProfileType        string
	DisplayName        string
	VerificationStatus string
	ExecutorTier       string
	CapabilityTags     []string
	RegionAllowlist    []string
	WorkClassAllowlist []string
}

type PatchProfileRequest struct {
	TenantID           string
	ProfileID          string
	DisplayName        *string
	VerificationStatus *string
	ExecutorTier       *string
	Status             *string
	CapabilityTags     *[]string
	RegionAllowlist    *[]string
	WorkClassAllowlist *[]string
}

type CreateFundingAccountRequest struct {
	TenantID        string
	AccountID       string
	OwnerProfileID  string
	Currency        string
	InitialBalance  float64
	ReservePolicyID string
}

type CreateListingRequest struct {
	TenantID           string
	ListingID          string
	TaskID             string
	RequesterProfileID string
	WorkClass          string
	ListingMode        string
	BudgetTotal        float64
	BudgetPerShard     float64
	Currency           string
	FundingAccountID   string
	ContractSnapshot   map[string]any
}

type PublishListingRequest struct {
	TenantID    string
	ListingID   string
	Reason      string
	PublishedBy string
}

type CancelListingRequest struct {
	TenantID    string
	ListingID   string
	Reason      string
	CancelledBy string
}

type CreateClaimRequest struct {
	TenantID           string
	ClaimID            string
	ListingID          string
	ExecutorProfileID  string
	ClaimType          string
	BondAmount         float64
	PolicyChecksPassed bool
	Metadata           map[string]any
}

type WithdrawClaimRequest struct {
	TenantID string
	ClaimID  string
	Reason   string
}

type AwardClaimRequest struct {
	TenantID           string
	ClaimID            string
	PolicyChecksPassed bool
}

type CreateBidRequest struct {
	TenantID          string
	BidID             string
	ListingID         string
	ExecutorProfileID string
	Amount            float64
	Currency          string
	Metadata          map[string]any
}

type AcceptBidRequest struct {
	TenantID           string
	BidID              string
	PolicyChecksPassed bool
}

type ListListingsOptions struct {
	Status    string
	WorkClass string
	Limit     int
}

type Service struct {
	mu sync.RWMutex

	profiles        map[string]Profile
	fundingAccounts map[string]FundingAccount
	listings        map[string]Listing
	claims          map[string]Claim
	bids            map[string]Bid
	seq             uint64
}

func New() *Service {
	return &Service{
		profiles:        map[string]Profile{},
		fundingAccounts: map[string]FundingAccount{},
		listings:        map[string]Listing{},
		claims:          map[string]Claim{},
		bids:            map[string]Bid{},
	}
}

func (s *Service) CreateProfile(_ context.Context, req CreateProfileRequest) (Profile, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.ProfileID) == "" || strings.TrimSpace(req.ProfileType) == "" || strings.TrimSpace(req.DisplayName) == "" {
		return Profile{}, fmt.Errorf("%w: tenant_id, profile_id, profile_type, and display_name are required", ErrInvalidRequest)
	}
	if !isProfileType(req.ProfileType) {
		return Profile{}, fmt.Errorf("%w: invalid profile_type", ErrInvalidRequest)
	}
	if req.VerificationStatus == "" {
		req.VerificationStatus = VerificationPending
	}
	if !isVerificationStatus(req.VerificationStatus) {
		return Profile{}, fmt.Errorf("%w: invalid verification_status", ErrInvalidRequest)
	}

	key := tenantScoped(req.TenantID, req.ProfileID)
	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.profiles[key]; ok {
		return Profile{}, fmt.Errorf("%w: profile_id", ErrAlreadyExists)
	}

	profile := Profile{
		TenantID:           strings.TrimSpace(req.TenantID),
		ProfileID:          strings.TrimSpace(req.ProfileID),
		ProfileType:        strings.TrimSpace(req.ProfileType),
		DisplayName:        strings.TrimSpace(req.DisplayName),
		VerificationStatus: strings.TrimSpace(req.VerificationStatus),
		ExecutorTier:       strings.TrimSpace(req.ExecutorTier),
		Status:             "active",
		CapabilityTags:     copySlice(req.CapabilityTags),
		RegionAllowlist:    copySlice(req.RegionAllowlist),
		WorkClassAllowlist: copySlice(req.WorkClassAllowlist),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	s.profiles[key] = profile
	return profile, nil
}

func (s *Service) GetProfile(_ context.Context, tenantID, profileID string) (Profile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	profile, ok := s.profiles[tenantScoped(tenantID, profileID)]
	if !ok {
		return Profile{}, ErrNotFound
	}
	return profile, nil
}

func (s *Service) PatchProfile(_ context.Context, req PatchProfileRequest) (Profile, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.ProfileID) == "" {
		return Profile{}, fmt.Errorf("%w: tenant_id and profile_id are required", ErrInvalidRequest)
	}

	key := tenantScoped(req.TenantID, req.ProfileID)
	s.mu.Lock()
	defer s.mu.Unlock()

	profile, ok := s.profiles[key]
	if !ok {
		return Profile{}, ErrNotFound
	}

	if req.DisplayName != nil {
		v := strings.TrimSpace(*req.DisplayName)
		if v == "" {
			return Profile{}, fmt.Errorf("%w: display_name cannot be empty", ErrInvalidRequest)
		}
		profile.DisplayName = v
	}
	if req.VerificationStatus != nil {
		v := strings.TrimSpace(*req.VerificationStatus)
		if !isVerificationStatus(v) {
			return Profile{}, fmt.Errorf("%w: invalid verification_status", ErrInvalidRequest)
		}
		profile.VerificationStatus = v
	}
	if req.ExecutorTier != nil {
		profile.ExecutorTier = strings.TrimSpace(*req.ExecutorTier)
	}
	if req.Status != nil {
		v := strings.TrimSpace(*req.Status)
		if v == "" {
			return Profile{}, fmt.Errorf("%w: status cannot be empty", ErrInvalidRequest)
		}
		profile.Status = v
	}
	if req.CapabilityTags != nil {
		profile.CapabilityTags = copySlice(*req.CapabilityTags)
	}
	if req.RegionAllowlist != nil {
		profile.RegionAllowlist = copySlice(*req.RegionAllowlist)
	}
	if req.WorkClassAllowlist != nil {
		for _, workClass := range *req.WorkClassAllowlist {
			if !isWorkClass(workClass) {
				return Profile{}, fmt.Errorf("%w: invalid work_class_allowlist entry", ErrInvalidRequest)
			}
		}
		profile.WorkClassAllowlist = copySlice(*req.WorkClassAllowlist)
	}
	profile.UpdatedAt = time.Now().UTC()
	s.profiles[key] = profile
	return profile, nil
}

func (s *Service) GetReputation(_ context.Context, tenantID, profileID string) (ReputationSnapshot, error) {
	profile, err := s.GetProfile(context.Background(), tenantID, profileID)
	if err != nil {
		return ReputationSnapshot{}, err
	}

	h := sha1.Sum([]byte(profile.TenantID + "::" + profile.ProfileID))
	rejectionRatio := boundedRatio(h[0], 0.01, 0.25)
	claimCompletion := boundedRatio(h[1], 0.65, 0.99)
	disputeLoss := boundedRatio(h[2], 0.01, 0.20)
	payoutSuccess := boundedRatio(h[3], 0.75, 0.995)
	score := (0.35 * claimCompletion) + (0.30 * payoutSuccess) + (0.20 * (1.0 - rejectionRatio)) + (0.15 * (1.0 - disputeLoss))

	return ReputationSnapshot{
		TenantID:            profile.TenantID,
		ProfileID:           profile.ProfileID,
		RejectionRatio:      rejectionRatio,
		ClaimCompletionRate: claimCompletion,
		DisputeLossRate:     disputeLoss,
		PayoutSuccessRate:   payoutSuccess,
		Score:               score,
		AsOf:                time.Now().UTC(),
	}, nil
}

func (s *Service) CreateFundingAccount(_ context.Context, req CreateFundingAccountRequest) (FundingAccount, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.AccountID) == "" || strings.TrimSpace(req.OwnerProfileID) == "" || strings.TrimSpace(req.Currency) == "" {
		return FundingAccount{}, fmt.Errorf("%w: tenant_id, account_id, owner_profile_id, and currency are required", ErrInvalidRequest)
	}
	if req.InitialBalance < 0 {
		return FundingAccount{}, fmt.Errorf("%w: initial_balance must be >= 0", ErrInvalidRequest)
	}

	key := tenantScoped(req.TenantID, req.AccountID)
	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.fundingAccounts[key]; ok {
		return FundingAccount{}, fmt.Errorf("%w: funding account", ErrAlreadyExists)
	}
	if _, ok := s.profiles[tenantScoped(req.TenantID, req.OwnerProfileID)]; !ok {
		return FundingAccount{}, ErrNotFound
	}

	acct := FundingAccount{
		TenantID:         req.TenantID,
		AccountID:        req.AccountID,
		OwnerProfileID:   req.OwnerProfileID,
		Currency:         strings.ToUpper(strings.TrimSpace(req.Currency)),
		AvailableBalance: req.InitialBalance,
		ReservedBalance:  0,
		ReservePolicyID:  strings.TrimSpace(req.ReservePolicyID),
		Status:           FundingStatusActive,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	s.fundingAccounts[key] = acct
	return acct, nil
}

func (s *Service) GetFundingAccount(_ context.Context, tenantID, accountID string) (FundingAccount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	acct, ok := s.fundingAccounts[tenantScoped(tenantID, accountID)]
	if !ok {
		return FundingAccount{}, ErrNotFound
	}
	return acct, nil
}

func (s *Service) CreateListing(_ context.Context, req CreateListingRequest) (Listing, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.ListingID) == "" || strings.TrimSpace(req.TaskID) == "" || strings.TrimSpace(req.RequesterProfileID) == "" || strings.TrimSpace(req.WorkClass) == "" || strings.TrimSpace(req.ListingMode) == "" || strings.TrimSpace(req.Currency) == "" || strings.TrimSpace(req.FundingAccountID) == "" {
		return Listing{}, fmt.Errorf("%w: tenant_id, listing_id, task_id, requester_profile_id, work_class, listing_mode, currency, and funding_account_id are required", ErrInvalidRequest)
	}
	if req.BudgetTotal <= 0 {
		return Listing{}, fmt.Errorf("%w: budget_total must be > 0", ErrInvalidRequest)
	}
	if !isWorkClass(req.WorkClass) {
		return Listing{}, fmt.Errorf("%w: invalid work_class", ErrInvalidRequest)
	}
	if !isListingMode(req.ListingMode) {
		return Listing{}, fmt.Errorf("%w: invalid listing_mode", ErrInvalidRequest)
	}

	now := time.Now().UTC()
	key := tenantScoped(req.TenantID, req.ListingID)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.listings[key]; ok {
		return Listing{}, fmt.Errorf("%w: listing_id", ErrAlreadyExists)
	}
	profile, ok := s.profiles[tenantScoped(req.TenantID, req.RequesterProfileID)]
	if !ok {
		return Listing{}, ErrNotFound
	}
	if profile.ProfileType != ProfileTypeRequester && profile.ProfileType != ProfileTypeHybrid {
		return Listing{}, fmt.Errorf("%w: requester_profile_id must be requester or hybrid", ErrInvalidRequest)
	}
	if _, ok := s.fundingAccounts[tenantScoped(req.TenantID, req.FundingAccountID)]; !ok {
		return Listing{}, ErrNotFound
	}

	listing := Listing{
		TenantID:           req.TenantID,
		ListingID:          req.ListingID,
		TaskID:             req.TaskID,
		RequesterProfileID: req.RequesterProfileID,
		WorkClass:          req.WorkClass,
		ListingMode:        req.ListingMode,
		BudgetTotal:        req.BudgetTotal,
		BudgetPerShard:     req.BudgetPerShard,
		Currency:           strings.ToUpper(strings.TrimSpace(req.Currency)),
		FundingAccountID:   req.FundingAccountID,
		Status:             ListingStatusDraft,
		ContractSnapshot:   copyMap(req.ContractSnapshot),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	s.listings[key] = listing
	return listing, nil
}

func (s *Service) GetListing(_ context.Context, tenantID, listingID string) (Listing, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	listing, ok := s.listings[tenantScoped(tenantID, listingID)]
	if !ok {
		return Listing{}, ErrNotFound
	}
	return listing, nil
}

func (s *Service) ListListings(_ context.Context, tenantID string, opts ListListingsOptions) []Listing {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if opts.Limit <= 0 {
		opts.Limit = 50
	}

	statusFilter := strings.TrimSpace(opts.Status)
	workClassFilter := strings.TrimSpace(opts.WorkClass)

	items := make([]Listing, 0)
	for _, listing := range s.listings {
		if listing.TenantID != tenantID {
			continue
		}
		if statusFilter != "" && listing.Status != statusFilter {
			continue
		}
		if workClassFilter != "" && listing.WorkClass != workClassFilter {
			continue
		}
		items = append(items, listing)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	if len(items) > opts.Limit {
		return items[:opts.Limit]
	}
	return items
}

func (s *Service) PublishListing(_ context.Context, req PublishListingRequest) (Listing, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.ListingID) == "" {
		return Listing{}, fmt.Errorf("%w: tenant_id and listing_id are required", ErrInvalidRequest)
	}

	key := tenantScoped(req.TenantID, req.ListingID)

	s.mu.Lock()
	defer s.mu.Unlock()

	listing, ok := s.listings[key]
	if !ok {
		return Listing{}, ErrNotFound
	}
	if listing.Status != ListingStatusDraft {
		return Listing{}, fmt.Errorf("%w: listing must be draft to publish", ErrInvalidRequest)
	}
	if listing.WorkClass == WorkClassPrivateTenant {
		return Listing{}, fmt.Errorf("%w: private_tenant_only listings cannot be published to open market", ErrInvalidRequest)
	}
	requesterProfile, ok := s.profiles[tenantScoped(req.TenantID, listing.RequesterProfileID)]
	if !ok {
		return Listing{}, ErrNotFound
	}
	if requesterProfile.VerificationStatus == VerificationRejected {
		return Listing{}, fmt.Errorf("%w: requester verification rejected for market publication", ErrSybilControl)
	}
	if requesterProfile.VerificationStatus == VerificationPending && listing.BudgetTotal > pendingRequesterBudgetCap {
		return Listing{}, fmt.Errorf("%w: pending requester verification budget cap exceeded", ErrSybilControl)
	}
	publishedForRequester := 0
	for _, existingListing := range s.listings {
		if existingListing.TenantID != req.TenantID || existingListing.RequesterProfileID != listing.RequesterProfileID {
			continue
		}
		if existingListing.Status == ListingStatusPublished {
			publishedForRequester++
		}
	}
	if publishedForRequester >= maxPublishedListingsPerRequester {
		return Listing{}, fmt.Errorf("%w: requester published listing quota exceeded", ErrAntiSpamControl)
	}

	accountKey := tenantScoped(req.TenantID, listing.FundingAccountID)
	account, ok := s.fundingAccounts[accountKey]
	if !ok {
		return Listing{}, ErrNotFound
	}
	if strings.ToUpper(strings.TrimSpace(account.Currency)) != strings.ToUpper(strings.TrimSpace(listing.Currency)) {
		return Listing{}, fmt.Errorf("%w: listing and funding account currency mismatch", ErrInvalidRequest)
	}
	if account.AvailableBalance < listing.BudgetTotal {
		return Listing{}, ErrInsufficientFunds
	}

	account.AvailableBalance -= listing.BudgetTotal
	account.ReservedBalance += listing.BudgetTotal
	account.UpdatedAt = time.Now().UTC()
	s.fundingAccounts[accountKey] = account

	listing.Status = ListingStatusPublished
	listing.PublishReason = strings.TrimSpace(req.Reason)
	listing.PublishedAt = time.Now().UTC()
	listing.UpdatedAt = listing.PublishedAt
	s.listings[key] = listing

	return listing, nil
}

func (s *Service) CancelListing(_ context.Context, req CancelListingRequest) (Listing, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.ListingID) == "" {
		return Listing{}, fmt.Errorf("%w: tenant_id and listing_id are required", ErrInvalidRequest)
	}

	key := tenantScoped(req.TenantID, req.ListingID)
	s.mu.Lock()
	defer s.mu.Unlock()

	listing, ok := s.listings[key]
	if !ok {
		return Listing{}, ErrNotFound
	}
	if listing.Status == ListingStatusCancelled {
		return listing, nil
	}

	if listing.Status == ListingStatusPublished {
		accountKey := tenantScoped(req.TenantID, listing.FundingAccountID)
		account, ok := s.fundingAccounts[accountKey]
		if !ok {
			return Listing{}, ErrNotFound
		}
		unlockAmount := listing.BudgetTotal
		if account.ReservedBalance < unlockAmount {
			unlockAmount = account.ReservedBalance
		}
		account.ReservedBalance -= unlockAmount
		account.AvailableBalance += unlockAmount
		account.UpdatedAt = time.Now().UTC()
		s.fundingAccounts[accountKey] = account
	}

	listing.Status = ListingStatusCancelled
	listing.CancelReason = strings.TrimSpace(req.Reason)
	listing.CancelledAt = time.Now().UTC()
	listing.UpdatedAt = listing.CancelledAt
	s.listings[key] = listing
	return listing, nil
}

func (s *Service) CreateClaim(_ context.Context, req CreateClaimRequest) (Claim, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.ListingID) == "" || strings.TrimSpace(req.ExecutorProfileID) == "" {
		return Claim{}, fmt.Errorf("%w: tenant_id, listing_id, and executor_profile_id are required", ErrInvalidRequest)
	}
	if req.ClaimType == "" {
		req.ClaimType = ClaimTypeWholeTask
	}
	if !isClaimType(req.ClaimType) {
		return Claim{}, fmt.Errorf("%w: invalid claim_type", ErrInvalidRequest)
	}
	if req.BondAmount < 0 {
		return Claim{}, fmt.Errorf("%w: bond_amount must be >= 0", ErrInvalidRequest)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	listing, ok := s.listings[tenantScoped(req.TenantID, req.ListingID)]
	if !ok {
		return Claim{}, ErrNotFound
	}
	if listing.Status != ListingStatusPublished {
		return Claim{}, fmt.Errorf("%w: listing must be published", ErrInvalidRequest)
	}

	executor, ok := s.profiles[tenantScoped(req.TenantID, req.ExecutorProfileID)]
	if !ok {
		return Claim{}, ErrNotFound
	}
	if executor.ProfileType != ProfileTypeExecutor && executor.ProfileType != ProfileTypeHybrid {
		return Claim{}, fmt.Errorf("%w: executor_profile_id must be executor or hybrid", ErrInvalidRequest)
	}
	if listing.WorkClass == WorkClassPublicSealed && executor.VerificationStatus != VerificationVerified {
		return Claim{}, fmt.Errorf("%w: sealed-work claims require verified executor profile", ErrSybilControl)
	}

	activeCount := 0
	for _, claim := range s.claims {
		if claim.TenantID != req.TenantID || claim.ExecutorProfileID != req.ExecutorProfileID {
			continue
		}
		if isActiveClaimStatus(claim.Status) {
			activeCount++
		}
		if claim.ListingID == req.ListingID && isActiveClaimStatus(claim.Status) {
			return Claim{}, fmt.Errorf("%w: executor already has active claim on listing", ErrInvalidRequest)
		}
	}
	if activeCount >= 3 {
		return Claim{}, ErrClaimHoarding
	}

	claimID := strings.TrimSpace(req.ClaimID)
	if claimID == "" {
		claimID = s.nextIDLocked("claim")
	}
	key := tenantScoped(req.TenantID, claimID)
	if _, ok := s.claims[key]; ok {
		return Claim{}, fmt.Errorf("%w: claim_id", ErrAlreadyExists)
	}

	now := time.Now().UTC()
	claim := Claim{
		TenantID:           req.TenantID,
		ClaimID:            claimID,
		ListingID:          req.ListingID,
		ExecutorProfileID:  req.ExecutorProfileID,
		ClaimType:          req.ClaimType,
		BondAmount:         req.BondAmount,
		Status:             ClaimStatusSubmitted,
		PolicyChecksPassed: req.PolicyChecksPassed,
		Metadata:           copyMap(req.Metadata),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	s.claims[key] = claim
	return claim, nil
}

func (s *Service) GetClaim(_ context.Context, tenantID, claimID string) (Claim, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	claim, ok := s.claims[tenantScoped(tenantID, claimID)]
	if !ok {
		return Claim{}, ErrNotFound
	}
	return claim, nil
}

func (s *Service) WithdrawClaim(_ context.Context, req WithdrawClaimRequest) (Claim, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.ClaimID) == "" {
		return Claim{}, fmt.Errorf("%w: tenant_id and claim_id are required", ErrInvalidRequest)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := tenantScoped(req.TenantID, req.ClaimID)
	claim, ok := s.claims[key]
	if !ok {
		return Claim{}, ErrNotFound
	}
	if claim.Status != ClaimStatusSubmitted {
		return Claim{}, fmt.Errorf("%w: only submitted claims can be withdrawn", ErrInvalidRequest)
	}
	claim.Status = ClaimStatusWithdrawn
	claim.UpdatedAt = time.Now().UTC()
	s.claims[key] = claim
	return claim, nil
}

func (s *Service) AwardClaim(_ context.Context, req AwardClaimRequest) (Claim, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.ClaimID) == "" {
		return Claim{}, fmt.Errorf("%w: tenant_id and claim_id are required", ErrInvalidRequest)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := tenantScoped(req.TenantID, req.ClaimID)
	claim, ok := s.claims[key]
	if !ok {
		return Claim{}, ErrNotFound
	}
	if claim.Status != ClaimStatusSubmitted {
		return Claim{}, fmt.Errorf("%w: claim must be submitted to award", ErrInvalidRequest)
	}

	listing, ok := s.listings[tenantScoped(req.TenantID, claim.ListingID)]
	if !ok {
		return Claim{}, ErrNotFound
	}
	if listing.WorkClass == WorkClassPublicSealed && !req.PolicyChecksPassed {
		return Claim{}, fmt.Errorf("%w: sealed work requires policy checks before award", ErrInvalidRequest)
	}

	for k, existing := range s.claims {
		if existing.TenantID != req.TenantID || existing.ListingID != claim.ListingID || existing.ClaimID == claim.ClaimID {
			continue
		}
		if existing.Status == ClaimStatusSubmitted {
			existing.Status = ClaimStatusRejected
			existing.UpdatedAt = time.Now().UTC()
			s.claims[k] = existing
		}
	}

	claim.Status = ClaimStatusAwarded
	claim.PolicyChecksPassed = req.PolicyChecksPassed
	claim.UpdatedAt = time.Now().UTC()
	s.claims[key] = claim
	return claim, nil
}

func (s *Service) CreateBid(_ context.Context, req CreateBidRequest) (Bid, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.ListingID) == "" || strings.TrimSpace(req.ExecutorProfileID) == "" || req.Amount <= 0 || strings.TrimSpace(req.Currency) == "" {
		return Bid{}, fmt.Errorf("%w: tenant_id, listing_id, executor_profile_id, amount, and currency are required", ErrInvalidRequest)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	listing, ok := s.listings[tenantScoped(req.TenantID, req.ListingID)]
	if !ok {
		return Bid{}, ErrNotFound
	}
	if listing.Status != ListingStatusPublished {
		return Bid{}, fmt.Errorf("%w: listing must be published", ErrInvalidRequest)
	}
	if strings.ToUpper(strings.TrimSpace(listing.Currency)) != strings.ToUpper(strings.TrimSpace(req.Currency)) {
		return Bid{}, fmt.Errorf("%w: bid currency mismatch", ErrInvalidRequest)
	}

	executor, ok := s.profiles[tenantScoped(req.TenantID, req.ExecutorProfileID)]
	if !ok {
		return Bid{}, ErrNotFound
	}
	if executor.ProfileType != ProfileTypeExecutor && executor.ProfileType != ProfileTypeHybrid {
		return Bid{}, fmt.Errorf("%w: executor_profile_id must be executor or hybrid", ErrInvalidRequest)
	}
	if listing.WorkClass == WorkClassPublicSealed && executor.VerificationStatus != VerificationVerified {
		return Bid{}, fmt.Errorf("%w: sealed-work bids require verified executor profile", ErrSybilControl)
	}

	bidID := strings.TrimSpace(req.BidID)
	if bidID == "" {
		bidID = s.nextIDLocked("bid")
	}
	key := tenantScoped(req.TenantID, bidID)
	if _, ok := s.bids[key]; ok {
		return Bid{}, fmt.Errorf("%w: bid_id", ErrAlreadyExists)
	}

	now := time.Now().UTC()
	bid := Bid{
		TenantID:          req.TenantID,
		BidID:             bidID,
		ListingID:         req.ListingID,
		ExecutorProfileID: req.ExecutorProfileID,
		Amount:            req.Amount,
		Currency:          strings.ToUpper(strings.TrimSpace(req.Currency)),
		Status:            BidStatusSubmitted,
		Metadata:          copyMap(req.Metadata),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	s.bids[key] = bid
	return bid, nil
}

func (s *Service) GetBid(_ context.Context, tenantID, bidID string) (Bid, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bid, ok := s.bids[tenantScoped(tenantID, bidID)]
	if !ok {
		return Bid{}, ErrNotFound
	}
	return bid, nil
}

func (s *Service) AcceptBid(_ context.Context, req AcceptBidRequest) (Bid, error) {
	if strings.TrimSpace(req.TenantID) == "" || strings.TrimSpace(req.BidID) == "" {
		return Bid{}, fmt.Errorf("%w: tenant_id and bid_id are required", ErrInvalidRequest)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := tenantScoped(req.TenantID, req.BidID)
	bid, ok := s.bids[key]
	if !ok {
		return Bid{}, ErrNotFound
	}
	if bid.Status != BidStatusSubmitted {
		return Bid{}, fmt.Errorf("%w: bid must be submitted to accept", ErrInvalidRequest)
	}

	listing, ok := s.listings[tenantScoped(req.TenantID, bid.ListingID)]
	if !ok {
		return Bid{}, ErrNotFound
	}
	if listing.WorkClass == WorkClassPublicSealed && !req.PolicyChecksPassed {
		return Bid{}, fmt.Errorf("%w: sealed work requires policy checks before bid acceptance", ErrInvalidRequest)
	}

	for k, existing := range s.bids {
		if existing.TenantID != req.TenantID || existing.ListingID != bid.ListingID || existing.BidID == bid.BidID {
			continue
		}
		if existing.Status == BidStatusSubmitted {
			existing.Status = BidStatusRejected
			existing.UpdatedAt = time.Now().UTC()
			s.bids[k] = existing
		}
	}

	bid.Status = BidStatusAccepted
	bid.UpdatedAt = time.Now().UTC()
	s.bids[key] = bid
	return bid, nil
}

func (s *Service) ListClaims(_ context.Context, tenantID, listingID string) []Claim {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Claim, 0)
	for _, claim := range s.claims {
		if claim.TenantID != tenantID {
			continue
		}
		if strings.TrimSpace(listingID) != "" && claim.ListingID != listingID {
			continue
		}
		items = append(items, claim)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items
}

func (s *Service) ListBids(_ context.Context, tenantID, listingID string) []Bid {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Bid, 0)
	for _, bid := range s.bids {
		if bid.TenantID != tenantID {
			continue
		}
		if strings.TrimSpace(listingID) != "" && bid.ListingID != listingID {
			continue
		}
		items = append(items, bid)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items
}

func isActiveClaimStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case ClaimStatusSubmitted, ClaimStatusAwarded:
		return true
	default:
		return false
	}
}

func isProfileType(v string) bool {
	switch strings.TrimSpace(v) {
	case ProfileTypeRequester, ProfileTypeExecutor, ProfileTypeHybrid:
		return true
	default:
		return false
	}
}

func isVerificationStatus(v string) bool {
	switch strings.TrimSpace(v) {
	case VerificationPending, VerificationVerified, VerificationRejected:
		return true
	default:
		return false
	}
}

func isWorkClass(v string) bool {
	switch strings.TrimSpace(v) {
	case WorkClassPublicOpen, WorkClassPublicSealed, WorkClassRestrictedMarket, WorkClassPrivateTenant:
		return true
	default:
		return false
	}
}

func isListingMode(v string) bool {
	switch strings.TrimSpace(v) {
	case ListingModeFixedOpenClaim, ListingModeFixedBidSelect, ListingModeReserveAuction, ListingModeRedundantCompete, ListingModeShardMarket:
		return true
	default:
		return false
	}
}

func isClaimType(v string) bool {
	switch strings.TrimSpace(v) {
	case ClaimTypeWholeTask, ClaimTypeShard, ClaimTypeVerifier, ClaimTypeReducer, ClaimTypeRedundant:
		return true
	default:
		return false
	}
}

func tenantScoped(tenantID, id string) string {
	return strings.TrimSpace(tenantID) + "::" + strings.TrimSpace(id)
}

func boundedRatio(b byte, low, high float64) float64 {
	r := float64(b) / 255.0
	return low + ((high - low) * r)
}

func copySlice(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
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
