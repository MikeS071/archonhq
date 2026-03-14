package disputes

import (
	"context"
	"errors"
	"testing"
)

func TestDisputeLifecycle(t *testing.T) {
	svc := New()
	ctx := context.Background()

	dispute, err := svc.OpenDispute(ctx, OpenDisputeRequest{
		TenantID:    "ten_01",
		DisputeID:   "disp_01",
		ListingID:   "ml_01",
		EscrowID:    "escrow_01",
		ClaimID:     "claim_01",
		DisputeType: TypeAcceptanceDisagree,
		Reason:      "acceptance mismatch",
		OpenedBy:    "user_requester",
	})
	if err != nil {
		t.Fatalf("open dispute: %v", err)
	}
	if dispute.Status != StatusOpen {
		t.Fatalf("status=%q want %q", dispute.Status, StatusOpen)
	}

	resolved, err := svc.ResolveDispute(ctx, ResolveDisputeRequest{
		TenantID:            "ten_01",
		DisputeID:           dispute.DisputeID,
		Decision:            "split_refund",
		FeeShift:            0.15,
		EscrowReleaseAction: EscrowActionRefund,
		AppealAllowed:       true,
	})
	if err != nil {
		t.Fatalf("resolve dispute: %v", err)
	}
	if resolved.Status != StatusResolved {
		t.Fatalf("status=%q want %q", resolved.Status, StatusResolved)
	}

	appealed, err := svc.AppealDispute(ctx, AppealDisputeRequest{
		TenantID:  "ten_01",
		DisputeID: dispute.DisputeID,
		AppealBy:  "user_executor",
		Reason:    "new evidence",
	})
	if err != nil {
		t.Fatalf("appeal dispute: %v", err)
	}
	if appealed.Status != StatusAppealed {
		t.Fatalf("status=%q want %q", appealed.Status, StatusAppealed)
	}

	if _, err := svc.GetDispute(ctx, "ten_01", dispute.DisputeID); err != nil {
		t.Fatalf("get dispute: %v", err)
	}
	if len(svc.ListDisputes(ctx, "ten_01")) != 1 {
		t.Fatalf("expected one dispute in list")
	}
}

func TestDisputeErrorPaths(t *testing.T) {
	svc := New()
	ctx := context.Background()

	if _, err := svc.OpenDispute(ctx, OpenDisputeRequest{
		TenantID:    "ten_01",
		ListingID:   "ml_01",
		DisputeType: "bad_type",
		Reason:      "x",
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid dispute type error, got %v", err)
	}

	dispute, err := svc.OpenDispute(ctx, OpenDisputeRequest{
		TenantID:    "ten_01",
		DisputeID:   "disp_02",
		ListingID:   "ml_02",
		DisputeType: TypeNonDelivery,
		Reason:      "not delivered",
	})
	if err != nil {
		t.Fatalf("open dispute: %v", err)
	}

	if _, err := svc.AppealDispute(ctx, AppealDisputeRequest{
		TenantID:  "ten_01",
		DisputeID: dispute.DisputeID,
		Reason:    "too early",
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid appeal-before-resolution error, got %v", err)
	}

	if _, err := svc.ResolveDispute(ctx, ResolveDisputeRequest{
		TenantID:            "ten_01",
		DisputeID:           dispute.DisputeID,
		Decision:            "deny",
		EscrowReleaseAction: "bad_action",
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid escrow action error, got %v", err)
	}

	resolved, err := svc.ResolveDispute(ctx, ResolveDisputeRequest{
		TenantID:      "ten_01",
		DisputeID:     dispute.DisputeID,
		Decision:      "deny",
		AppealAllowed: false,
	})
	if err != nil {
		t.Fatalf("resolve dispute: %v", err)
	}
	if _, err := svc.AppealDispute(ctx, AppealDisputeRequest{
		TenantID:  "ten_01",
		DisputeID: resolved.DisputeID,
		Reason:    "not allowed",
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected appeal not allowed error, got %v", err)
	}

	if _, err := svc.GetDispute(ctx, "ten_02", dispute.DisputeID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected tenant isolation not found error, got %v", err)
	}
}

func TestDisputeAutoIDAndReputationAdjustmentCopy(t *testing.T) {
	svc := New()
	ctx := context.Background()

	dispute, err := svc.OpenDispute(ctx, OpenDisputeRequest{
		TenantID:    "ten_01",
		ListingID:   "ml_auto",
		DisputeType: TypeSpecDrift,
		Reason:      "drift detected",
	})
	if err != nil {
		t.Fatalf("open dispute with auto id: %v", err)
	}
	if dispute.DisputeID == "" || dispute.DisputeID[:5] != "disp_" {
		t.Fatalf("expected auto dispute id, got %q", dispute.DisputeID)
	}

	adjustment := map[string]any{"reputation_delta": -0.2}
	resolved, err := svc.ResolveDispute(ctx, ResolveDisputeRequest{
		TenantID:             "ten_01",
		DisputeID:            dispute.DisputeID,
		Decision:             "partial_refund",
		EscrowReleaseAction:  EscrowActionRefund,
		ReputationAdjustment: adjustment,
		AppealAllowed:        true,
	})
	if err != nil {
		t.Fatalf("resolve dispute with reputation adjustment: %v", err)
	}
	adjustment["reputation_delta"] = -0.9
	if got := resolved.ReputationAdjustment["reputation_delta"]; got != -0.2 {
		t.Fatalf("expected copied reputation adjustment, got %v", got)
	}

	if _, err := svc.ResolveDispute(ctx, ResolveDisputeRequest{
		TenantID:  "ten_01",
		DisputeID: dispute.DisputeID,
		Decision:  "retry",
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected resolve on resolved dispute to fail, got %v", err)
	}
}
