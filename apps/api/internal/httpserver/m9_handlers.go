package httpserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/MikeS071/archonhq/pkg/apierrors"
	disputessvc "github.com/MikeS071/archonhq/services/disputes"
	escrowsvc "github.com/MikeS071/archonhq/services/escrow"
	marketplacesvc "github.com/MikeS071/archonhq/services/marketplace"
	payoutssvc "github.com/MikeS071/archonhq/services/payouts"
	simulationsvc "github.com/MikeS071/archonhq/services/simulation"
)

var marketRolloutScenarioIDs = []string{
	"requester_default_v1",
	"dispute_griefing_v1",
	"sealed_task_leakage_v1",
	"claim_hoarding_v1",
}

type createMarketProfileRequest struct {
	ProfileID          string   `json:"profile_id,omitempty"`
	ProfileType        string   `json:"profile_type"`
	DisplayName        string   `json:"display_name"`
	VerificationStatus string   `json:"verification_status,omitempty"`
	ExecutorTier       string   `json:"executor_tier,omitempty"`
	CapabilityTags     []string `json:"capability_tags,omitempty"`
	RegionAllowlist    []string `json:"region_allowlist,omitempty"`
	WorkClassAllowlist []string `json:"work_class_allowlist,omitempty"`
}

type patchMarketProfileRequest struct {
	DisplayName        *string   `json:"display_name,omitempty"`
	VerificationStatus *string   `json:"verification_status,omitempty"`
	ExecutorTier       *string   `json:"executor_tier,omitempty"`
	Status             *string   `json:"status,omitempty"`
	CapabilityTags     *[]string `json:"capability_tags,omitempty"`
	RegionAllowlist    *[]string `json:"region_allowlist,omitempty"`
	WorkClassAllowlist *[]string `json:"work_class_allowlist,omitempty"`
}

type createFundingAccountRequest struct {
	AccountID       string  `json:"account_id,omitempty"`
	OwnerProfileID  string  `json:"owner_profile_id"`
	Currency        string  `json:"currency"`
	InitialBalance  float64 `json:"initial_balance"`
	ReservePolicyID string  `json:"reserve_policy_id,omitempty"`
}

type createMarketListingRequest struct {
	ListingID          string         `json:"listing_id,omitempty"`
	TaskID             string         `json:"task_id"`
	RequesterProfileID string         `json:"requester_profile_id"`
	WorkClass          string         `json:"work_class"`
	ListingMode        string         `json:"listing_mode"`
	BudgetTotal        float64        `json:"budget_total"`
	BudgetPerShard     float64        `json:"budget_per_shard,omitempty"`
	Currency           string         `json:"currency"`
	FundingAccountID   string         `json:"funding_account_id"`
	ContractSnapshot   map[string]any `json:"contract_snapshot,omitempty"`
}

type publishMarketListingRequest struct {
	Reason string `json:"reason,omitempty"`
}

type cancelMarketListingRequest struct {
	Reason string `json:"reason,omitempty"`
}

type createMarketClaimRequest struct {
	ClaimID            string         `json:"claim_id,omitempty"`
	ExecutorProfileID  string         `json:"executor_profile_id"`
	ClaimType          string         `json:"claim_type,omitempty"`
	BondAmount         float64        `json:"bond_amount,omitempty"`
	PolicyChecksPassed bool           `json:"policy_checks_passed,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
}

type withdrawMarketClaimRequest struct {
	Reason string `json:"reason,omitempty"`
}

type awardMarketClaimRequest struct {
	PolicyChecksPassed bool `json:"policy_checks_passed,omitempty"`
}

type createMarketBidRequest struct {
	BidID             string         `json:"bid_id,omitempty"`
	ExecutorProfileID string         `json:"executor_profile_id"`
	Amount            float64        `json:"amount"`
	Currency          string         `json:"currency"`
	Metadata          map[string]any `json:"metadata,omitempty"`
}

type acceptMarketBidRequest struct {
	PolicyChecksPassed bool `json:"policy_checks_passed,omitempty"`
}

type fundMarketListingRequest struct {
	EscrowID string         `json:"escrow_id,omitempty"`
	Amount   float64        `json:"amount,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type adjustEscrowRequest struct {
	Amount   float64        `json:"amount,omitempty"`
	Reason   string         `json:"reason,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type createPayoutAccountRequest struct {
	PayoutAccountID    string         `json:"payout_account_id,omitempty"`
	OwnerProfileID     string         `json:"owner_profile_id"`
	Provider           string         `json:"provider"`
	ProviderAccountRef string         `json:"provider_account_ref"`
	Jurisdiction       string         `json:"jurisdiction"`
	Status             string         `json:"status,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
}

type createPayoutRequest struct {
	PayoutID        string         `json:"payout_id,omitempty"`
	PayoutAccountID string         `json:"payout_account_id"`
	EscrowID        string         `json:"escrow_id,omitempty"`
	Amount          float64        `json:"amount"`
	Currency        string         `json:"currency"`
	AutoComplete    bool           `json:"auto_complete,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

type openMarketDisputeRequest struct {
	DisputeID   string `json:"dispute_id,omitempty"`
	ListingID   string `json:"listing_id"`
	EscrowID    string `json:"escrow_id,omitempty"`
	ClaimID     string `json:"claim_id,omitempty"`
	DisputeType string `json:"dispute_type"`
	Reason      string `json:"reason"`
}

type resolveMarketDisputeRequest struct {
	Decision             string         `json:"decision"`
	FeeShift             float64        `json:"fee_shift,omitempty"`
	EscrowReleaseAction  string         `json:"escrow_release_action,omitempty"`
	EscrowAmount         float64        `json:"escrow_amount,omitempty"`
	ReputationAdjustment map[string]any `json:"reputation_adjustment,omitempty"`
	AppealAllowed        bool           `json:"appeal_allowed,omitempty"`
}

type appealMarketDisputeRequest struct {
	Reason string `json:"reason"`
}

func (s *Server) handleCreateMarketProfileV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "requester") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for market profile create.", corrID, nil)
		return
	}

	var req createMarketProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	profileID := strings.TrimSpace(req.ProfileID)
	if profileID == "" {
		profileID = "mprof_" + randomID(6)
	}

	profile, err := s.marketplace.CreateProfile(r.Context(), marketplacesvc.CreateProfileRequest{
		TenantID:           actor.TenantID,
		ProfileID:          profileID,
		ProfileType:        strings.TrimSpace(req.ProfileType),
		DisplayName:        strings.TrimSpace(req.DisplayName),
		VerificationStatus: strings.TrimSpace(req.VerificationStatus),
		ExecutorTier:       strings.TrimSpace(req.ExecutorTier),
		CapabilityTags:     req.CapabilityTags,
		RegionAllowlist:    req.RegionAllowlist,
		WorkClassAllowlist: req.WorkClassAllowlist,
	})
	if err != nil {
		s.writeMarketplaceError(w, corrID, "market_profile_create_failed", "Failed to create market profile.", err)
		return
	}

	s.appendEvent(r, actor.TenantID, "market_profile", profile.ProfileID, "market.profile_created", map[string]any{
		"profile_type": profile.ProfileType,
		"display_name": profile.DisplayName,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"profile":        profile,
		"correlation_id": corrID,
	})
}

func (s *Server) handleGetMarketProfileV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "approver", "auditor", "requester") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for market profile read.", corrID, nil)
		return
	}

	profile, err := s.marketplace.GetProfile(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("profile_id")))
	if err != nil {
		s.writeMarketplaceError(w, corrID, "market_profile_not_found", "Market profile not found.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"profile":        profile,
		"correlation_id": corrID,
	})
}

func (s *Server) handlePatchMarketProfileV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "requester") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for market profile update.", corrID, nil)
		return
	}

	var req patchMarketProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	profile, err := s.marketplace.PatchProfile(r.Context(), marketplacesvc.PatchProfileRequest{
		TenantID:           actor.TenantID,
		ProfileID:          strings.TrimSpace(r.PathValue("profile_id")),
		DisplayName:        req.DisplayName,
		VerificationStatus: req.VerificationStatus,
		ExecutorTier:       req.ExecutorTier,
		Status:             req.Status,
		CapabilityTags:     req.CapabilityTags,
		RegionAllowlist:    req.RegionAllowlist,
		WorkClassAllowlist: req.WorkClassAllowlist,
	})
	if err != nil {
		s.writeMarketplaceError(w, corrID, "market_profile_update_failed", "Failed to update market profile.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"profile":        profile,
		"correlation_id": corrID,
	})
}

func (s *Server) handleMarketProfileReputationV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "approver", "auditor", "requester") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for market profile reputation read.", corrID, nil)
		return
	}

	profileID := strings.TrimSpace(r.PathValue("profile_id"))
	reputation, err := s.marketplace.GetReputation(r.Context(), actor.TenantID, profileID)
	if err != nil {
		s.writeMarketplaceError(w, corrID, "market_reputation_not_found", "Market reputation not found.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"profile_id":     profileID,
		"reputation":     reputation,
		"correlation_id": corrID,
	})
}

func (s *Server) handleCreateMarketFundingAccountV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "requester") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for funding account create.", corrID, nil)
		return
	}

	var req createFundingAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	accountID := strings.TrimSpace(req.AccountID)
	if accountID == "" {
		accountID = "fund_" + randomID(6)
	}

	account, err := s.marketplace.CreateFundingAccount(r.Context(), marketplacesvc.CreateFundingAccountRequest{
		TenantID:        actor.TenantID,
		AccountID:       accountID,
		OwnerProfileID:  strings.TrimSpace(req.OwnerProfileID),
		Currency:        strings.TrimSpace(req.Currency),
		InitialBalance:  req.InitialBalance,
		ReservePolicyID: strings.TrimSpace(req.ReservePolicyID),
	})
	if err != nil {
		s.writeMarketplaceError(w, corrID, "funding_account_create_failed", "Failed to create funding account.", err)
		return
	}

	s.appendEvent(r, actor.TenantID, "market_funding_account", account.AccountID, "market.funding_account_created", map[string]any{
		"owner_profile_id": account.OwnerProfileID,
		"currency":         account.Currency,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"funding_account": account,
		"correlation_id":  corrID,
	})
}

func (s *Server) handleGetMarketFundingAccountV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "approver", "auditor", "requester") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for funding account read.", corrID, nil)
		return
	}

	account, err := s.marketplace.GetFundingAccount(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("account_id")))
	if err != nil {
		s.writeMarketplaceError(w, corrID, "funding_account_not_found", "Funding account not found.", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"funding_account": account,
		"correlation_id":  corrID,
	})
}

func (s *Server) handleCreateMarketListingV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "requester") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for market listing create.", corrID, nil)
		return
	}

	var req createMarketListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	listingID := strings.TrimSpace(req.ListingID)
	if listingID == "" {
		listingID = "ml_" + randomID(6)
	}

	listing, err := s.marketplace.CreateListing(r.Context(), marketplacesvc.CreateListingRequest{
		TenantID:           actor.TenantID,
		ListingID:          listingID,
		TaskID:             strings.TrimSpace(req.TaskID),
		RequesterProfileID: strings.TrimSpace(req.RequesterProfileID),
		WorkClass:          strings.TrimSpace(req.WorkClass),
		ListingMode:        strings.TrimSpace(req.ListingMode),
		BudgetTotal:        req.BudgetTotal,
		BudgetPerShard:     req.BudgetPerShard,
		Currency:           strings.TrimSpace(req.Currency),
		FundingAccountID:   strings.TrimSpace(req.FundingAccountID),
		ContractSnapshot:   req.ContractSnapshot,
	})
	if err != nil {
		s.writeMarketplaceError(w, corrID, "market_listing_create_failed", "Failed to create market listing.", err)
		return
	}

	s.appendEvent(r, actor.TenantID, "market_listing", listing.ListingID, "market.listing_created", map[string]any{
		"work_class":   listing.WorkClass,
		"listing_mode": listing.ListingMode,
		"budget_total": listing.BudgetTotal,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"listing":        listing,
		"correlation_id": corrID,
	})
}

func (s *Server) handleListMarketListingsV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "approver", "auditor", "requester") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for market listing read.", corrID, nil)
		return
	}

	statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))
	if statusFilter == "" {
		statusFilter = marketplacesvc.ListingStatusPublished
	}
	items := s.marketplace.ListListings(r.Context(), actor.TenantID, marketplacesvc.ListListingsOptions{
		Status:    statusFilter,
		WorkClass: strings.TrimSpace(r.URL.Query().Get("work_class")),
		Limit:     atoiQueryDefault(r.URL.Query().Get("limit"), 50),
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"listings":       items,
		"correlation_id": corrID,
	})
}

func (s *Server) handleGetMarketListingV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "approver", "auditor", "requester") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for market listing read.", corrID, nil)
		return
	}

	listing, err := s.marketplace.GetListing(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("listing_id")))
	if err != nil {
		s.writeMarketplaceError(w, corrID, "market_listing_not_found", "Market listing not found.", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"listing":        listing,
		"correlation_id": corrID,
	})
}

func (s *Server) handlePublishMarketListingV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "requester") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for market listing publish.", corrID, nil)
		return
	}

	var req publishMarketListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	listing, err := s.marketplace.PublishListing(r.Context(), marketplacesvc.PublishListingRequest{
		TenantID:    actor.TenantID,
		ListingID:   strings.TrimSpace(r.PathValue("listing_id")),
		Reason:      strings.TrimSpace(req.Reason),
		PublishedBy: actor.ID,
	})
	if err != nil {
		s.writeMarketplaceError(w, corrID, "market_listing_publish_failed", "Failed to publish market listing.", err)
		return
	}

	s.appendEvent(r, actor.TenantID, "market_listing", listing.ListingID, "market.listing_published", map[string]any{
		"work_class":   listing.WorkClass,
		"budget_total": listing.BudgetTotal,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"listing":        listing,
		"correlation_id": corrID,
	})
}

func (s *Server) handleCancelMarketListingV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "requester") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for market listing cancel.", corrID, nil)
		return
	}

	var req cancelMarketListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	listing, err := s.marketplace.CancelListing(r.Context(), marketplacesvc.CancelListingRequest{
		TenantID:    actor.TenantID,
		ListingID:   strings.TrimSpace(r.PathValue("listing_id")),
		Reason:      strings.TrimSpace(req.Reason),
		CancelledBy: actor.ID,
	})
	if err != nil {
		s.writeMarketplaceError(w, corrID, "market_listing_cancel_failed", "Failed to cancel market listing.", err)
		return
	}

	s.appendEvent(r, actor.TenantID, "market_listing", listing.ListingID, "market.listing_cancelled", map[string]any{
		"reason": listing.CancelReason,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"listing":        listing,
		"correlation_id": corrID,
	})
}

func (s *Server) handleCreateMarketClaimV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "requester", "executor") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for market claim create.", corrID, nil)
		return
	}

	var req createMarketClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	claimID := strings.TrimSpace(req.ClaimID)
	if claimID == "" {
		claimID = "claim_" + randomID(6)
	}
	claim, err := s.marketplace.CreateClaim(r.Context(), marketplacesvc.CreateClaimRequest{
		TenantID:           actor.TenantID,
		ClaimID:            claimID,
		ListingID:          strings.TrimSpace(r.PathValue("listing_id")),
		ExecutorProfileID:  strings.TrimSpace(req.ExecutorProfileID),
		ClaimType:          strings.TrimSpace(req.ClaimType),
		BondAmount:         req.BondAmount,
		PolicyChecksPassed: req.PolicyChecksPassed,
		Metadata:           req.Metadata,
	})
	if err != nil {
		s.writeMarketplaceError(w, corrID, "market_claim_create_failed", "Failed to create market claim.", err)
		return
	}
	s.appendEvent(r, actor.TenantID, "market_claim", claim.ClaimID, "market.claim_created", map[string]any{
		"listing_id": claim.ListingID,
		"claim_type": claim.ClaimType,
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"claim":          claim,
		"correlation_id": corrID,
	})
}

func (s *Server) handleWithdrawMarketClaimV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "requester", "executor") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for market claim withdraw.", corrID, nil)
		return
	}

	var req withdrawMarketClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	claim, err := s.marketplace.WithdrawClaim(r.Context(), marketplacesvc.WithdrawClaimRequest{
		TenantID: actor.TenantID,
		ClaimID:  strings.TrimSpace(r.PathValue("claim_id")),
		Reason:   strings.TrimSpace(req.Reason),
	})
	if err != nil {
		s.writeMarketplaceError(w, corrID, "market_claim_withdraw_failed", "Failed to withdraw market claim.", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"claim":          claim,
		"correlation_id": corrID,
	})
}

func (s *Server) handleAwardMarketClaimV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "approver", "arbitrator", "requester") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for market claim award.", corrID, nil)
		return
	}

	var req awardMarketClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	claim, err := s.marketplace.AwardClaim(r.Context(), marketplacesvc.AwardClaimRequest{
		TenantID:           actor.TenantID,
		ClaimID:            strings.TrimSpace(r.PathValue("claim_id")),
		PolicyChecksPassed: req.PolicyChecksPassed,
	})
	if err != nil {
		s.writeMarketplaceError(w, corrID, "market_claim_award_failed", "Failed to award market claim.", err)
		return
	}

	listing, err := s.marketplace.GetListing(r.Context(), actor.TenantID, claim.ListingID)
	if err != nil {
		s.writeMarketplaceError(w, corrID, "market_listing_not_found", "Market listing not found.", err)
		return
	}
	escrowRecord, err := s.escrow.EnsureEscrow(r.Context(), escrowsvc.EnsureEscrowRequest{
		TenantID:         actor.TenantID,
		ListingID:        listing.ListingID,
		FundingAccountID: listing.FundingAccountID,
		Currency:         listing.Currency,
		Metadata:         map[string]any{"source": "claim_award", "claim_id": claim.ClaimID},
	})
	if err != nil {
		s.writeEscrowError(w, corrID, "escrow_ensure_failed", "Failed to ensure escrow for claim award.", err)
		return
	}
	escrowRecord, transfer, err := s.escrow.Lock(r.Context(), escrowsvc.AdjustEscrowRequest{
		TenantID:     actor.TenantID,
		EscrowID:     escrowRecord.EscrowID,
		Amount:       listing.BudgetTotal,
		TransferType: escrowsvc.TransferLock,
		Metadata:     map[string]any{"source": "claim_award", "claim_id": claim.ClaimID},
	})
	if err != nil {
		s.writeEscrowError(w, corrID, "escrow_lock_failed", "Failed to lock escrow after claim award.", err)
		return
	}

	s.appendEvent(r, actor.TenantID, "market_claim", claim.ClaimID, "market.claim_awarded", map[string]any{"listing_id": claim.ListingID})
	s.appendEvent(r, actor.TenantID, "escrow", escrowRecord.EscrowID, "escrow.locked", map[string]any{"amount": transfer.Amount, "source": "claim_award"})

	writeJSON(w, http.StatusOK, map[string]any{
		"claim":          claim,
		"escrow":         escrowRecord,
		"transfer":       transfer,
		"correlation_id": corrID,
	})
}

func (s *Server) handleCreateMarketBidV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "requester", "executor") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for market bid create.", corrID, nil)
		return
	}

	var req createMarketBidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	bidID := strings.TrimSpace(req.BidID)
	if bidID == "" {
		bidID = "bid_" + randomID(6)
	}
	bid, err := s.marketplace.CreateBid(r.Context(), marketplacesvc.CreateBidRequest{
		TenantID:          actor.TenantID,
		BidID:             bidID,
		ListingID:         strings.TrimSpace(r.PathValue("listing_id")),
		ExecutorProfileID: strings.TrimSpace(req.ExecutorProfileID),
		Amount:            req.Amount,
		Currency:          strings.TrimSpace(req.Currency),
		Metadata:          req.Metadata,
	})
	if err != nil {
		s.writeMarketplaceError(w, corrID, "market_bid_create_failed", "Failed to create market bid.", err)
		return
	}
	s.appendEvent(r, actor.TenantID, "market_bid", bid.BidID, "market.bid_submitted", map[string]any{
		"listing_id": bid.ListingID,
		"amount":     bid.Amount,
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"bid":            bid,
		"correlation_id": corrID,
	})
}

func (s *Server) handleAcceptMarketBidV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "approver", "arbitrator", "requester") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for market bid accept.", corrID, nil)
		return
	}

	var req acceptMarketBidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	bid, err := s.marketplace.AcceptBid(r.Context(), marketplacesvc.AcceptBidRequest{
		TenantID:           actor.TenantID,
		BidID:              strings.TrimSpace(r.PathValue("bid_id")),
		PolicyChecksPassed: req.PolicyChecksPassed,
	})
	if err != nil {
		s.writeMarketplaceError(w, corrID, "market_bid_accept_failed", "Failed to accept market bid.", err)
		return
	}

	listing, err := s.marketplace.GetListing(r.Context(), actor.TenantID, bid.ListingID)
	if err != nil {
		s.writeMarketplaceError(w, corrID, "market_listing_not_found", "Market listing not found.", err)
		return
	}
	escrowRecord, err := s.escrow.EnsureEscrow(r.Context(), escrowsvc.EnsureEscrowRequest{
		TenantID:         actor.TenantID,
		ListingID:        listing.ListingID,
		FundingAccountID: listing.FundingAccountID,
		Currency:         listing.Currency,
		Metadata:         map[string]any{"source": "bid_accept", "bid_id": bid.BidID},
	})
	if err != nil {
		s.writeEscrowError(w, corrID, "escrow_ensure_failed", "Failed to ensure escrow for bid acceptance.", err)
		return
	}
	escrowRecord, transfer, err := s.escrow.Lock(r.Context(), escrowsvc.AdjustEscrowRequest{
		TenantID:     actor.TenantID,
		EscrowID:     escrowRecord.EscrowID,
		Amount:       bid.Amount,
		TransferType: escrowsvc.TransferLock,
		Metadata:     map[string]any{"source": "bid_accept", "bid_id": bid.BidID},
	})
	if err != nil {
		s.writeEscrowError(w, corrID, "escrow_lock_failed", "Failed to lock escrow after bid acceptance.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"bid":            bid,
		"escrow":         escrowRecord,
		"transfer":       transfer,
		"correlation_id": corrID,
	})
}

func (s *Server) handleFundMarketListingEscrowV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "requester") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for listing funding.", corrID, nil)
		return
	}

	var req fundMarketListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	listing, err := s.marketplace.GetListing(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("listing_id")))
	if err != nil {
		s.writeMarketplaceError(w, corrID, "market_listing_not_found", "Market listing not found.", err)
		return
	}
	if listing.Status != marketplacesvc.ListingStatusPublished {
		apierrors.Write(w, http.StatusConflict, "listing_not_published", "Listing must be published before funding escrow.", corrID, nil)
		return
	}
	amount := req.Amount
	if amount <= 0 {
		amount = listing.BudgetTotal
	}

	escrowRecord, err := s.escrow.EnsureEscrow(r.Context(), escrowsvc.EnsureEscrowRequest{
		TenantID:         actor.TenantID,
		EscrowID:         strings.TrimSpace(req.EscrowID),
		ListingID:        listing.ListingID,
		FundingAccountID: listing.FundingAccountID,
		Currency:         listing.Currency,
		Metadata:         map[string]any{"source": "listing_fund"},
	})
	if err != nil {
		s.writeEscrowError(w, corrID, "escrow_ensure_failed", "Failed to ensure escrow.", err)
		return
	}
	escrowRecord, transfer, err := s.escrow.Lock(r.Context(), escrowsvc.AdjustEscrowRequest{
		TenantID:     actor.TenantID,
		EscrowID:     escrowRecord.EscrowID,
		Amount:       amount,
		TransferType: escrowsvc.TransferLock,
		Metadata:     req.Metadata,
	})
	if err != nil {
		s.writeEscrowError(w, corrID, "escrow_lock_failed", "Failed to fund escrow.", err)
		return
	}

	s.appendEvent(r, actor.TenantID, "escrow", escrowRecord.EscrowID, "escrow.funded", map[string]any{"listing_id": listing.ListingID, "amount": amount})
	writeJSON(w, http.StatusOK, map[string]any{
		"escrow":         escrowRecord,
		"transfer":       transfer,
		"correlation_id": corrID,
	})
}

func (s *Server) handleGetMarketEscrowV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "approver", "auditor", "requester", "executor") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for escrow read.", corrID, nil)
		return
	}

	escrowRecord, err := s.escrow.GetEscrow(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("escrow_id")))
	if err != nil {
		s.writeEscrowError(w, corrID, "escrow_not_found", "Escrow not found.", err)
		return
	}
	transfers, err := s.escrow.ListTransfers(r.Context(), actor.TenantID, escrowRecord.EscrowID)
	if err != nil {
		s.writeEscrowError(w, corrID, "escrow_transfer_list_failed", "Failed to list escrow transfers.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"escrow":         escrowRecord,
		"transfers":      transfers,
		"correlation_id": corrID,
	})
}

func (s *Server) handleReleaseMarketEscrowV2(w http.ResponseWriter, r *http.Request) {
	s.handleAdjustEscrowV2(w, r, escrowsvc.TransferRelease)
}

func (s *Server) handleRefundMarketEscrowV2(w http.ResponseWriter, r *http.Request) {
	s.handleAdjustEscrowV2(w, r, escrowsvc.TransferRefund)
}

func (s *Server) handleAdjustEscrowV2(w http.ResponseWriter, r *http.Request, transferType string) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "approver", "arbitrator", "requester") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for escrow update.", corrID, nil)
		return
	}

	var req adjustEscrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	escrowID := strings.TrimSpace(r.PathValue("escrow_id"))
	escrowRecord, err := s.escrow.GetEscrow(r.Context(), actor.TenantID, escrowID)
	if err != nil {
		s.writeEscrowError(w, corrID, "escrow_not_found", "Escrow not found.", err)
		return
	}
	remaining := escrowRecord.TotalLocked - (escrowRecord.ReleasedAmount + escrowRecord.RefundedAmount)
	amount := req.Amount
	if amount <= 0 {
		amount = remaining
	}

	adjustReq := escrowsvc.AdjustEscrowRequest{
		TenantID:     actor.TenantID,
		EscrowID:     escrowID,
		Amount:       amount,
		TransferType: transferType,
		Metadata: map[string]any{
			"reason": req.Reason,
		},
	}

	var transfer escrowsvc.Transfer
	if transferType == escrowsvc.TransferRelease {
		escrowRecord, transfer, err = s.escrow.Release(r.Context(), adjustReq)
	} else {
		escrowRecord, transfer, err = s.escrow.Refund(r.Context(), adjustReq)
	}
	if err != nil {
		s.writeEscrowError(w, corrID, "escrow_adjust_failed", "Failed to adjust escrow.", err)
		return
	}

	eventType := "escrow.released"
	if transferType == escrowsvc.TransferRefund {
		eventType = "escrow.refunded"
	}
	s.appendEvent(r, actor.TenantID, "escrow", escrowRecord.EscrowID, eventType, map[string]any{"amount": transfer.Amount, "reason": req.Reason})

	writeJSON(w, http.StatusOK, map[string]any{
		"escrow":         escrowRecord,
		"transfer":       transfer,
		"correlation_id": corrID,
	})
}

func (s *Server) handleCreatePayoutAccountV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "executor") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for payout account create.", corrID, nil)
		return
	}

	var req createPayoutAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	accountID := strings.TrimSpace(req.PayoutAccountID)
	if accountID == "" {
		accountID = "payoutacct_" + randomID(6)
	}
	account, err := s.payouts.CreateAccount(r.Context(), payoutssvc.CreateAccountRequest{
		TenantID:           actor.TenantID,
		PayoutAccountID:    accountID,
		OwnerProfileID:     strings.TrimSpace(req.OwnerProfileID),
		Provider:           strings.TrimSpace(req.Provider),
		ProviderAccountRef: strings.TrimSpace(req.ProviderAccountRef),
		Jurisdiction:       strings.TrimSpace(req.Jurisdiction),
		Status:             strings.TrimSpace(req.Status),
		Metadata:           req.Metadata,
	})
	if err != nil {
		s.writePayoutError(w, corrID, "payout_account_create_failed", "Failed to create payout account.", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"payout_account": account,
		"correlation_id": corrID,
	})
}

func (s *Server) handleGetPayoutAccountV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "approver", "auditor", "requester", "executor") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for payout account read.", corrID, nil)
		return
	}
	account, err := s.payouts.GetAccount(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("payout_account_id")))
	if err != nil {
		s.writePayoutError(w, corrID, "payout_account_not_found", "Payout account not found.", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"payout_account": account,
		"correlation_id": corrID,
	})
}

func (s *Server) handleCreatePayoutV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "requester", "executor", "finance_viewer") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for payout request.", corrID, nil)
		return
	}

	var req createPayoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	if strings.TrimSpace(req.EscrowID) != "" {
		if _, err := s.escrow.GetEscrow(r.Context(), actor.TenantID, strings.TrimSpace(req.EscrowID)); err != nil {
			s.writeEscrowError(w, corrID, "escrow_not_found", "Escrow not found.", err)
			return
		}
	}
	payoutID := strings.TrimSpace(req.PayoutID)
	if payoutID == "" {
		payoutID = "payout_" + randomID(6)
	}
	payout, err := s.payouts.RequestPayout(r.Context(), payoutssvc.RequestPayoutRequest{
		TenantID:        actor.TenantID,
		PayoutID:        payoutID,
		PayoutAccountID: strings.TrimSpace(req.PayoutAccountID),
		EscrowID:        strings.TrimSpace(req.EscrowID),
		Amount:          req.Amount,
		Currency:        strings.TrimSpace(req.Currency),
		AutoComplete:    req.AutoComplete,
		Metadata:        req.Metadata,
	})
	if err != nil {
		s.writePayoutError(w, corrID, "payout_request_failed", "Failed to request payout.", err)
		return
	}
	eventType := "payout.requested"
	if payout.Status == payoutssvc.PayoutStatusCompleted {
		eventType = "payout.completed"
	}
	s.appendEvent(r, actor.TenantID, "payout", payout.PayoutID, eventType, map[string]any{"amount": payout.Amount, "escrow_id": payout.EscrowID})
	writeJSON(w, http.StatusOK, map[string]any{
		"payout":         payout,
		"correlation_id": corrID,
	})
}

func (s *Server) handleGetPayoutV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "approver", "auditor", "requester", "executor", "finance_viewer") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for payout read.", corrID, nil)
		return
	}
	payout, err := s.payouts.GetPayout(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("payout_id")))
	if err != nil {
		s.writePayoutError(w, corrID, "payout_not_found", "Payout not found.", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"payout":         payout,
		"correlation_id": corrID,
	})
}

func (s *Server) handleOpenMarketDisputeV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "requester", "executor", "approver") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for dispute open.", corrID, nil)
		return
	}

	var req openMarketDisputeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	disputeID := strings.TrimSpace(req.DisputeID)
	if disputeID == "" {
		disputeID = "disp_" + randomID(6)
	}
	dispute, err := s.disputes.OpenDispute(r.Context(), disputessvc.OpenDisputeRequest{
		TenantID:    actor.TenantID,
		DisputeID:   disputeID,
		ListingID:   strings.TrimSpace(req.ListingID),
		EscrowID:    strings.TrimSpace(req.EscrowID),
		ClaimID:     strings.TrimSpace(req.ClaimID),
		DisputeType: strings.TrimSpace(req.DisputeType),
		Reason:      strings.TrimSpace(req.Reason),
		OpenedBy:    actor.ID,
	})
	if err != nil {
		s.writeDisputeError(w, corrID, "dispute_open_failed", "Failed to open dispute.", err)
		return
	}
	s.appendEvent(r, actor.TenantID, "dispute", dispute.DisputeID, "dispute.opened", map[string]any{"listing_id": dispute.ListingID, "dispute_type": dispute.DisputeType})
	writeJSON(w, http.StatusOK, map[string]any{
		"dispute":        dispute,
		"correlation_id": corrID,
	})
}

func (s *Server) handleGetMarketDisputeV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "requester", "executor", "approver", "auditor") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for dispute read.", corrID, nil)
		return
	}
	dispute, err := s.disputes.GetDispute(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("dispute_id")))
	if err != nil {
		s.writeDisputeError(w, corrID, "dispute_not_found", "Dispute not found.", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"dispute":        dispute,
		"correlation_id": corrID,
	})
}

func (s *Server) handleResolveMarketDisputeV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "approver", "arbitrator") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for dispute resolve.", corrID, nil)
		return
	}

	var req resolveMarketDisputeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	dispute, err := s.disputes.ResolveDispute(r.Context(), disputessvc.ResolveDisputeRequest{
		TenantID:             actor.TenantID,
		DisputeID:            strings.TrimSpace(r.PathValue("dispute_id")),
		Decision:             strings.TrimSpace(req.Decision),
		FeeShift:             req.FeeShift,
		EscrowReleaseAction:  strings.TrimSpace(req.EscrowReleaseAction),
		ReputationAdjustment: req.ReputationAdjustment,
		AppealAllowed:        req.AppealAllowed,
	})
	if err != nil {
		s.writeDisputeError(w, corrID, "dispute_resolve_failed", "Failed to resolve dispute.", err)
		return
	}

	var escrowRecord any
	var transfer any
	if dispute.EscrowID != "" {
		currentEscrow, err := s.escrow.GetEscrow(r.Context(), actor.TenantID, dispute.EscrowID)
		if err == nil {
			remaining := currentEscrow.TotalLocked - (currentEscrow.ReleasedAmount + currentEscrow.RefundedAmount)
			amount := req.EscrowAmount
			if amount <= 0 {
				amount = remaining
			}
			switch dispute.EscrowReleaseAction {
			case disputessvc.EscrowActionRelease:
				currentEscrow, t, err := s.escrow.Release(r.Context(), escrowsvc.AdjustEscrowRequest{
					TenantID:     actor.TenantID,
					EscrowID:     dispute.EscrowID,
					Amount:       amount,
					TransferType: escrowsvc.TransferRelease,
					Metadata:     map[string]any{"source": "dispute_resolve"},
				})
				if err == nil {
					escrowRecord = currentEscrow
					transfer = t
				}
			case disputessvc.EscrowActionRefund:
				currentEscrow, t, err := s.escrow.Refund(r.Context(), escrowsvc.AdjustEscrowRequest{
					TenantID:     actor.TenantID,
					EscrowID:     dispute.EscrowID,
					Amount:       amount,
					TransferType: escrowsvc.TransferRefund,
					Metadata:     map[string]any{"source": "dispute_resolve"},
				})
				if err == nil {
					escrowRecord = currentEscrow
					transfer = t
				}
			}
		}
	}

	s.appendEvent(r, actor.TenantID, "dispute", dispute.DisputeID, "dispute.resolved", map[string]any{
		"decision":              dispute.Decision,
		"escrow_release_action": dispute.EscrowReleaseAction,
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"dispute":        dispute,
		"escrow":         escrowRecord,
		"transfer":       transfer,
		"correlation_id": corrID,
	})
}

func (s *Server) handleAppealMarketDisputeV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "requester", "executor", "approver") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for dispute appeal.", corrID, nil)
		return
	}

	var req appealMarketDisputeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	dispute, err := s.disputes.AppealDispute(r.Context(), disputessvc.AppealDisputeRequest{
		TenantID:  actor.TenantID,
		DisputeID: strings.TrimSpace(r.PathValue("dispute_id")),
		AppealBy:  actor.ID,
		Reason:    strings.TrimSpace(req.Reason),
	})
	if err != nil {
		s.writeDisputeError(w, corrID, "dispute_appeal_failed", "Failed to appeal dispute.", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"dispute":        dispute,
		"correlation_id": corrID,
	})
}

func (s *Server) handleMarketDashboardV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "approver", "auditor", "requester", "executor") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for market dashboard read.", corrID, nil)
		return
	}

	listingsDraft := s.marketplace.ListListings(r.Context(), actor.TenantID, marketplacesvc.ListListingsOptions{Status: marketplacesvc.ListingStatusDraft, Limit: 1000})
	listingsPublished := s.marketplace.ListListings(r.Context(), actor.TenantID, marketplacesvc.ListListingsOptions{Status: marketplacesvc.ListingStatusPublished, Limit: 1000})
	listingsCancelled := s.marketplace.ListListings(r.Context(), actor.TenantID, marketplacesvc.ListListingsOptions{Status: marketplacesvc.ListingStatusCancelled, Limit: 1000})
	claims := s.marketplace.ListClaims(r.Context(), actor.TenantID, "")
	bids := s.marketplace.ListBids(r.Context(), actor.TenantID, "")
	escrows := s.escrow.ListEscrows(r.Context(), actor.TenantID)
	disputes := s.disputes.ListDisputes(r.Context(), actor.TenantID)
	payoutAccounts := s.payouts.ListAccounts(r.Context(), actor.TenantID)
	payouts := s.payouts.ListPayouts(r.Context(), actor.TenantID)

	claimSubmitted := 0
	claimAwarded := 0
	claimWithdrawn := 0
	for _, claim := range claims {
		switch claim.Status {
		case marketplacesvc.ClaimStatusSubmitted:
			claimSubmitted++
		case marketplacesvc.ClaimStatusAwarded:
			claimAwarded++
		case marketplacesvc.ClaimStatusWithdrawn:
			claimWithdrawn++
		}
	}

	bidSubmitted := 0
	bidAccepted := 0
	for _, bid := range bids {
		switch bid.Status {
		case marketplacesvc.BidStatusSubmitted:
			bidSubmitted++
		case marketplacesvc.BidStatusAccepted:
			bidAccepted++
		}
	}

	disputeOpen := 0
	disputeResolved := 0
	for _, dispute := range disputes {
		switch dispute.Status {
		case disputessvc.StatusOpen:
			disputeOpen++
		case disputessvc.StatusResolved, disputessvc.StatusAppealed:
			disputeResolved++
		}
	}

	payoutCompleted := 0
	payoutFailed := 0
	for _, payout := range payouts {
		switch payout.Status {
		case payoutssvc.PayoutStatusCompleted:
			payoutCompleted++
		case payoutssvc.PayoutStatusFailed:
			payoutFailed++
		}
	}

	claimAbandonmentRate := 0.0
	if len(claims) > 0 {
		claimAbandonmentRate = float64(claimWithdrawn) / float64(len(claims))
	}

	requesterDefaultRate := 0.0
	if len(disputes) > 0 {
		defaultCount := 0
		for _, dispute := range disputes {
			if dispute.DisputeType == disputessvc.TypeRequesterDefault {
				defaultCount++
			}
		}
		requesterDefaultRate = float64(defaultCount) / float64(len(disputes))
	}
	rolloutGate := s.evaluateMarketRolloutGate(r, actor.TenantID, disputes, payoutAccounts, payouts)

	writeJSON(w, http.StatusOK, map[string]any{
		"market_dashboard": map[string]any{
			"listings": map[string]any{
				"draft":     len(listingsDraft),
				"published": len(listingsPublished),
				"cancelled": len(listingsCancelled),
			},
			"claims": map[string]any{
				"submitted": claimSubmitted,
				"awarded":   claimAwarded,
				"withdrawn": claimWithdrawn,
				"total":     len(claims),
			},
			"bids": map[string]any{
				"submitted": bidSubmitted,
				"accepted":  bidAccepted,
				"total":     len(bids),
			},
			"escrows": map[string]any{
				"total": len(escrows),
			},
			"disputes": map[string]any{
				"open":     disputeOpen,
				"resolved": disputeResolved,
				"total":    len(disputes),
			},
			"payouts": map[string]any{
				"completed": payoutCompleted,
				"failed":    payoutFailed,
				"total":     len(payouts),
			},
			"payout_accounts": map[string]any{
				"total": len(payoutAccounts),
			},
			"risk_metrics": map[string]any{
				"claim_abandonment_rate": claimAbandonmentRate,
				"requester_default_rate": requesterDefaultRate,
			},
			"rollout_gate": rolloutGate,
		},
		"correlation_id": corrID,
	})
}

func (s *Server) evaluateMarketRolloutGate(
	r *http.Request,
	tenantID string,
	disputes []disputessvc.Dispute,
	payoutAccounts []payoutssvc.Account,
	payouts []payoutssvc.Payout,
) map[string]any {
	simulationReady := true
	disputeReady := false
	payoutReady := len(payoutAccounts) > 0 && len(payouts) > 0
	blockingReasons := make([]string, 0)
	requiredScenarioStatus := make([]map[string]any, 0, len(marketRolloutScenarioIDs))

	if err := s.simulation.EnsureV1ScenarioLibrary(r.Context(), tenantID); err != nil {
		simulationReady = false
		blockingReasons = append(blockingReasons, "simulation_seed_failed")
	}

	baselineByScenario := map[string]simulationsvc.Baseline{}
	for _, baseline := range s.simulation.ListBaselines(r.Context(), tenantID, "") {
		existing, ok := baselineByScenario[baseline.ScenarioID]
		if !ok || baseline.CreatedAt.After(existing.CreatedAt) {
			baselineByScenario[baseline.ScenarioID] = baseline
		}
	}

	for _, scenarioID := range marketRolloutScenarioIDs {
		status := map[string]any{
			"scenario_id": scenarioID,
			"ready":       false,
		}
		baseline, ok := baselineByScenario[scenarioID]
		if !ok {
			simulationReady = false
			status["reason"] = "baseline_missing"
			requiredScenarioStatus = append(requiredScenarioStatus, status)
			blockingReasons = append(blockingReasons, "simulation_baseline_missing:"+scenarioID)
			continue
		}

		compare, err := s.simulation.Compare(r.Context(), simulationsvc.CompareRequest{
			TenantID:         tenantID,
			CandidateRunID:   baseline.RunID,
			BaselineID:       baseline.BaselineID,
			FailOnSeverities: []string{"critical"},
		})
		if err != nil {
			simulationReady = false
			status["baseline_id"] = baseline.BaselineID
			status["reason"] = "compare_failed"
			requiredScenarioStatus = append(requiredScenarioStatus, status)
			blockingReasons = append(blockingReasons, "simulation_compare_failed:"+scenarioID)
			continue
		}

		status["baseline_id"] = baseline.BaselineID
		status["candidate_run_id"] = compare.CandidateRunID
		status["verdict"] = compare.Verdict
		status["reasons"] = compare.Reasons
		if compare.Verdict == "pass" {
			status["ready"] = true
		} else {
			simulationReady = false
			blockingReasons = append(blockingReasons, "simulation_compare_not_pass:"+scenarioID)
		}
		requiredScenarioStatus = append(requiredScenarioStatus, status)
	}

	resolvedDisputes := 0
	for _, dispute := range disputes {
		if dispute.Status == disputessvc.StatusResolved || dispute.Status == disputessvc.StatusAppealed {
			resolvedDisputes++
		}
	}
	disputeReady = resolvedDisputes > 0
	if !disputeReady {
		blockingReasons = append(blockingReasons, "dispute_resolution_readiness_missing")
	}
	if !payoutReady {
		blockingReasons = append(blockingReasons, "payout_readiness_missing")
	}

	ready := simulationReady && disputeReady && payoutReady
	return map[string]any{
		"ready":              ready,
		"simulation_ready":   simulationReady,
		"dispute_ready":      disputeReady,
		"payout_ready":       payoutReady,
		"required_scenarios": requiredScenarioStatus,
		"blocking_reasons":   blockingReasons,
		"resolved_disputes":  resolvedDisputes,
	}
}

func (s *Server) writeMarketplaceError(w http.ResponseWriter, corrID, code, message string, err error) {
	switch {
	case errors.Is(err, marketplacesvc.ErrNotFound):
		apierrors.Write(w, http.StatusNotFound, code, message, corrID, nil)
	case errors.Is(err, marketplacesvc.ErrAlreadyExists):
		apierrors.Write(w, http.StatusConflict, code, message, corrID, nil)
	case errors.Is(err, marketplacesvc.ErrClaimHoarding):
		apierrors.Write(w, http.StatusConflict, "claim_hoarding_limit", "Executor claim concurrency cap reached.", corrID, nil)
	case errors.Is(err, marketplacesvc.ErrAntiSpamControl):
		apierrors.Write(w, http.StatusTooManyRequests, "market_spam_control", err.Error(), corrID, nil)
	case errors.Is(err, marketplacesvc.ErrSybilControl):
		apierrors.Write(w, http.StatusForbidden, "market_identity_verification_required", err.Error(), corrID, nil)
	case errors.Is(err, marketplacesvc.ErrInsufficientFunds):
		apierrors.Write(w, http.StatusConflict, "insufficient_funded_reserve", "Listing publication requires sufficient funded reserve.", corrID, nil)
	case errors.Is(err, marketplacesvc.ErrInvalidRequest):
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", err.Error(), corrID, nil)
	default:
		apierrors.Write(w, http.StatusInternalServerError, code, message, corrID, map[string]any{"reason": err.Error()})
	}
}

func (s *Server) writeEscrowError(w http.ResponseWriter, corrID, code, message string, err error) {
	switch {
	case errors.Is(err, escrowsvc.ErrNotFound):
		apierrors.Write(w, http.StatusNotFound, code, message, corrID, nil)
	case errors.Is(err, escrowsvc.ErrAlreadyExists):
		apierrors.Write(w, http.StatusConflict, code, message, corrID, nil)
	case errors.Is(err, escrowsvc.ErrInvalidRequest):
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", err.Error(), corrID, nil)
	default:
		apierrors.Write(w, http.StatusInternalServerError, code, message, corrID, map[string]any{"reason": err.Error()})
	}
}

func (s *Server) writePayoutError(w http.ResponseWriter, corrID, code, message string, err error) {
	switch {
	case errors.Is(err, payoutssvc.ErrNotFound):
		apierrors.Write(w, http.StatusNotFound, code, message, corrID, nil)
	case errors.Is(err, payoutssvc.ErrAlreadyExists):
		apierrors.Write(w, http.StatusConflict, code, message, corrID, nil)
	case errors.Is(err, payoutssvc.ErrInvalidRequest):
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", err.Error(), corrID, nil)
	default:
		apierrors.Write(w, http.StatusInternalServerError, code, message, corrID, map[string]any{"reason": err.Error()})
	}
}

func (s *Server) writeDisputeError(w http.ResponseWriter, corrID, code, message string, err error) {
	switch {
	case errors.Is(err, disputessvc.ErrNotFound):
		apierrors.Write(w, http.StatusNotFound, code, message, corrID, nil)
	case errors.Is(err, disputessvc.ErrAlreadyExists):
		apierrors.Write(w, http.StatusConflict, code, message, corrID, nil)
	case errors.Is(err, disputessvc.ErrInvalidRequest):
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", err.Error(), corrID, nil)
	default:
		apierrors.Write(w, http.StatusInternalServerError, code, message, corrID, map[string]any{"reason": err.Error()})
	}
}
