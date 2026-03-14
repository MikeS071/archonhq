package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	disputessvc "github.com/MikeS071/archonhq/services/disputes"
	escrowsvc "github.com/MikeS071/archonhq/services/escrow"
	marketplacesvc "github.com/MikeS071/archonhq/services/marketplace"
	payoutssvc "github.com/MikeS071/archonhq/services/payouts"
	simulationsvc "github.com/MikeS071/archonhq/services/simulation"
)

func TestM9MarketProfileFundingAndListingFlow(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	createProfileReq := newJSONRequest(t, http.MethodPost, "/v1/market/profiles", "human:ten_01:user_admin:tenant_admin", "idem_m9_profile_create", map[string]any{
		"profile_id":   "mp_req_01",
		"profile_type": "requester",
		"display_name": "Requester One",
	})
	rrCreateProfile := httptest.NewRecorder()
	h.ServeHTTP(rrCreateProfile, createProfileReq)
	if rrCreateProfile.Code != http.StatusOK {
		t.Fatalf("create market profile expected 200 got %d body=%s", rrCreateProfile.Code, rrCreateProfile.Body.String())
	}

	createFundingReq := newJSONRequest(t, http.MethodPost, "/v1/market/funding-accounts", "human:ten_01:user_admin:tenant_admin", "idem_m9_funding_create", map[string]any{
		"account_id":        "fund_01",
		"owner_profile_id":  "mp_req_01",
		"currency":          "JWUSD",
		"initial_balance":   200,
		"reserve_policy_id": "default",
	})
	rrCreateFunding := httptest.NewRecorder()
	h.ServeHTTP(rrCreateFunding, createFundingReq)
	if rrCreateFunding.Code != http.StatusOK {
		t.Fatalf("create funding account expected 200 got %d body=%s", rrCreateFunding.Code, rrCreateFunding.Body.String())
	}

	createListingReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings", "human:ten_01:user_admin:tenant_admin", "idem_m9_listing_create", map[string]any{
		"listing_id":           "ml_01",
		"task_id":              "task_01",
		"requester_profile_id": "mp_req_01",
		"work_class":           "public_open",
		"listing_mode":         "fixed_price_open_claim",
		"budget_total":         120,
		"currency":             "JWUSD",
		"funding_account_id":   "fund_01",
	})
	rrCreateListing := httptest.NewRecorder()
	h.ServeHTTP(rrCreateListing, createListingReq)
	if rrCreateListing.Code != http.StatusOK {
		t.Fatalf("create listing expected 200 got %d body=%s", rrCreateListing.Code, rrCreateListing.Body.String())
	}

	publishListingReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings/ml_01/publish", "human:ten_01:user_admin:tenant_admin", "idem_m9_listing_publish", map[string]any{"reason": "ready"})
	rrPublishListing := httptest.NewRecorder()
	h.ServeHTTP(rrPublishListing, publishListingReq)
	if rrPublishListing.Code != http.StatusOK {
		t.Fatalf("publish listing expected 200 got %d body=%s", rrPublishListing.Code, rrPublishListing.Body.String())
	}

	listingsReq := httptest.NewRequest(http.MethodGet, "/v1/market/listings", nil)
	listingsReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrListings := httptest.NewRecorder()
	h.ServeHTTP(rrListings, listingsReq)
	if rrListings.Code != http.StatusOK {
		t.Fatalf("list listings expected 200 got %d body=%s", rrListings.Code, rrListings.Body.String())
	}
	if !strings.Contains(rrListings.Body.String(), "ml_01") {
		t.Fatalf("listing feed missing listing id body=%s", rrListings.Body.String())
	}

	getListingReq := httptest.NewRequest(http.MethodGet, "/v1/market/listings/ml_01", nil)
	getListingReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrGetListing := httptest.NewRecorder()
	h.ServeHTTP(rrGetListing, getListingReq)
	if rrGetListing.Code != http.StatusOK {
		t.Fatalf("get listing expected 200 got %d body=%s", rrGetListing.Code, rrGetListing.Body.String())
	}

	cancelListingReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings/ml_01/cancel", "human:ten_01:user_admin:tenant_admin", "idem_m9_listing_cancel", map[string]any{"reason": "no_longer_needed"})
	rrCancelListing := httptest.NewRecorder()
	h.ServeHTTP(rrCancelListing, cancelListingReq)
	if rrCancelListing.Code != http.StatusOK {
		t.Fatalf("cancel listing expected 200 got %d body=%s", rrCancelListing.Code, rrCancelListing.Body.String())
	}

	getFundingReq := httptest.NewRequest(http.MethodGet, "/v1/market/funding-accounts/fund_01", nil)
	getFundingReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrGetFunding := httptest.NewRecorder()
	h.ServeHTTP(rrGetFunding, getFundingReq)
	if rrGetFunding.Code != http.StatusOK {
		t.Fatalf("get funding account expected 200 got %d body=%s", rrGetFunding.Code, rrGetFunding.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(rrGetFunding.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode funding payload: %v", err)
	}
	funding, _ := payload["funding_account"].(map[string]any)
	if funding == nil {
		t.Fatalf("missing funding_account payload: %s", rrGetFunding.Body.String())
	}
	if got := funding["available_balance"]; got != float64(200) {
		t.Fatalf("available_balance=%v want 200", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestM9WorkClassAndReserveGuardrails(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	createProfileForbiddenReq := newJSONRequest(t, http.MethodPost, "/v1/market/profiles", "human:ten_01:user_dev:developer", "idem_m9_profile_forbidden", map[string]any{
		"profile_type": "requester",
		"display_name": "Forbidden",
	})
	rrCreateProfileForbidden := httptest.NewRecorder()
	h.ServeHTTP(rrCreateProfileForbidden, createProfileForbiddenReq)
	if rrCreateProfileForbidden.Code != http.StatusForbidden {
		t.Fatalf("create profile forbidden expected 403 got %d body=%s", rrCreateProfileForbidden.Code, rrCreateProfileForbidden.Body.String())
	}

	createProfileReq := newJSONRequest(t, http.MethodPost, "/v1/market/profiles", "human:ten_01:user_admin:tenant_admin", "idem_m9_profile_create_2", map[string]any{
		"profile_id":   "mp_req_02",
		"profile_type": "requester",
		"display_name": "Requester Two",
	})
	rrCreateProfile := httptest.NewRecorder()
	h.ServeHTTP(rrCreateProfile, createProfileReq)
	if rrCreateProfile.Code != http.StatusOK {
		t.Fatalf("create market profile expected 200 got %d body=%s", rrCreateProfile.Code, rrCreateProfile.Body.String())
	}

	createFundingReq := newJSONRequest(t, http.MethodPost, "/v1/market/funding-accounts", "human:ten_01:user_admin:tenant_admin", "idem_m9_funding_create_2", map[string]any{
		"account_id":       "fund_02",
		"owner_profile_id": "mp_req_02",
		"currency":         "JWUSD",
		"initial_balance":  50,
	})
	rrCreateFunding := httptest.NewRecorder()
	h.ServeHTTP(rrCreateFunding, createFundingReq)
	if rrCreateFunding.Code != http.StatusOK {
		t.Fatalf("create funding account expected 200 got %d body=%s", rrCreateFunding.Code, rrCreateFunding.Body.String())
	}

	createPrivateListingReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings", "human:ten_01:user_admin:tenant_admin", "idem_m9_listing_private_create", map[string]any{
		"listing_id":           "ml_private",
		"task_id":              "task_private",
		"requester_profile_id": "mp_req_02",
		"work_class":           "private_tenant_only",
		"listing_mode":         "fixed_price_open_claim",
		"budget_total":         20,
		"currency":             "JWUSD",
		"funding_account_id":   "fund_02",
	})
	rrCreatePrivateListing := httptest.NewRecorder()
	h.ServeHTTP(rrCreatePrivateListing, createPrivateListingReq)
	if rrCreatePrivateListing.Code != http.StatusOK {
		t.Fatalf("create private listing expected 200 got %d body=%s", rrCreatePrivateListing.Code, rrCreatePrivateListing.Body.String())
	}

	publishPrivateListingReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings/ml_private/publish", "human:ten_01:user_admin:tenant_admin", "idem_m9_listing_private_publish", map[string]any{"reason": "not_allowed"})
	rrPublishPrivateListing := httptest.NewRecorder()
	h.ServeHTTP(rrPublishPrivateListing, publishPrivateListingReq)
	if rrPublishPrivateListing.Code != http.StatusBadRequest {
		t.Fatalf("publish private listing expected 400 got %d body=%s", rrPublishPrivateListing.Code, rrPublishPrivateListing.Body.String())
	}

	createUnfundedListingReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings", "human:ten_01:user_admin:tenant_admin", "idem_m9_listing_unfunded_create", map[string]any{
		"listing_id":           "ml_unfunded",
		"task_id":              "task_unfunded",
		"requester_profile_id": "mp_req_02",
		"work_class":           "public_open",
		"listing_mode":         "fixed_price_open_claim",
		"budget_total":         80,
		"currency":             "JWUSD",
		"funding_account_id":   "fund_02",
	})
	rrCreateUnfundedListing := httptest.NewRecorder()
	h.ServeHTTP(rrCreateUnfundedListing, createUnfundedListingReq)
	if rrCreateUnfundedListing.Code != http.StatusOK {
		t.Fatalf("create unfunded listing expected 200 got %d body=%s", rrCreateUnfundedListing.Code, rrCreateUnfundedListing.Body.String())
	}

	publishUnfundedListingReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings/ml_unfunded/publish", "human:ten_01:user_admin:tenant_admin", "idem_m9_listing_unfunded_publish", map[string]any{"reason": "insufficient"})
	rrPublishUnfundedListing := httptest.NewRecorder()
	h.ServeHTTP(rrPublishUnfundedListing, publishUnfundedListingReq)
	if rrPublishUnfundedListing.Code != http.StatusConflict {
		t.Fatalf("publish unfunded listing expected 409 got %d body=%s", rrPublishUnfundedListing.Code, rrPublishUnfundedListing.Body.String())
	}
	if !strings.Contains(rrPublishUnfundedListing.Body.String(), "insufficient_funded_reserve") {
		t.Fatalf("expected insufficient_funded_reserve code body=%s", rrPublishUnfundedListing.Body.String())
	}

	getListingCrossTenantReq := httptest.NewRequest(http.MethodGet, "/v1/market/listings/ml_unfunded", nil)
	getListingCrossTenantReq.Header.Set("Authorization", "Bearer human:ten_02:user_admin:tenant_admin")
	rrGetListingCrossTenant := httptest.NewRecorder()
	h.ServeHTTP(rrGetListingCrossTenant, getListingCrossTenantReq)
	if rrGetListingCrossTenant.Code != http.StatusNotFound {
		t.Fatalf("cross tenant listing read expected 404 got %d body=%s", rrGetListingCrossTenant.Code, rrGetListingCrossTenant.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestM9ProfilePatchReadAndInvalidJSON(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	createProfileReq := newJSONRequest(t, http.MethodPost, "/v1/market/profiles", "human:ten_01:user_admin:tenant_admin", "idem_m9_profile_create_3", map[string]any{
		"profile_id":   "mp_patch_01",
		"profile_type": "requester",
		"display_name": "Patch Me",
	})
	rrCreateProfile := httptest.NewRecorder()
	h.ServeHTTP(rrCreateProfile, createProfileReq)
	if rrCreateProfile.Code != http.StatusOK {
		t.Fatalf("create profile expected 200 got %d body=%s", rrCreateProfile.Code, rrCreateProfile.Body.String())
	}

	getProfileReq := httptest.NewRequest(http.MethodGet, "/v1/market/profiles/mp_patch_01", nil)
	getProfileReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrGetProfile := httptest.NewRecorder()
	h.ServeHTTP(rrGetProfile, getProfileReq)
	if rrGetProfile.Code != http.StatusOK {
		t.Fatalf("get profile expected 200 got %d body=%s", rrGetProfile.Code, rrGetProfile.Body.String())
	}

	patchProfileReq := newJSONRequest(t, http.MethodPatch, "/v1/market/profiles/mp_patch_01", "human:ten_01:user_admin:tenant_admin", "idem_m9_profile_patch_1", map[string]any{
		"display_name":        "Patched Name",
		"verification_status": "verified",
		"work_class_allowlist": []string{
			"public_open",
			"restricted_market",
		},
	})
	rrPatchProfile := httptest.NewRecorder()
	h.ServeHTTP(rrPatchProfile, patchProfileReq)
	if rrPatchProfile.Code != http.StatusOK {
		t.Fatalf("patch profile expected 200 got %d body=%s", rrPatchProfile.Code, rrPatchProfile.Body.String())
	}

	reputationReq := httptest.NewRequest(http.MethodGet, "/v1/market/profiles/mp_patch_01/reputation", nil)
	reputationReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrReputation := httptest.NewRecorder()
	h.ServeHTTP(rrReputation, reputationReq)
	if rrReputation.Code != http.StatusOK {
		t.Fatalf("profile reputation expected 200 got %d body=%s", rrReputation.Code, rrReputation.Body.String())
	}

	for _, tc := range []struct {
		name   string
		method string
		path   string
		token  string
		idem   string
	}{
		{name: "profile create bad json", method: http.MethodPost, path: "/v1/market/profiles", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m9_bad_1"},
		{name: "profile patch bad json", method: http.MethodPatch, path: "/v1/market/profiles/mp_patch_01", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m9_bad_2"},
		{name: "funding create bad json", method: http.MethodPost, path: "/v1/market/funding-accounts", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m9_bad_3"},
		{name: "listing create bad json", method: http.MethodPost, path: "/v1/market/listings", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m9_bad_4"},
		{name: "listing publish bad json", method: http.MethodPost, path: "/v1/market/listings/ml_missing/publish", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m9_bad_5"},
		{name: "listing cancel bad json", method: http.MethodPost, path: "/v1/market/listings/ml_missing/cancel", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m9_bad_6"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader("{"))
			req.Header.Set("Authorization", "Bearer "+tc.token)
			req.Header.Set("Idempotency-Key", tc.idem)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("%s expected 400 got %d body=%s", tc.name, rr.Code, rr.Body.String())
			}
		})
	}

	getMissingFundingReq := httptest.NewRequest(http.MethodGet, "/v1/market/funding-accounts/missing", nil)
	getMissingFundingReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrGetMissingFunding := httptest.NewRecorder()
	h.ServeHTTP(rrGetMissingFunding, getMissingFundingReq)
	if rrGetMissingFunding.Code != http.StatusNotFound {
		t.Fatalf("missing funding account expected 404 got %d body=%s", rrGetMissingFunding.Code, rrGetMissingFunding.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestM9MarketplaceErrorWriter(t *testing.T) {
	dbMock, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})

	rr := httptest.NewRecorder()
	srv.writeMarketplaceError(rr, "corr_1", "code_1", "message", marketplacesvc.ErrNotFound)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("not found expected 404 got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	srv.writeMarketplaceError(rr, "corr_1", "code_1", "message", marketplacesvc.ErrAlreadyExists)
	if rr.Code != http.StatusConflict {
		t.Fatalf("already exists expected 409 got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	srv.writeMarketplaceError(rr, "corr_1", "code_1", "message", marketplacesvc.ErrClaimHoarding)
	if rr.Code != http.StatusConflict || !strings.Contains(rr.Body.String(), "claim_hoarding_limit") {
		t.Fatalf("claim hoarding expected 409 with code, got %d body=%s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	srv.writeMarketplaceError(rr, "corr_1", "code_1", "message", marketplacesvc.ErrAntiSpamControl)
	if rr.Code != http.StatusTooManyRequests || !strings.Contains(rr.Body.String(), "market_spam_control") {
		t.Fatalf("anti-spam expected 429 with code, got %d body=%s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	srv.writeMarketplaceError(rr, "corr_1", "code_1", "message", marketplacesvc.ErrSybilControl)
	if rr.Code != http.StatusForbidden || !strings.Contains(rr.Body.String(), "market_identity_verification_required") {
		t.Fatalf("sybil guard expected 403 with code, got %d body=%s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	srv.writeMarketplaceError(rr, "corr_1", "code_1", "message", marketplacesvc.ErrInsufficientFunds)
	if rr.Code != http.StatusConflict || !strings.Contains(rr.Body.String(), "insufficient_funded_reserve") {
		t.Fatalf("insufficient funds expected 409 with code, got %d body=%s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	srv.writeMarketplaceError(rr, "corr_1", "code_1", "message", marketplacesvc.ErrInvalidRequest)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("invalid request expected 400 got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	srv.writeMarketplaceError(rr, "corr_1", "code_1", "message", errors.New("boom"))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("fallback expected 500 got %d", rr.Code)
	}
}

func TestM9ClaimsEscrowPayoutDisputeFlows(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	createRequesterReq := newJSONRequest(t, http.MethodPost, "/v1/market/profiles", "human:ten_01:user_admin:tenant_admin", "idem_m9_p2_profile_req", map[string]any{
		"profile_id":   "mp_req_p2",
		"profile_type": "requester",
		"display_name": "Requester P2",
	})
	rrCreateRequester := httptest.NewRecorder()
	h.ServeHTTP(rrCreateRequester, createRequesterReq)
	if rrCreateRequester.Code != http.StatusOK {
		t.Fatalf("create requester profile expected 200 got %d body=%s", rrCreateRequester.Code, rrCreateRequester.Body.String())
	}

	createExecutorReq := newJSONRequest(t, http.MethodPost, "/v1/market/profiles", "human:ten_01:user_admin:tenant_admin", "idem_m9_p2_profile_exec", map[string]any{
		"profile_id":   "mp_exec_p2",
		"profile_type": "executor",
		"display_name": "Executor P2",
	})
	rrCreateExecutor := httptest.NewRecorder()
	h.ServeHTTP(rrCreateExecutor, createExecutorReq)
	if rrCreateExecutor.Code != http.StatusOK {
		t.Fatalf("create executor profile expected 200 got %d body=%s", rrCreateExecutor.Code, rrCreateExecutor.Body.String())
	}

	createFundingReq := newJSONRequest(t, http.MethodPost, "/v1/market/funding-accounts", "human:ten_01:user_admin:tenant_admin", "idem_m9_p2_funding", map[string]any{
		"account_id":       "fund_p2",
		"owner_profile_id": "mp_req_p2",
		"currency":         "JWUSD",
		"initial_balance":  500,
	})
	rrCreateFunding := httptest.NewRecorder()
	h.ServeHTTP(rrCreateFunding, createFundingReq)
	if rrCreateFunding.Code != http.StatusOK {
		t.Fatalf("create funding account expected 200 got %d body=%s", rrCreateFunding.Code, rrCreateFunding.Body.String())
	}

	createListingReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings", "human:ten_01:user_admin:tenant_admin", "idem_m9_p2_listing", map[string]any{
		"listing_id":           "ml_p2",
		"task_id":              "task_p2",
		"requester_profile_id": "mp_req_p2",
		"work_class":           "public_open",
		"listing_mode":         "fixed_price_open_claim",
		"budget_total":         100,
		"currency":             "JWUSD",
		"funding_account_id":   "fund_p2",
	})
	rrCreateListing := httptest.NewRecorder()
	h.ServeHTTP(rrCreateListing, createListingReq)
	if rrCreateListing.Code != http.StatusOK {
		t.Fatalf("create listing expected 200 got %d body=%s", rrCreateListing.Code, rrCreateListing.Body.String())
	}

	publishListingReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings/ml_p2/publish", "human:ten_01:user_admin:tenant_admin", "idem_m9_p2_publish", map[string]any{"reason": "ready"})
	rrPublishListing := httptest.NewRecorder()
	h.ServeHTTP(rrPublishListing, publishListingReq)
	if rrPublishListing.Code != http.StatusOK {
		t.Fatalf("publish listing expected 200 got %d body=%s", rrPublishListing.Code, rrPublishListing.Body.String())
	}

	createClaimReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings/ml_p2/claims", "human:ten_01:user_exec:executor", "idem_m9_p2_claim", map[string]any{
		"claim_id":             "claim_p2",
		"executor_profile_id":  "mp_exec_p2",
		"claim_type":           "whole_task",
		"policy_checks_passed": true,
	})
	rrCreateClaim := httptest.NewRecorder()
	h.ServeHTTP(rrCreateClaim, createClaimReq)
	if rrCreateClaim.Code != http.StatusOK {
		t.Fatalf("create claim expected 200 got %d body=%s", rrCreateClaim.Code, rrCreateClaim.Body.String())
	}

	awardClaimReq := newJSONRequest(t, http.MethodPost, "/v1/market/claims/claim_p2/award", "human:ten_01:user_app:approver", "idem_m9_p2_claim_award", map[string]any{"policy_checks_passed": true})
	rrAwardClaim := httptest.NewRecorder()
	h.ServeHTTP(rrAwardClaim, awardClaimReq)
	if rrAwardClaim.Code != http.StatusOK {
		t.Fatalf("award claim expected 200 got %d body=%s", rrAwardClaim.Code, rrAwardClaim.Body.String())
	}
	escrowID := jsonValue(t, rrAwardClaim.Body.Bytes(), "escrow.escrow_id")
	if escrowID == "" {
		escrowID = jsonValue(t, rrAwardClaim.Body.Bytes(), "escrow_id")
	}

	createBidReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings/ml_p2/bids", "human:ten_01:user_exec:executor", "idem_m9_p2_bid", map[string]any{
		"bid_id":              "bid_p2",
		"executor_profile_id": "mp_exec_p2",
		"amount":              90,
		"currency":            "JWUSD",
	})
	rrCreateBid := httptest.NewRecorder()
	h.ServeHTTP(rrCreateBid, createBidReq)
	if rrCreateBid.Code != http.StatusOK {
		t.Fatalf("create bid expected 200 got %d body=%s", rrCreateBid.Code, rrCreateBid.Body.String())
	}

	acceptBidReq := newJSONRequest(t, http.MethodPost, "/v1/market/bids/bid_p2/accept", "human:ten_01:user_app:approver", "idem_m9_p2_bid_accept", map[string]any{"policy_checks_passed": true})
	rrAcceptBid := httptest.NewRecorder()
	h.ServeHTTP(rrAcceptBid, acceptBidReq)
	if rrAcceptBid.Code != http.StatusOK {
		t.Fatalf("accept bid expected 200 got %d body=%s", rrAcceptBid.Code, rrAcceptBid.Body.String())
	}

	fundReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings/ml_p2/fund", "human:ten_01:user_admin:tenant_admin", "idem_m9_p2_fund", map[string]any{"amount": 40})
	rrFund := httptest.NewRecorder()
	h.ServeHTTP(rrFund, fundReq)
	if rrFund.Code != http.StatusOK {
		t.Fatalf("fund listing expected 200 got %d body=%s", rrFund.Code, rrFund.Body.String())
	}
	if escrowID == "" {
		escrowID = jsonValue(t, rrFund.Body.Bytes(), "escrow.escrow_id")
	}
	if escrowID == "" {
		t.Fatalf("missing escrow id from award/fund responses")
	}

	getEscrowReq := httptest.NewRequest(http.MethodGet, "/v1/market/escrows/"+escrowID, nil)
	getEscrowReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrGetEscrow := httptest.NewRecorder()
	h.ServeHTTP(rrGetEscrow, getEscrowReq)
	if rrGetEscrow.Code != http.StatusOK {
		t.Fatalf("get escrow expected 200 got %d body=%s", rrGetEscrow.Code, rrGetEscrow.Body.String())
	}

	releaseReq := newJSONRequest(t, http.MethodPost, "/v1/market/escrows/"+escrowID+"/release", "human:ten_01:user_app:approver", "idem_m9_p2_release", map[string]any{"amount": 20, "reason": "accepted"})
	rrRelease := httptest.NewRecorder()
	h.ServeHTTP(rrRelease, releaseReq)
	if rrRelease.Code != http.StatusOK {
		t.Fatalf("release escrow expected 200 got %d body=%s", rrRelease.Code, rrRelease.Body.String())
	}

	refundReq := newJSONRequest(t, http.MethodPost, "/v1/market/escrows/"+escrowID+"/refund", "human:ten_01:user_app:approver", "idem_m9_p2_refund", map[string]any{"amount": 10, "reason": "adjustment"})
	rrRefund := httptest.NewRecorder()
	h.ServeHTTP(rrRefund, refundReq)
	if rrRefund.Code != http.StatusOK {
		t.Fatalf("refund escrow expected 200 got %d body=%s", rrRefund.Code, rrRefund.Body.String())
	}

	createPayoutAccountReq := newJSONRequest(t, http.MethodPost, "/v1/payout-accounts", "human:ten_01:user_exec:executor", "idem_m9_p2_payout_acct", map[string]any{
		"payout_account_id":    "pa_p2",
		"owner_profile_id":     "mp_exec_p2",
		"provider":             "stripe",
		"provider_account_ref": "acct_p2",
		"jurisdiction":         "AU",
	})
	rrCreatePayoutAccount := httptest.NewRecorder()
	h.ServeHTTP(rrCreatePayoutAccount, createPayoutAccountReq)
	if rrCreatePayoutAccount.Code != http.StatusOK {
		t.Fatalf("create payout account expected 200 got %d body=%s", rrCreatePayoutAccount.Code, rrCreatePayoutAccount.Body.String())
	}

	createPayoutReq := newJSONRequest(t, http.MethodPost, "/v1/payouts", "human:ten_01:user_exec:executor", "idem_m9_p2_payout", map[string]any{
		"payout_id":         "po_p2",
		"payout_account_id": "pa_p2",
		"escrow_id":         escrowID,
		"amount":            15,
		"currency":          "JWUSD",
		"auto_complete":     true,
	})
	rrCreatePayout := httptest.NewRecorder()
	h.ServeHTTP(rrCreatePayout, createPayoutReq)
	if rrCreatePayout.Code != http.StatusOK {
		t.Fatalf("create payout expected 200 got %d body=%s", rrCreatePayout.Code, rrCreatePayout.Body.String())
	}

	getPayoutReq := httptest.NewRequest(http.MethodGet, "/v1/payouts/po_p2", nil)
	getPayoutReq.Header.Set("Authorization", "Bearer human:ten_01:user_exec:executor")
	rrGetPayout := httptest.NewRecorder()
	h.ServeHTTP(rrGetPayout, getPayoutReq)
	if rrGetPayout.Code != http.StatusOK {
		t.Fatalf("get payout expected 200 got %d body=%s", rrGetPayout.Code, rrGetPayout.Body.String())
	}

	openDisputeReq := newJSONRequest(t, http.MethodPost, "/v1/market/disputes", "human:ten_01:user_req:requester", "idem_m9_p2_dispute_open", map[string]any{
		"dispute_id":   "disp_p2",
		"listing_id":   "ml_p2",
		"escrow_id":    escrowID,
		"claim_id":     "claim_p2",
		"dispute_type": "acceptance_disagreement",
		"reason":       "quality mismatch",
	})
	rrOpenDispute := httptest.NewRecorder()
	h.ServeHTTP(rrOpenDispute, openDisputeReq)
	if rrOpenDispute.Code != http.StatusOK {
		t.Fatalf("open dispute expected 200 got %d body=%s", rrOpenDispute.Code, rrOpenDispute.Body.String())
	}

	resolveDisputeReq := newJSONRequest(t, http.MethodPost, "/v1/market/disputes/disp_p2/resolve", "human:ten_01:user_arb:arbitrator", "idem_m9_p2_dispute_resolve", map[string]any{
		"decision":              "partial_refund",
		"escrow_release_action": "refund",
		"escrow_amount":         5,
		"appeal_allowed":        true,
	})
	rrResolveDispute := httptest.NewRecorder()
	h.ServeHTTP(rrResolveDispute, resolveDisputeReq)
	if rrResolveDispute.Code != http.StatusOK {
		t.Fatalf("resolve dispute expected 200 got %d body=%s", rrResolveDispute.Code, rrResolveDispute.Body.String())
	}

	appealDisputeReq := newJSONRequest(t, http.MethodPost, "/v1/market/disputes/disp_p2/appeal", "human:ten_01:user_exec:executor", "idem_m9_p2_dispute_appeal", map[string]any{"reason": "new evidence"})
	rrAppealDispute := httptest.NewRecorder()
	h.ServeHTTP(rrAppealDispute, appealDisputeReq)
	if rrAppealDispute.Code != http.StatusOK {
		t.Fatalf("appeal dispute expected 200 got %d body=%s", rrAppealDispute.Code, rrAppealDispute.Body.String())
	}

	getDisputeReq := httptest.NewRequest(http.MethodGet, "/v1/market/disputes/disp_p2", nil)
	getDisputeReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrGetDispute := httptest.NewRecorder()
	h.ServeHTTP(rrGetDispute, getDisputeReq)
	if rrGetDispute.Code != http.StatusOK {
		t.Fatalf("get dispute expected 200 got %d body=%s", rrGetDispute.Code, rrGetDispute.Body.String())
	}

	dashboardReq := httptest.NewRequest(http.MethodGet, "/v1/market/dashboard", nil)
	dashboardReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrDashboard := httptest.NewRecorder()
	h.ServeHTTP(rrDashboard, dashboardReq)
	if rrDashboard.Code != http.StatusOK {
		t.Fatalf("market dashboard expected 200 got %d body=%s", rrDashboard.Code, rrDashboard.Body.String())
	}
	if !strings.Contains(rrDashboard.Body.String(), "market_dashboard") {
		t.Fatalf("market dashboard payload missing key body=%s", rrDashboard.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestM9EscrowPayoutDisputeErrorWriters(t *testing.T) {
	dbMock, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})

	rr := httptest.NewRecorder()
	srv.writeEscrowError(rr, "corr", "code", "msg", escrowsvc.ErrNotFound)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("escrow not found expected 404 got %d", rr.Code)
	}
	rr = httptest.NewRecorder()
	srv.writeEscrowError(rr, "corr", "code", "msg", escrowsvc.ErrInvalidRequest)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("escrow invalid expected 400 got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	srv.writePayoutError(rr, "corr", "code", "msg", payoutssvc.ErrNotFound)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("payout not found expected 404 got %d", rr.Code)
	}
	rr = httptest.NewRecorder()
	srv.writePayoutError(rr, "corr", "code", "msg", payoutssvc.ErrInvalidRequest)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("payout invalid expected 400 got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	srv.writeDisputeError(rr, "corr", "code", "msg", disputessvc.ErrNotFound)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("dispute not found expected 404 got %d", rr.Code)
	}
	rr = httptest.NewRecorder()
	srv.writeDisputeError(rr, "corr", "code", "msg", disputessvc.ErrInvalidRequest)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("dispute invalid expected 400 got %d", rr.Code)
	}
}

func TestM9WithdrawClaimAndGetPayoutAccountEndpoints(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	createRequesterReq := newJSONRequest(t, http.MethodPost, "/v1/market/profiles", "human:ten_01:user_admin:tenant_admin", "idem_m9_w_profile_req", map[string]any{
		"profile_id":   "mp_req_w",
		"profile_type": "requester",
		"display_name": "Requester W",
	})
	rrCreateRequester := httptest.NewRecorder()
	h.ServeHTTP(rrCreateRequester, createRequesterReq)
	if rrCreateRequester.Code != http.StatusOK {
		t.Fatalf("create requester profile expected 200 got %d body=%s", rrCreateRequester.Code, rrCreateRequester.Body.String())
	}

	createExecutorReq := newJSONRequest(t, http.MethodPost, "/v1/market/profiles", "human:ten_01:user_admin:tenant_admin", "idem_m9_w_profile_exec", map[string]any{
		"profile_id":   "mp_exec_w",
		"profile_type": "executor",
		"display_name": "Executor W",
	})
	rrCreateExecutor := httptest.NewRecorder()
	h.ServeHTTP(rrCreateExecutor, createExecutorReq)
	if rrCreateExecutor.Code != http.StatusOK {
		t.Fatalf("create executor profile expected 200 got %d body=%s", rrCreateExecutor.Code, rrCreateExecutor.Body.String())
	}

	createFundingReq := newJSONRequest(t, http.MethodPost, "/v1/market/funding-accounts", "human:ten_01:user_admin:tenant_admin", "idem_m9_w_funding", map[string]any{
		"account_id":       "fund_w",
		"owner_profile_id": "mp_req_w",
		"currency":         "JWUSD",
		"initial_balance":  100,
	})
	rrCreateFunding := httptest.NewRecorder()
	h.ServeHTTP(rrCreateFunding, createFundingReq)
	if rrCreateFunding.Code != http.StatusOK {
		t.Fatalf("create funding expected 200 got %d body=%s", rrCreateFunding.Code, rrCreateFunding.Body.String())
	}

	createListingReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings", "human:ten_01:user_admin:tenant_admin", "idem_m9_w_listing", map[string]any{
		"listing_id":           "ml_w",
		"task_id":              "task_w",
		"requester_profile_id": "mp_req_w",
		"work_class":           "public_open",
		"listing_mode":         "fixed_price_open_claim",
		"budget_total":         50,
		"currency":             "JWUSD",
		"funding_account_id":   "fund_w",
	})
	rrCreateListing := httptest.NewRecorder()
	h.ServeHTTP(rrCreateListing, createListingReq)
	if rrCreateListing.Code != http.StatusOK {
		t.Fatalf("create listing expected 200 got %d body=%s", rrCreateListing.Code, rrCreateListing.Body.String())
	}

	publishListingReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings/ml_w/publish", "human:ten_01:user_admin:tenant_admin", "idem_m9_w_publish", map[string]any{"reason": "ready"})
	rrPublishListing := httptest.NewRecorder()
	h.ServeHTTP(rrPublishListing, publishListingReq)
	if rrPublishListing.Code != http.StatusOK {
		t.Fatalf("publish listing expected 200 got %d body=%s", rrPublishListing.Code, rrPublishListing.Body.String())
	}

	createClaimReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings/ml_w/claims", "human:ten_01:user_exec:executor", "idem_m9_w_claim", map[string]any{
		"claim_id":            "claim_w",
		"executor_profile_id": "mp_exec_w",
	})
	rrCreateClaim := httptest.NewRecorder()
	h.ServeHTTP(rrCreateClaim, createClaimReq)
	if rrCreateClaim.Code != http.StatusOK {
		t.Fatalf("create claim expected 200 got %d body=%s", rrCreateClaim.Code, rrCreateClaim.Body.String())
	}

	invalidWithdrawReq := httptest.NewRequest(http.MethodPost, "/v1/market/claims/claim_w/withdraw", strings.NewReader("{"))
	invalidWithdrawReq.Header.Set("Authorization", "Bearer human:ten_01:user_exec:executor")
	invalidWithdrawReq.Header.Set("Idempotency-Key", "idem_m9_w_withdraw_bad")
	invalidWithdrawReq.Header.Set("Content-Type", "application/json")
	rrInvalidWithdraw := httptest.NewRecorder()
	h.ServeHTTP(rrInvalidWithdraw, invalidWithdrawReq)
	if rrInvalidWithdraw.Code != http.StatusBadRequest {
		t.Fatalf("invalid withdraw json expected 400 got %d body=%s", rrInvalidWithdraw.Code, rrInvalidWithdraw.Body.String())
	}

	withdrawClaimReq := newJSONRequest(t, http.MethodPost, "/v1/market/claims/claim_w/withdraw", "human:ten_01:user_exec:executor", "idem_m9_w_withdraw", map[string]any{"reason": "not available"})
	rrWithdrawClaim := httptest.NewRecorder()
	h.ServeHTTP(rrWithdrawClaim, withdrawClaimReq)
	if rrWithdrawClaim.Code != http.StatusOK {
		t.Fatalf("withdraw claim expected 200 got %d body=%s", rrWithdrawClaim.Code, rrWithdrawClaim.Body.String())
	}

	createPayoutAccountReq := newJSONRequest(t, http.MethodPost, "/v1/payout-accounts", "human:ten_01:user_exec:executor", "idem_m9_w_payout_acct", map[string]any{
		"payout_account_id":    "pa_w",
		"owner_profile_id":     "mp_exec_w",
		"provider":             "stripe",
		"provider_account_ref": "acct_w",
		"jurisdiction":         "AU",
	})
	rrCreatePayoutAccount := httptest.NewRecorder()
	h.ServeHTTP(rrCreatePayoutAccount, createPayoutAccountReq)
	if rrCreatePayoutAccount.Code != http.StatusOK {
		t.Fatalf("create payout account expected 200 got %d body=%s", rrCreatePayoutAccount.Code, rrCreatePayoutAccount.Body.String())
	}

	getPayoutAccountReq := httptest.NewRequest(http.MethodGet, "/v1/payout-accounts/pa_w", nil)
	getPayoutAccountReq.Header.Set("Authorization", "Bearer human:ten_01:user_exec:executor")
	rrGetPayoutAccount := httptest.NewRecorder()
	h.ServeHTTP(rrGetPayoutAccount, getPayoutAccountReq)
	if rrGetPayoutAccount.Code != http.StatusOK {
		t.Fatalf("get payout account expected 200 got %d body=%s", rrGetPayoutAccount.Code, rrGetPayoutAccount.Body.String())
	}

	getMissingPayoutAccountReq := httptest.NewRequest(http.MethodGet, "/v1/payout-accounts/missing", nil)
	getMissingPayoutAccountReq.Header.Set("Authorization", "Bearer human:ten_01:user_exec:executor")
	rrGetMissingPayoutAccount := httptest.NewRecorder()
	h.ServeHTTP(rrGetMissingPayoutAccount, getMissingPayoutAccountReq)
	if rrGetMissingPayoutAccount.Code != http.StatusNotFound {
		t.Fatalf("missing payout account expected 404 got %d body=%s", rrGetMissingPayoutAccount.Code, rrGetMissingPayoutAccount.Body.String())
	}

	getPayoutAccountForbiddenReq := httptest.NewRequest(http.MethodGet, "/v1/payout-accounts/pa_w", nil)
	getPayoutAccountForbiddenReq.Header.Set("Authorization", "Bearer human:ten_01:user_dev:developer")
	rrGetPayoutAccountForbidden := httptest.NewRecorder()
	h.ServeHTTP(rrGetPayoutAccountForbidden, getPayoutAccountForbiddenReq)
	if rrGetPayoutAccountForbidden.Code != http.StatusForbidden {
		t.Fatalf("forbidden payout account read expected 403 got %d body=%s", rrGetPayoutAccountForbidden.Code, rrGetPayoutAccountForbidden.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestM9EscrowPayoutAndDisputeErrorPaths(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	createRequesterReq := newJSONRequest(t, http.MethodPost, "/v1/market/profiles", "human:ten_01:user_admin:tenant_admin", "idem_m9_e_profile_req", map[string]any{
		"profile_id":   "mp_req_e",
		"profile_type": "requester",
		"display_name": "Requester E",
	})
	rrCreateRequester := httptest.NewRecorder()
	h.ServeHTTP(rrCreateRequester, createRequesterReq)
	if rrCreateRequester.Code != http.StatusOK {
		t.Fatalf("create requester profile expected 200 got %d body=%s", rrCreateRequester.Code, rrCreateRequester.Body.String())
	}

	createExecutorReq := newJSONRequest(t, http.MethodPost, "/v1/market/profiles", "human:ten_01:user_admin:tenant_admin", "idem_m9_e_profile_exec", map[string]any{
		"profile_id":   "mp_exec_e",
		"profile_type": "executor",
		"display_name": "Executor E",
	})
	rrCreateExecutor := httptest.NewRecorder()
	h.ServeHTTP(rrCreateExecutor, createExecutorReq)
	if rrCreateExecutor.Code != http.StatusOK {
		t.Fatalf("create executor profile expected 200 got %d body=%s", rrCreateExecutor.Code, rrCreateExecutor.Body.String())
	}

	createFundingReq := newJSONRequest(t, http.MethodPost, "/v1/market/funding-accounts", "human:ten_01:user_admin:tenant_admin", "idem_m9_e_funding", map[string]any{
		"account_id":       "fund_e",
		"owner_profile_id": "mp_req_e",
		"currency":         "JWUSD",
		"initial_balance":  120,
	})
	rrCreateFunding := httptest.NewRecorder()
	h.ServeHTTP(rrCreateFunding, createFundingReq)
	if rrCreateFunding.Code != http.StatusOK {
		t.Fatalf("create funding expected 200 got %d body=%s", rrCreateFunding.Code, rrCreateFunding.Body.String())
	}

	createListingReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings", "human:ten_01:user_admin:tenant_admin", "idem_m9_e_listing", map[string]any{
		"listing_id":           "ml_e",
		"task_id":              "task_e",
		"requester_profile_id": "mp_req_e",
		"work_class":           "public_open",
		"listing_mode":         "fixed_price_open_claim",
		"budget_total":         70,
		"currency":             "JWUSD",
		"funding_account_id":   "fund_e",
	})
	rrCreateListing := httptest.NewRecorder()
	h.ServeHTTP(rrCreateListing, createListingReq)
	if rrCreateListing.Code != http.StatusOK {
		t.Fatalf("create listing expected 200 got %d body=%s", rrCreateListing.Code, rrCreateListing.Body.String())
	}

	fundBeforePublishReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings/ml_e/fund", "human:ten_01:user_admin:tenant_admin", "idem_m9_e_fund_early", map[string]any{"amount": 10})
	rrFundBeforePublish := httptest.NewRecorder()
	h.ServeHTTP(rrFundBeforePublish, fundBeforePublishReq)
	if rrFundBeforePublish.Code != http.StatusConflict {
		t.Fatalf("fund unpublished listing expected 409 got %d body=%s", rrFundBeforePublish.Code, rrFundBeforePublish.Body.String())
	}

	publishListingReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings/ml_e/publish", "human:ten_01:user_admin:tenant_admin", "idem_m9_e_publish", map[string]any{"reason": "ready"})
	rrPublishListing := httptest.NewRecorder()
	h.ServeHTTP(rrPublishListing, publishListingReq)
	if rrPublishListing.Code != http.StatusOK {
		t.Fatalf("publish listing expected 200 got %d body=%s", rrPublishListing.Code, rrPublishListing.Body.String())
	}

	fundReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings/ml_e/fund", "human:ten_01:user_admin:tenant_admin", "idem_m9_e_fund", map[string]any{"amount": 15})
	rrFund := httptest.NewRecorder()
	h.ServeHTTP(rrFund, fundReq)
	if rrFund.Code != http.StatusOK {
		t.Fatalf("fund listing expected 200 got %d body=%s", rrFund.Code, rrFund.Body.String())
	}
	escrowID := jsonValue(t, rrFund.Body.Bytes(), "escrow.escrow_id")
	if escrowID == "" {
		t.Fatalf("missing escrow id in fund response body=%s", rrFund.Body.String())
	}

	getEscrowForbiddenReq := httptest.NewRequest(http.MethodGet, "/v1/market/escrows/"+escrowID, nil)
	getEscrowForbiddenReq.Header.Set("Authorization", "Bearer human:ten_01:user_dev:developer")
	rrGetEscrowForbidden := httptest.NewRecorder()
	h.ServeHTTP(rrGetEscrowForbidden, getEscrowForbiddenReq)
	if rrGetEscrowForbidden.Code != http.StatusForbidden {
		t.Fatalf("forbidden escrow read expected 403 got %d body=%s", rrGetEscrowForbidden.Code, rrGetEscrowForbidden.Body.String())
	}

	releaseBadReq := httptest.NewRequest(http.MethodPost, "/v1/market/escrows/"+escrowID+"/release", strings.NewReader("{"))
	releaseBadReq.Header.Set("Authorization", "Bearer human:ten_01:user_app:approver")
	releaseBadReq.Header.Set("Idempotency-Key", "idem_m9_e_release_bad")
	releaseBadReq.Header.Set("Content-Type", "application/json")
	rrReleaseBad := httptest.NewRecorder()
	h.ServeHTTP(rrReleaseBad, releaseBadReq)
	if rrReleaseBad.Code != http.StatusBadRequest {
		t.Fatalf("invalid release json expected 400 got %d body=%s", rrReleaseBad.Code, rrReleaseBad.Body.String())
	}

	releaseExcessReq := newJSONRequest(t, http.MethodPost, "/v1/market/escrows/"+escrowID+"/release", "human:ten_01:user_app:approver", "idem_m9_e_release_excess", map[string]any{"amount": 20})
	rrReleaseExcess := httptest.NewRecorder()
	h.ServeHTTP(rrReleaseExcess, releaseExcessReq)
	if rrReleaseExcess.Code != http.StatusBadRequest {
		t.Fatalf("release excess expected 400 got %d body=%s", rrReleaseExcess.Code, rrReleaseExcess.Body.String())
	}

	createPayoutAccountReq := newJSONRequest(t, http.MethodPost, "/v1/payout-accounts", "human:ten_01:user_exec:executor", "idem_m9_e_payout_acct", map[string]any{
		"payout_account_id":    "pa_e",
		"owner_profile_id":     "mp_exec_e",
		"provider":             "stripe",
		"provider_account_ref": "acct_e",
		"jurisdiction":         "AU",
	})
	rrCreatePayoutAccount := httptest.NewRecorder()
	h.ServeHTTP(rrCreatePayoutAccount, createPayoutAccountReq)
	if rrCreatePayoutAccount.Code != http.StatusOK {
		t.Fatalf("create payout account expected 200 got %d body=%s", rrCreatePayoutAccount.Code, rrCreatePayoutAccount.Body.String())
	}

	createPayoutMissingEscrowReq := newJSONRequest(t, http.MethodPost, "/v1/payouts", "human:ten_01:user_exec:executor", "idem_m9_e_payout_missing_escrow", map[string]any{
		"payout_id":         "po_e_missing",
		"payout_account_id": "pa_e",
		"escrow_id":         "escrow_missing",
		"amount":            5,
		"currency":          "JWUSD",
	})
	rrCreatePayoutMissingEscrow := httptest.NewRecorder()
	h.ServeHTTP(rrCreatePayoutMissingEscrow, createPayoutMissingEscrowReq)
	if rrCreatePayoutMissingEscrow.Code != http.StatusNotFound {
		t.Fatalf("missing escrow payout expected 404 got %d body=%s", rrCreatePayoutMissingEscrow.Code, rrCreatePayoutMissingEscrow.Body.String())
	}

	openDisputeForbiddenReq := newJSONRequest(t, http.MethodPost, "/v1/market/disputes", "human:ten_01:user_dev:developer", "idem_m9_e_dispute_forbidden", map[string]any{
		"dispute_id":   "disp_e",
		"listing_id":   "ml_e",
		"escrow_id":    escrowID,
		"dispute_type": "acceptance_disagreement",
		"reason":       "forbidden role",
	})
	rrOpenDisputeForbidden := httptest.NewRecorder()
	h.ServeHTTP(rrOpenDisputeForbidden, openDisputeForbiddenReq)
	if rrOpenDisputeForbidden.Code != http.StatusForbidden {
		t.Fatalf("forbidden dispute open expected 403 got %d body=%s", rrOpenDisputeForbidden.Code, rrOpenDisputeForbidden.Body.String())
	}

	openDisputeReq := newJSONRequest(t, http.MethodPost, "/v1/market/disputes", "human:ten_01:user_req:requester", "idem_m9_e_dispute_open", map[string]any{
		"dispute_id":   "disp_e",
		"listing_id":   "ml_e",
		"escrow_id":    escrowID,
		"dispute_type": "acceptance_disagreement",
		"reason":       "need review",
	})
	rrOpenDispute := httptest.NewRecorder()
	h.ServeHTTP(rrOpenDispute, openDisputeReq)
	if rrOpenDispute.Code != http.StatusOK {
		t.Fatalf("open dispute expected 200 got %d body=%s", rrOpenDispute.Code, rrOpenDispute.Body.String())
	}

	appealBadReq := httptest.NewRequest(http.MethodPost, "/v1/market/disputes/disp_e/appeal", strings.NewReader("{"))
	appealBadReq.Header.Set("Authorization", "Bearer human:ten_01:user_exec:executor")
	appealBadReq.Header.Set("Idempotency-Key", "idem_m9_e_appeal_bad")
	appealBadReq.Header.Set("Content-Type", "application/json")
	rrAppealBad := httptest.NewRecorder()
	h.ServeHTTP(rrAppealBad, appealBadReq)
	if rrAppealBad.Code != http.StatusBadRequest {
		t.Fatalf("invalid appeal json expected 400 got %d body=%s", rrAppealBad.Code, rrAppealBad.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestM9MarketRolloutGate(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	dashboardReq := httptest.NewRequest(http.MethodGet, "/v1/market/dashboard", nil)
	dashboardReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrDashboard := httptest.NewRecorder()
	h.ServeHTTP(rrDashboard, dashboardReq)
	if rrDashboard.Code != http.StatusOK {
		t.Fatalf("initial dashboard expected 200 got %d body=%s", rrDashboard.Code, rrDashboard.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(rrDashboard.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode initial dashboard payload: %v", err)
	}
	dashboard, _ := payload["market_dashboard"].(map[string]any)
	rolloutGate, _ := dashboard["rollout_gate"].(map[string]any)
	if rolloutGate == nil {
		t.Fatalf("missing rollout_gate payload: %s", rrDashboard.Body.String())
	}
	if ready, _ := rolloutGate["ready"].(bool); ready {
		t.Fatalf("expected rollout gate to be false before readiness")
	}

	ctx := context.Background()
	if err := srv.simulation.EnsureV1ScenarioLibrary(ctx, "ten_01"); err != nil {
		t.Fatalf("seed simulation library: %v", err)
	}
	for idx, scenarioID := range marketRolloutScenarioIDs {
		run, err := srv.simulation.StartRun(ctx, simulationsvc.StartRunRequest{
			TenantID:        "ten_01",
			ScenarioID:      scenarioID,
			ScenarioVersion: 1,
			RunMode:         simulationsvc.RunModeDeterministicStub,
			Seed:            "gate_seed_" + string(rune('a'+idx)),
		})
		if err != nil {
			t.Fatalf("start rollout scenario run %s: %v", scenarioID, err)
		}
		if _, err := srv.simulation.PromoteBaseline(ctx, simulationsvc.PromoteBaselineRequest{
			TenantID:   "ten_01",
			RunID:      run.RunID,
			Reason:     "market_rollout_gate",
			PromotedBy: "user_admin",
		}); err != nil {
			t.Fatalf("promote baseline for %s: %v", scenarioID, err)
		}
	}

	if _, err := srv.payouts.CreateAccount(ctx, payoutssvc.CreateAccountRequest{
		TenantID:           "ten_01",
		PayoutAccountID:    "pa_gate",
		OwnerProfileID:     "mp_exec_gate",
		Provider:           "stripe",
		ProviderAccountRef: "acct_gate",
		Jurisdiction:       "AU",
	}); err != nil {
		t.Fatalf("create payout account readiness seed: %v", err)
	}
	if _, err := srv.payouts.RequestPayout(ctx, payoutssvc.RequestPayoutRequest{
		TenantID:        "ten_01",
		PayoutID:        "po_gate",
		PayoutAccountID: "pa_gate",
		Amount:          10,
		Currency:        "JWUSD",
		AutoComplete:    true,
	}); err != nil {
		t.Fatalf("request payout readiness seed: %v", err)
	}

	dispute, err := srv.disputes.OpenDispute(ctx, disputessvc.OpenDisputeRequest{
		TenantID:    "ten_01",
		DisputeID:   "disp_gate",
		ListingID:   "ml_gate",
		DisputeType: disputessvc.TypeRequesterDefault,
		Reason:      "rollout readiness",
		OpenedBy:    "user_req",
	})
	if err != nil {
		t.Fatalf("open dispute readiness seed: %v", err)
	}
	if _, err := srv.disputes.ResolveDispute(ctx, disputessvc.ResolveDisputeRequest{
		TenantID:      "ten_01",
		DisputeID:     dispute.DisputeID,
		Decision:      "resolved",
		AppealAllowed: false,
	}); err != nil {
		t.Fatalf("resolve dispute readiness seed: %v", err)
	}

	rrDashboardReady := httptest.NewRecorder()
	h.ServeHTTP(rrDashboardReady, dashboardReq)
	if rrDashboardReady.Code != http.StatusOK {
		t.Fatalf("ready dashboard expected 200 got %d body=%s", rrDashboardReady.Code, rrDashboardReady.Body.String())
	}
	var readyPayload map[string]any
	if err := json.Unmarshal(rrDashboardReady.Body.Bytes(), &readyPayload); err != nil {
		t.Fatalf("decode ready dashboard payload: %v", err)
	}
	readyDashboard, _ := readyPayload["market_dashboard"].(map[string]any)
	readyRolloutGate, _ := readyDashboard["rollout_gate"].(map[string]any)
	if readyRolloutGate == nil {
		t.Fatalf("missing ready rollout gate payload: %s", rrDashboardReady.Body.String())
	}
	if ready, _ := readyRolloutGate["ready"].(bool); !ready {
		t.Fatalf("expected rollout gate to be ready after prerequisites, payload=%s", rrDashboardReady.Body.String())
	}
	requiredScenarios, _ := readyRolloutGate["required_scenarios"].([]any)
	if len(requiredScenarios) != len(marketRolloutScenarioIDs) {
		t.Fatalf("required scenario count=%d want %d", len(requiredScenarios), len(marketRolloutScenarioIDs))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestM9SealedWorkSybilGuard(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	createRequesterReq := newJSONRequest(t, http.MethodPost, "/v1/market/profiles", "human:ten_01:user_admin:tenant_admin", "idem_m9_sybil_req", map[string]any{
		"profile_id":           "mp_req_sybil",
		"profile_type":         "requester",
		"display_name":         "Requester Sybil",
		"verification_status":  "verified",
		"work_class_allowlist": []string{"public_sealed"},
	})
	rrCreateRequester := httptest.NewRecorder()
	h.ServeHTTP(rrCreateRequester, createRequesterReq)
	if rrCreateRequester.Code != http.StatusOK {
		t.Fatalf("create requester expected 200 got %d body=%s", rrCreateRequester.Code, rrCreateRequester.Body.String())
	}

	createExecutorReq := newJSONRequest(t, http.MethodPost, "/v1/market/profiles", "human:ten_01:user_admin:tenant_admin", "idem_m9_sybil_exec", map[string]any{
		"profile_id":   "mp_exec_sybil",
		"profile_type": "executor",
		"display_name": "Executor Sybil",
	})
	rrCreateExecutor := httptest.NewRecorder()
	h.ServeHTTP(rrCreateExecutor, createExecutorReq)
	if rrCreateExecutor.Code != http.StatusOK {
		t.Fatalf("create executor expected 200 got %d body=%s", rrCreateExecutor.Code, rrCreateExecutor.Body.String())
	}

	createFundingReq := newJSONRequest(t, http.MethodPost, "/v1/market/funding-accounts", "human:ten_01:user_admin:tenant_admin", "idem_m9_sybil_fund", map[string]any{
		"account_id":       "fund_sybil",
		"owner_profile_id": "mp_req_sybil",
		"currency":         "JWUSD",
		"initial_balance":  200,
	})
	rrCreateFunding := httptest.NewRecorder()
	h.ServeHTTP(rrCreateFunding, createFundingReq)
	if rrCreateFunding.Code != http.StatusOK {
		t.Fatalf("create funding expected 200 got %d body=%s", rrCreateFunding.Code, rrCreateFunding.Body.String())
	}

	createListingReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings", "human:ten_01:user_admin:tenant_admin", "idem_m9_sybil_listing", map[string]any{
		"listing_id":           "ml_sybil",
		"task_id":              "task_sybil",
		"requester_profile_id": "mp_req_sybil",
		"work_class":           "public_sealed",
		"listing_mode":         "fixed_price_open_claim",
		"budget_total":         50,
		"currency":             "JWUSD",
		"funding_account_id":   "fund_sybil",
	})
	rrCreateListing := httptest.NewRecorder()
	h.ServeHTTP(rrCreateListing, createListingReq)
	if rrCreateListing.Code != http.StatusOK {
		t.Fatalf("create listing expected 200 got %d body=%s", rrCreateListing.Code, rrCreateListing.Body.String())
	}

	publishListingReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings/ml_sybil/publish", "human:ten_01:user_admin:tenant_admin", "idem_m9_sybil_publish", map[string]any{"reason": "ready"})
	rrPublishListing := httptest.NewRecorder()
	h.ServeHTTP(rrPublishListing, publishListingReq)
	if rrPublishListing.Code != http.StatusOK {
		t.Fatalf("publish listing expected 200 got %d body=%s", rrPublishListing.Code, rrPublishListing.Body.String())
	}

	createClaimReq := newJSONRequest(t, http.MethodPost, "/v1/market/listings/ml_sybil/claims", "human:ten_01:user_exec:executor", "idem_m9_sybil_claim", map[string]any{
		"claim_id":            "claim_sybil",
		"executor_profile_id": "mp_exec_sybil",
	})
	rrCreateClaim := httptest.NewRecorder()
	h.ServeHTTP(rrCreateClaim, createClaimReq)
	if rrCreateClaim.Code != http.StatusForbidden || !strings.Contains(rrCreateClaim.Body.String(), "market_identity_verification_required") {
		t.Fatalf("sealed-work sybil guard expected 403 with code got %d body=%s", rrCreateClaim.Code, rrCreateClaim.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
