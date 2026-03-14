package marketplace

import (
	"context"
	"errors"
	"testing"
)

func TestMarketProfileListingFundingFlow(t *testing.T) {
	svc := New()
	ctx := context.Background()

	profile, err := svc.CreateProfile(ctx, CreateProfileRequest{
		TenantID:    "ten_01",
		ProfileID:   "mp_req_01",
		ProfileType: ProfileTypeRequester,
		DisplayName: "Requester One",
	})
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}
	if _, err := svc.GetReputation(ctx, "ten_01", profile.ProfileID); err != nil {
		t.Fatalf("get reputation: %v", err)
	}

	account, err := svc.CreateFundingAccount(ctx, CreateFundingAccountRequest{
		TenantID:       "ten_01",
		AccountID:      "fund_01",
		OwnerProfileID: profile.ProfileID,
		Currency:       "JWUSD",
		InitialBalance: 200,
	})
	if err != nil {
		t.Fatalf("create funding account: %v", err)
	}

	listing, err := svc.CreateListing(ctx, CreateListingRequest{
		TenantID:           "ten_01",
		ListingID:          "ml_01",
		TaskID:             "task_01",
		RequesterProfileID: profile.ProfileID,
		WorkClass:          WorkClassPublicOpen,
		ListingMode:        ListingModeFixedOpenClaim,
		BudgetTotal:        120,
		Currency:           "JWUSD",
		FundingAccountID:   account.AccountID,
	})
	if err != nil {
		t.Fatalf("create listing: %v", err)
	}

	published, err := svc.PublishListing(ctx, PublishListingRequest{
		TenantID:  "ten_01",
		ListingID: listing.ListingID,
		Reason:    "ready",
	})
	if err != nil {
		t.Fatalf("publish listing: %v", err)
	}
	if published.Status != ListingStatusPublished {
		t.Fatalf("listing status=%q want %q", published.Status, ListingStatusPublished)
	}

	acctAfterPublish, err := svc.GetFundingAccount(ctx, "ten_01", account.AccountID)
	if err != nil {
		t.Fatalf("get funding account after publish: %v", err)
	}
	if acctAfterPublish.AvailableBalance != 80 || acctAfterPublish.ReservedBalance != 120 {
		t.Fatalf("unexpected account balances after publish: %+v", acctAfterPublish)
	}

	cancelled, err := svc.CancelListing(ctx, CancelListingRequest{
		TenantID:  "ten_01",
		ListingID: listing.ListingID,
		Reason:    "cancelled",
	})
	if err != nil {
		t.Fatalf("cancel listing: %v", err)
	}
	if cancelled.Status != ListingStatusCancelled {
		t.Fatalf("listing status=%q want %q", cancelled.Status, ListingStatusCancelled)
	}

	acctAfterCancel, err := svc.GetFundingAccount(ctx, "ten_01", account.AccountID)
	if err != nil {
		t.Fatalf("get funding account after cancel: %v", err)
	}
	if acctAfterCancel.AvailableBalance != 200 || acctAfterCancel.ReservedBalance != 0 {
		t.Fatalf("unexpected account balances after cancel: %+v", acctAfterCancel)
	}
}

func TestMarketGuardrails(t *testing.T) {
	svc := New()
	ctx := context.Background()

	if _, err := svc.CreateProfile(ctx, CreateProfileRequest{
		TenantID:    "ten_01",
		ProfileID:   "mp_bad",
		ProfileType: "bad_type",
		DisplayName: "Bad",
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid profile type, got %v", err)
	}

	profile, err := svc.CreateProfile(ctx, CreateProfileRequest{
		TenantID:    "ten_01",
		ProfileID:   "mp_req_02",
		ProfileType: ProfileTypeRequester,
		DisplayName: "Requester Two",
	})
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}
	_, _ = svc.CreateProfile(ctx, CreateProfileRequest{
		TenantID:    "ten_01",
		ProfileID:   "mp_exec_01",
		ProfileType: ProfileTypeExecutor,
		DisplayName: "Executor One",
	})

	if _, err := svc.CreateFundingAccount(ctx, CreateFundingAccountRequest{
		TenantID:       "ten_01",
		AccountID:      "fund_02",
		OwnerProfileID: profile.ProfileID,
		Currency:       "JWUSD",
		InitialBalance: 50,
	}); err != nil {
		t.Fatalf("create funding account: %v", err)
	}

	if _, err := svc.CreateListing(ctx, CreateListingRequest{
		TenantID:           "ten_01",
		ListingID:          "ml_private",
		TaskID:             "task_private",
		RequesterProfileID: profile.ProfileID,
		WorkClass:          WorkClassPrivateTenant,
		ListingMode:        ListingModeFixedOpenClaim,
		BudgetTotal:        20,
		Currency:           "JWUSD",
		FundingAccountID:   "fund_02",
	}); err != nil {
		t.Fatalf("create private listing: %v", err)
	}

	if _, err := svc.PublishListing(ctx, PublishListingRequest{
		TenantID:  "ten_01",
		ListingID: "ml_private",
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected private work-class publish guardrail, got %v", err)
	}

	if _, err := svc.CreateListing(ctx, CreateListingRequest{
		TenantID:           "ten_01",
		ListingID:          "ml_unfunded",
		TaskID:             "task_unfunded",
		RequesterProfileID: profile.ProfileID,
		WorkClass:          WorkClassPublicOpen,
		ListingMode:        ListingModeFixedOpenClaim,
		BudgetTotal:        80,
		Currency:           "JWUSD",
		FundingAccountID:   "fund_02",
	}); err != nil {
		t.Fatalf("create unfunded listing: %v", err)
	}
	if _, err := svc.PublishListing(ctx, PublishListingRequest{
		TenantID:  "ten_01",
		ListingID: "ml_unfunded",
	}); !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("expected insufficient funds error, got %v", err)
	}

	executorListing, err := svc.CreateListing(ctx, CreateListingRequest{
		TenantID:           "ten_01",
		ListingID:          "ml_bad_profile",
		TaskID:             "task_bad_profile",
		RequesterProfileID: "mp_exec_01",
		WorkClass:          WorkClassPublicOpen,
		ListingMode:        ListingModeFixedOpenClaim,
		BudgetTotal:        10,
		Currency:           "JWUSD",
		FundingAccountID:   "fund_02",
	})
	if err == nil || executorListing.ListingID != "" {
		t.Fatalf("expected invalid requester profile type error, got listing=%+v err=%v", executorListing, err)
	}
}

func TestMarketPatchListAndErrorPaths(t *testing.T) {
	svc := New()
	ctx := context.Background()

	if _, err := svc.GetProfile(ctx, "ten_01", "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing profile error, got %v", err)
	}
	if _, err := svc.CreateProfile(ctx, CreateProfileRequest{
		TenantID:           "ten_01",
		ProfileID:          "mp_01",
		ProfileType:        ProfileTypeHybrid,
		DisplayName:        "Hybrid One",
		VerificationStatus: "bad_status",
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid verification status, got %v", err)
	}

	profile, err := svc.CreateProfile(ctx, CreateProfileRequest{
		TenantID:    "ten_01",
		ProfileID:   "mp_02",
		ProfileType: ProfileTypeRequester,
		DisplayName: "Requester Patch",
	})
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}

	display := "Requester Patched"
	verified := VerificationVerified
	workClasses := []string{WorkClassPublicOpen, WorkClassRestrictedMarket}
	updated, err := svc.PatchProfile(ctx, PatchProfileRequest{
		TenantID:           "ten_01",
		ProfileID:          profile.ProfileID,
		DisplayName:        &display,
		VerificationStatus: &verified,
		WorkClassAllowlist: &workClasses,
	})
	if err != nil {
		t.Fatalf("patch profile: %v", err)
	}
	if updated.DisplayName != display || updated.VerificationStatus != verified {
		t.Fatalf("unexpected patched profile: %+v", updated)
	}

	invalidWorkClass := []string{"bad"}
	if _, err := svc.PatchProfile(ctx, PatchProfileRequest{
		TenantID:           "ten_01",
		ProfileID:          profile.ProfileID,
		WorkClassAllowlist: &invalidWorkClass,
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid work class allowlist error, got %v", err)
	}

	if _, err := svc.CreateFundingAccount(ctx, CreateFundingAccountRequest{
		TenantID:       "ten_01",
		AccountID:      "fund_03",
		OwnerProfileID: profile.ProfileID,
		Currency:       "JWUSD",
		InitialBalance: 100,
	}); err != nil {
		t.Fatalf("create funding account: %v", err)
	}
	if _, err := svc.CreateFundingAccount(ctx, CreateFundingAccountRequest{
		TenantID:       "ten_01",
		AccountID:      "fund_missing_profile",
		OwnerProfileID: "missing",
		Currency:       "JWUSD",
		InitialBalance: 100,
	}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing profile error, got %v", err)
	}

	listing, err := svc.CreateListing(ctx, CreateListingRequest{
		TenantID:           "ten_01",
		ListingID:          "ml_03",
		TaskID:             "task_03",
		RequesterProfileID: profile.ProfileID,
		WorkClass:          WorkClassPublicOpen,
		ListingMode:        ListingModeFixedOpenClaim,
		BudgetTotal:        30,
		Currency:           "JWUSD",
		FundingAccountID:   "fund_03",
		ContractSnapshot:   map[string]any{"v": 1},
	})
	if err != nil {
		t.Fatalf("create listing: %v", err)
	}
	if _, err := svc.GetListing(ctx, "ten_01", listing.ListingID); err != nil {
		t.Fatalf("get listing: %v", err)
	}

	if _, err := svc.CreateListing(ctx, CreateListingRequest{
		TenantID:           "ten_01",
		ListingID:          "ml_bad_mode",
		TaskID:             "task_04",
		RequesterProfileID: profile.ProfileID,
		WorkClass:          WorkClassPublicOpen,
		ListingMode:        "bad_mode",
		BudgetTotal:        10,
		Currency:           "JWUSD",
		FundingAccountID:   "fund_03",
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid listing mode error, got %v", err)
	}

	if _, err := svc.CreateListing(ctx, CreateListingRequest{
		TenantID:           "ten_01",
		ListingID:          "ml_bad_currency",
		TaskID:             "task_05",
		RequesterProfileID: profile.ProfileID,
		WorkClass:          WorkClassPublicOpen,
		ListingMode:        ListingModeFixedOpenClaim,
		BudgetTotal:        10,
		Currency:           "BAD",
		FundingAccountID:   "fund_03",
	}); err != nil {
		t.Fatalf("create listing for publish mismatch branch: %v", err)
	}
	if _, err := svc.PublishListing(ctx, PublishListingRequest{
		TenantID:  "ten_01",
		ListingID: "ml_bad_currency",
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected currency mismatch error, got %v", err)
	}

	if _, err := svc.PublishListing(ctx, PublishListingRequest{
		TenantID:  "ten_01",
		ListingID: listing.ListingID,
	}); err != nil {
		t.Fatalf("publish listing: %v", err)
	}
	if _, err := svc.PublishListing(ctx, PublishListingRequest{
		TenantID:  "ten_01",
		ListingID: listing.ListingID,
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected publish on non-draft error, got %v", err)
	}

	published := svc.ListListings(ctx, "ten_01", ListListingsOptions{Status: ListingStatusPublished, Limit: 10})
	if len(published) == 0 {
		t.Fatalf("expected published listings in feed")
	}
	if len(svc.ListListings(ctx, "ten_01", ListListingsOptions{Status: ListingStatusPublished, WorkClass: WorkClassRestrictedMarket, Limit: 10})) != 0 {
		t.Fatalf("expected no restricted listings in filtered feed")
	}

	cancelled, err := svc.CancelListing(ctx, CancelListingRequest{
		TenantID:  "ten_01",
		ListingID: listing.ListingID,
		Reason:    "done",
	})
	if err != nil {
		t.Fatalf("cancel listing: %v", err)
	}
	if cancelled.Status != ListingStatusCancelled {
		t.Fatalf("expected cancelled status, got %+v", cancelled)
	}
	if _, err := svc.CancelListing(ctx, CancelListingRequest{
		TenantID:  "ten_01",
		ListingID: listing.ListingID,
		Reason:    "already",
	}); err != nil {
		t.Fatalf("cancel already cancelled listing should be idempotent: %v", err)
	}
	if _, err := svc.GetListing(ctx, "ten_01", "missing_listing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing listing error, got %v", err)
	}
}

func TestMarketClaimsAndBids(t *testing.T) {
	svc := New()
	ctx := context.Background()

	requester, err := svc.CreateProfile(ctx, CreateProfileRequest{
		TenantID:    "ten_01",
		ProfileID:   "mp_req_claims",
		ProfileType: ProfileTypeRequester,
		DisplayName: "Requester Claims",
	})
	if err != nil {
		t.Fatalf("create requester profile: %v", err)
	}
	executor, err := svc.CreateProfile(ctx, CreateProfileRequest{
		TenantID:           "ten_01",
		ProfileID:          "mp_exec_claims",
		ProfileType:        ProfileTypeExecutor,
		DisplayName:        "Executor Claims",
		VerificationStatus: VerificationVerified,
	})
	if err != nil {
		t.Fatalf("create executor profile: %v", err)
	}
	if _, err := svc.CreateFundingAccount(ctx, CreateFundingAccountRequest{
		TenantID:       "ten_01",
		AccountID:      "fund_claims",
		OwnerProfileID: requester.ProfileID,
		Currency:       "JWUSD",
		InitialBalance: 200,
	}); err != nil {
		t.Fatalf("create funding account: %v", err)
	}
	listing, err := svc.CreateListing(ctx, CreateListingRequest{
		TenantID:           "ten_01",
		ListingID:          "ml_claims",
		TaskID:             "task_claims",
		RequesterProfileID: requester.ProfileID,
		WorkClass:          WorkClassPublicSealed,
		ListingMode:        ListingModeFixedOpenClaim,
		BudgetTotal:        60,
		Currency:           "JWUSD",
		FundingAccountID:   "fund_claims",
	})
	if err != nil {
		t.Fatalf("create listing: %v", err)
	}
	if _, err := svc.PublishListing(ctx, PublishListingRequest{
		TenantID:  "ten_01",
		ListingID: listing.ListingID,
	}); err != nil {
		t.Fatalf("publish listing: %v", err)
	}

	claim, err := svc.CreateClaim(ctx, CreateClaimRequest{
		TenantID:          "ten_01",
		ClaimID:           "claim_01",
		ListingID:         listing.ListingID,
		ExecutorProfileID: executor.ProfileID,
		ClaimType:         ClaimTypeWholeTask,
	})
	if err != nil {
		t.Fatalf("create claim: %v", err)
	}
	if _, err := svc.AwardClaim(ctx, AwardClaimRequest{
		TenantID:           "ten_01",
		ClaimID:            claim.ClaimID,
		PolicyChecksPassed: false,
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected sealed claim award policy check failure, got %v", err)
	}
	awarded, err := svc.AwardClaim(ctx, AwardClaimRequest{
		TenantID:           "ten_01",
		ClaimID:            claim.ClaimID,
		PolicyChecksPassed: true,
	})
	if err != nil {
		t.Fatalf("award claim: %v", err)
	}
	if awarded.Status != ClaimStatusAwarded {
		t.Fatalf("claim status=%q want %q", awarded.Status, ClaimStatusAwarded)
	}
	if _, err := svc.WithdrawClaim(ctx, WithdrawClaimRequest{
		TenantID: "ten_01",
		ClaimID:  claim.ClaimID,
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected withdraw non-submitted claim error, got %v", err)
	}

	bid, err := svc.CreateBid(ctx, CreateBidRequest{
		TenantID:          "ten_01",
		BidID:             "bid_01",
		ListingID:         listing.ListingID,
		ExecutorProfileID: executor.ProfileID,
		Amount:            55,
		Currency:          "JWUSD",
	})
	if err != nil {
		t.Fatalf("create bid: %v", err)
	}
	if _, err := svc.AcceptBid(ctx, AcceptBidRequest{
		TenantID:           "ten_01",
		BidID:              bid.BidID,
		PolicyChecksPassed: false,
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected sealed bid accept policy check failure, got %v", err)
	}
	accepted, err := svc.AcceptBid(ctx, AcceptBidRequest{
		TenantID:           "ten_01",
		BidID:              bid.BidID,
		PolicyChecksPassed: true,
	})
	if err != nil {
		t.Fatalf("accept bid: %v", err)
	}
	if accepted.Status != BidStatusAccepted {
		t.Fatalf("bid status=%q want %q", accepted.Status, BidStatusAccepted)
	}
	if _, err := svc.GetClaim(ctx, "ten_01", claim.ClaimID); err != nil {
		t.Fatalf("get claim: %v", err)
	}
	if _, err := svc.GetBid(ctx, "ten_01", bid.BidID); err != nil {
		t.Fatalf("get bid: %v", err)
	}

	for i := 2; i <= 4; i++ {
		listingID := "ml_claims_cap_" + string(rune('0'+i))
		taskID := "task_claims_cap_" + string(rune('0'+i))
		if _, err := svc.CreateListing(ctx, CreateListingRequest{
			TenantID:           "ten_01",
			ListingID:          listingID,
			TaskID:             taskID,
			RequesterProfileID: requester.ProfileID,
			WorkClass:          WorkClassPublicOpen,
			ListingMode:        ListingModeFixedOpenClaim,
			BudgetTotal:        5,
			Currency:           "JWUSD",
			FundingAccountID:   "fund_claims",
		}); err != nil {
			t.Fatalf("create listing %s: %v", listingID, err)
		}
		if _, err := svc.PublishListing(ctx, PublishListingRequest{
			TenantID:  "ten_01",
			ListingID: listingID,
		}); err != nil {
			t.Fatalf("publish listing %s: %v", listingID, err)
		}
		if i < 4 {
			if _, err := svc.CreateClaim(ctx, CreateClaimRequest{
				TenantID:          "ten_01",
				ListingID:         listingID,
				ExecutorProfileID: executor.ProfileID,
				ClaimType:         ClaimTypeWholeTask,
			}); err != nil {
				t.Fatalf("create claim %s: %v", listingID, err)
			}
		}
	}

	if _, err := svc.CreateClaim(ctx, CreateClaimRequest{
		TenantID:          "ten_01",
		ListingID:         "ml_claims_cap_4",
		ExecutorProfileID: executor.ProfileID,
		ClaimType:         ClaimTypeWholeTask,
	}); !errors.Is(err, ErrClaimHoarding) {
		t.Fatalf("expected claim hoarding error, got %v", err)
	}
}

func TestMarketClaimsBidsListAndWithdrawBranches(t *testing.T) {
	svc := New()
	ctx := context.Background()

	requester, err := svc.CreateProfile(ctx, CreateProfileRequest{
		TenantID:    "ten_01",
		ProfileID:   "mp_req_extra",
		ProfileType: ProfileTypeRequester,
		DisplayName: "Requester Extra",
	})
	if err != nil {
		t.Fatalf("create requester: %v", err)
	}
	execA, err := svc.CreateProfile(ctx, CreateProfileRequest{
		TenantID:    "ten_01",
		ProfileID:   "mp_exec_a",
		ProfileType: ProfileTypeExecutor,
		DisplayName: "Executor A",
	})
	if err != nil {
		t.Fatalf("create exec A: %v", err)
	}
	execB, err := svc.CreateProfile(ctx, CreateProfileRequest{
		TenantID:    "ten_01",
		ProfileID:   "mp_exec_b",
		ProfileType: ProfileTypeExecutor,
		DisplayName: "Executor B",
	})
	if err != nil {
		t.Fatalf("create exec B: %v", err)
	}
	if _, err := svc.CreateFundingAccount(ctx, CreateFundingAccountRequest{
		TenantID:       "ten_01",
		AccountID:      "fund_extra",
		OwnerProfileID: requester.ProfileID,
		Currency:       "JWUSD",
		InitialBalance: 300,
	}); err != nil {
		t.Fatalf("create funding: %v", err)
	}

	listing, err := svc.CreateListing(ctx, CreateListingRequest{
		TenantID:           "ten_01",
		ListingID:          "ml_extra",
		TaskID:             "task_extra",
		RequesterProfileID: requester.ProfileID,
		WorkClass:          WorkClassPublicOpen,
		ListingMode:        ListingModeFixedBidSelect,
		BudgetTotal:        40,
		Currency:           "JWUSD",
		FundingAccountID:   "fund_extra",
	})
	if err != nil {
		t.Fatalf("create listing: %v", err)
	}
	if _, err := svc.PublishListing(ctx, PublishListingRequest{
		TenantID:  "ten_01",
		ListingID: listing.ListingID,
	}); err != nil {
		t.Fatalf("publish listing: %v", err)
	}

	if _, err := svc.CreateClaim(ctx, CreateClaimRequest{
		TenantID:          "ten_01",
		ClaimID:           "claim_bad",
		ListingID:         listing.ListingID,
		ExecutorProfileID: execA.ProfileID,
		ClaimType:         "bad",
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid claim type error, got %v", err)
	}

	claimA, err := svc.CreateClaim(ctx, CreateClaimRequest{
		TenantID:          "ten_01",
		ClaimID:           "claim_a",
		ListingID:         listing.ListingID,
		ExecutorProfileID: execA.ProfileID,
		Metadata:          map[string]any{"source": "test"},
	})
	if err != nil {
		t.Fatalf("create claim A: %v", err)
	}
	claimB, err := svc.CreateClaim(ctx, CreateClaimRequest{
		TenantID:          "ten_01",
		ClaimID:           "claim_b",
		ListingID:         listing.ListingID,
		ExecutorProfileID: execB.ProfileID,
		ClaimType:         ClaimTypeShard,
	})
	if err != nil {
		t.Fatalf("create claim B: %v", err)
	}

	claims := svc.ListClaims(ctx, "ten_01", listing.ListingID)
	if len(claims) != 2 {
		t.Fatalf("expected two claims, got %d", len(claims))
	}
	if len(svc.ListClaims(ctx, "ten_01", "missing_listing")) != 0 {
		t.Fatalf("expected empty claim list for missing listing")
	}

	withdrawn, err := svc.WithdrawClaim(ctx, WithdrawClaimRequest{
		TenantID: "ten_01",
		ClaimID:  claimB.ClaimID,
		Reason:   "changed mind",
	})
	if err != nil {
		t.Fatalf("withdraw claim: %v", err)
	}
	if withdrawn.Status != ClaimStatusWithdrawn {
		t.Fatalf("withdrawn claim status=%q want %q", withdrawn.Status, ClaimStatusWithdrawn)
	}
	if _, err := svc.WithdrawClaim(ctx, WithdrawClaimRequest{
		TenantID: "ten_01",
		ClaimID:  claimB.ClaimID,
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected re-withdraw invalid request, got %v", err)
	}

	if _, err := svc.AwardClaim(ctx, AwardClaimRequest{
		TenantID: "ten_01",
		ClaimID:  claimA.ClaimID,
	}); err != nil {
		t.Fatalf("award claim A: %v", err)
	}

	bidA, err := svc.CreateBid(ctx, CreateBidRequest{
		TenantID:          "ten_01",
		BidID:             "bid_a",
		ListingID:         listing.ListingID,
		ExecutorProfileID: execA.ProfileID,
		Amount:            35,
		Currency:          "JWUSD",
		Metadata:          map[string]any{"rank": "a"},
	})
	if err != nil {
		t.Fatalf("create bid A: %v", err)
	}
	bidB, err := svc.CreateBid(ctx, CreateBidRequest{
		TenantID:          "ten_01",
		BidID:             "bid_b",
		ListingID:         listing.ListingID,
		ExecutorProfileID: execB.ProfileID,
		Amount:            34,
		Currency:          "JWUSD",
	})
	if err != nil {
		t.Fatalf("create bid B: %v", err)
	}

	accepted, err := svc.AcceptBid(ctx, AcceptBidRequest{
		TenantID: "ten_01",
		BidID:    bidB.BidID,
	})
	if err != nil {
		t.Fatalf("accept bid B: %v", err)
	}
	if accepted.Status != BidStatusAccepted {
		t.Fatalf("accepted bid status=%q want %q", accepted.Status, BidStatusAccepted)
	}
	rejected, err := svc.GetBid(ctx, "ten_01", bidA.BidID)
	if err != nil {
		t.Fatalf("get bid A: %v", err)
	}
	if rejected.Status != BidStatusRejected {
		t.Fatalf("rejected bid status=%q want %q", rejected.Status, BidStatusRejected)
	}
	if _, err := svc.AcceptBid(ctx, AcceptBidRequest{
		TenantID: "ten_01",
		BidID:    bidB.BidID,
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected accepting accepted bid to fail, got %v", err)
	}

	bids := svc.ListBids(ctx, "ten_01", listing.ListingID)
	if len(bids) != 2 {
		t.Fatalf("expected two bids, got %d", len(bids))
	}
	if len(svc.ListBids(ctx, "ten_02", listing.ListingID)) != 0 {
		t.Fatalf("expected tenant isolated bid list")
	}
}

func TestMarketAntiSpamAndSybilControls(t *testing.T) {
	svc := New()
	ctx := context.Background()

	pendingRequester, err := svc.CreateProfile(ctx, CreateProfileRequest{
		TenantID:    "ten_01",
		ProfileID:   "mp_req_pending",
		ProfileType: ProfileTypeRequester,
		DisplayName: "Pending Requester",
	})
	if err != nil {
		t.Fatalf("create pending requester: %v", err)
	}
	verifiedRequester, err := svc.CreateProfile(ctx, CreateProfileRequest{
		TenantID:           "ten_01",
		ProfileID:          "mp_req_verified",
		ProfileType:        ProfileTypeRequester,
		DisplayName:        "Verified Requester",
		VerificationStatus: VerificationVerified,
	})
	if err != nil {
		t.Fatalf("create verified requester: %v", err)
	}
	pendingExecutor, err := svc.CreateProfile(ctx, CreateProfileRequest{
		TenantID:    "ten_01",
		ProfileID:   "mp_exec_pending",
		ProfileType: ProfileTypeExecutor,
		DisplayName: "Pending Executor",
	})
	if err != nil {
		t.Fatalf("create pending executor: %v", err)
	}
	verifiedExecutor, err := svc.CreateProfile(ctx, CreateProfileRequest{
		TenantID:           "ten_01",
		ProfileID:          "mp_exec_verified",
		ProfileType:        ProfileTypeExecutor,
		DisplayName:        "Verified Executor",
		VerificationStatus: VerificationVerified,
	})
	if err != nil {
		t.Fatalf("create verified executor: %v", err)
	}
	if _, err := svc.CreateFundingAccount(ctx, CreateFundingAccountRequest{
		TenantID:       "ten_01",
		AccountID:      "fund_pending_req",
		OwnerProfileID: pendingRequester.ProfileID,
		Currency:       "JWUSD",
		InitialBalance: 1000,
	}); err != nil {
		t.Fatalf("create pending requester funding: %v", err)
	}
	if _, err := svc.CreateFundingAccount(ctx, CreateFundingAccountRequest{
		TenantID:       "ten_01",
		AccountID:      "fund_verified_req",
		OwnerProfileID: verifiedRequester.ProfileID,
		Currency:       "JWUSD",
		InitialBalance: 1000,
	}); err != nil {
		t.Fatalf("create verified requester funding: %v", err)
	}

	highBudgetListing, err := svc.CreateListing(ctx, CreateListingRequest{
		TenantID:           "ten_01",
		ListingID:          "ml_pending_high_budget",
		TaskID:             "task_pending_high_budget",
		RequesterProfileID: pendingRequester.ProfileID,
		WorkClass:          WorkClassPublicOpen,
		ListingMode:        ListingModeFixedOpenClaim,
		BudgetTotal:        250,
		Currency:           "JWUSD",
		FundingAccountID:   "fund_pending_req",
	})
	if err != nil {
		t.Fatalf("create high budget listing: %v", err)
	}
	if _, err := svc.PublishListing(ctx, PublishListingRequest{
		TenantID:  "ten_01",
		ListingID: highBudgetListing.ListingID,
	}); !errors.Is(err, ErrSybilControl) {
		t.Fatalf("expected pending requester budget cap error, got %v", err)
	}

	sealedListing, err := svc.CreateListing(ctx, CreateListingRequest{
		TenantID:           "ten_01",
		ListingID:          "ml_sealed_guard",
		TaskID:             "task_sealed_guard",
		RequesterProfileID: verifiedRequester.ProfileID,
		WorkClass:          WorkClassPublicSealed,
		ListingMode:        ListingModeFixedOpenClaim,
		BudgetTotal:        40,
		Currency:           "JWUSD",
		FundingAccountID:   "fund_verified_req",
	})
	if err != nil {
		t.Fatalf("create sealed listing: %v", err)
	}
	if _, err := svc.PublishListing(ctx, PublishListingRequest{
		TenantID:  "ten_01",
		ListingID: sealedListing.ListingID,
	}); err != nil {
		t.Fatalf("publish sealed listing: %v", err)
	}

	if _, err := svc.CreateClaim(ctx, CreateClaimRequest{
		TenantID:          "ten_01",
		ListingID:         sealedListing.ListingID,
		ExecutorProfileID: pendingExecutor.ProfileID,
	}); !errors.Is(err, ErrSybilControl) {
		t.Fatalf("expected sealed-work claim verification error, got %v", err)
	}
	if _, err := svc.CreateBid(ctx, CreateBidRequest{
		TenantID:          "ten_01",
		ListingID:         sealedListing.ListingID,
		ExecutorProfileID: pendingExecutor.ProfileID,
		Amount:            35,
		Currency:          "JWUSD",
	}); !errors.Is(err, ErrSybilControl) {
		t.Fatalf("expected sealed-work bid verification error, got %v", err)
	}
	if _, err := svc.CreateClaim(ctx, CreateClaimRequest{
		TenantID:          "ten_01",
		ListingID:         sealedListing.ListingID,
		ExecutorProfileID: verifiedExecutor.ProfileID,
	}); err != nil {
		t.Fatalf("expected verified executor claim success, got %v", err)
	}

	for i := 0; i < 5; i++ {
		suffix := string(rune('a' + i))
		listing, err := svc.CreateListing(ctx, CreateListingRequest{
			TenantID:           "ten_01",
			ListingID:          "ml_quota_" + suffix,
			TaskID:             "task_quota_" + suffix,
			RequesterProfileID: verifiedRequester.ProfileID,
			WorkClass:          WorkClassPublicOpen,
			ListingMode:        ListingModeFixedOpenClaim,
			BudgetTotal:        10,
			Currency:           "JWUSD",
			FundingAccountID:   "fund_verified_req",
		})
		if err != nil {
			t.Fatalf("create quota listing %d: %v", i, err)
		}

		_, err = svc.PublishListing(ctx, PublishListingRequest{
			TenantID:  "ten_01",
			ListingID: listing.ListingID,
		})
		if i < 4 && err != nil {
			t.Fatalf("publish quota listing %d: %v", i, err)
		}
		if i == 4 && !errors.Is(err, ErrAntiSpamControl) {
			t.Fatalf("expected quota violation on final publish, got %v", err)
		}
	}
}
