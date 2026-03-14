package escrow

import (
	"context"
	"errors"
	"testing"
)

func TestEscrowLifecycle(t *testing.T) {
	svc := New()
	ctx := context.Background()

	escrow, err := svc.EnsureEscrow(ctx, EnsureEscrowRequest{
		TenantID:         "ten_01",
		ListingID:        "ml_01",
		FundingAccountID: "fund_01",
		Currency:         "jwusd",
	})
	if err != nil {
		t.Fatalf("ensure escrow: %v", err)
	}
	if escrow.Status != StatusPendingLock {
		t.Fatalf("escrow status=%q want %q", escrow.Status, StatusPendingLock)
	}

	escrowAgain, err := svc.EnsureEscrow(ctx, EnsureEscrowRequest{
		TenantID:         "ten_01",
		ListingID:        "ml_01",
		FundingAccountID: "fund_01",
		Currency:         "JWUSD",
	})
	if err != nil || escrowAgain.EscrowID != escrow.EscrowID {
		t.Fatalf("ensure escrow idempotent expected same escrow got %+v err=%v", escrowAgain, err)
	}

	locked, transfer, err := svc.Lock(ctx, AdjustEscrowRequest{
		TenantID:     "ten_01",
		EscrowID:     escrow.EscrowID,
		Amount:       120,
		TransferType: TransferLock,
	})
	if err != nil {
		t.Fatalf("lock escrow: %v", err)
	}
	if locked.TotalLocked != 120 || transfer.TransferType != TransferLock {
		t.Fatalf("unexpected lock state escrow=%+v transfer=%+v", locked, transfer)
	}

	released, _, err := svc.Release(ctx, AdjustEscrowRequest{
		TenantID:     "ten_01",
		EscrowID:     escrow.EscrowID,
		Amount:       70,
		TransferType: TransferRelease,
	})
	if err != nil {
		t.Fatalf("release escrow: %v", err)
	}
	if released.ReleasedAmount != 70 {
		t.Fatalf("released_amount=%f want 70", released.ReleasedAmount)
	}

	refunded, _, err := svc.Refund(ctx, AdjustEscrowRequest{
		TenantID:     "ten_01",
		EscrowID:     escrow.EscrowID,
		Amount:       50,
		TransferType: TransferRefund,
	})
	if err != nil {
		t.Fatalf("refund escrow: %v", err)
	}
	if refunded.RefundedAmount != 50 {
		t.Fatalf("refunded_amount=%f want 50", refunded.RefundedAmount)
	}

	transfers, err := svc.ListTransfers(ctx, "ten_01", escrow.EscrowID)
	if err != nil {
		t.Fatalf("list transfers: %v", err)
	}
	if len(transfers) != 3 {
		t.Fatalf("expected 3 transfers, got %d", len(transfers))
	}
}

func TestEscrowErrorBranches(t *testing.T) {
	svc := New()
	ctx := context.Background()

	if _, err := svc.EnsureEscrow(ctx, EnsureEscrowRequest{
		TenantID:         "",
		ListingID:        "ml_01",
		FundingAccountID: "fund_01",
		Currency:         "JWUSD",
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid request for missing tenant, got %v", err)
	}

	if _, _, err := svc.Lock(ctx, AdjustEscrowRequest{
		TenantID:     "ten_01",
		EscrowID:     "missing",
		Amount:       10,
		TransferType: TransferLock,
	}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found lock error, got %v", err)
	}

	escrow, err := svc.EnsureEscrow(ctx, EnsureEscrowRequest{
		TenantID:         "ten_01",
		ListingID:        "ml_02",
		FundingAccountID: "fund_01",
		Currency:         "JWUSD",
	})
	if err != nil {
		t.Fatalf("ensure escrow: %v", err)
	}

	if _, _, err := svc.Release(ctx, AdjustEscrowRequest{
		TenantID:     "ten_01",
		EscrowID:     escrow.EscrowID,
		Amount:       10,
		TransferType: TransferRelease,
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid request release exceeds locked, got %v", err)
	}

	if _, _, err := svc.Lock(ctx, AdjustEscrowRequest{
		TenantID:     "ten_01",
		EscrowID:     escrow.EscrowID,
		Amount:       20,
		TransferType: TransferLock,
	}); err != nil {
		t.Fatalf("lock escrow: %v", err)
	}

	if _, _, err := svc.Refund(ctx, AdjustEscrowRequest{
		TenantID:     "ten_01",
		EscrowID:     escrow.EscrowID,
		Amount:       25,
		TransferType: TransferRefund,
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid refund exceed error, got %v", err)
	}

	if _, err := svc.GetEscrow(ctx, "ten_02", escrow.EscrowID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected tenant isolation not found, got %v", err)
	}
	if _, err := svc.GetEscrowByListing(ctx, "ten_01", "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing listing escrow, got %v", err)
	}
}

func TestEscrowListReadAndTransferValidation(t *testing.T) {
	svc := New()
	ctx := context.Background()

	meta := map[string]any{"source": "m9"}
	first, err := svc.EnsureEscrow(ctx, EnsureEscrowRequest{
		TenantID:         "ten_01",
		ListingID:        "ml_a",
		FundingAccountID: "fund_a",
		Currency:         "JWUSD",
		Metadata:         meta,
	})
	if err != nil {
		t.Fatalf("ensure first escrow: %v", err)
	}
	meta["source"] = "mutated"
	if first.Metadata["source"] != "m9" {
		t.Fatalf("expected metadata copy to be isolated, got %+v", first.Metadata)
	}

	second, err := svc.EnsureEscrow(ctx, EnsureEscrowRequest{
		TenantID:         "ten_01",
		ListingID:        "ml_b",
		FundingAccountID: "fund_b",
		Currency:         "JWUSD",
	})
	if err != nil {
		t.Fatalf("ensure second escrow: %v", err)
	}

	if _, err := svc.GetEscrowByListing(ctx, "ten_01", "ml_a"); err != nil {
		t.Fatalf("get escrow by listing: %v", err)
	}

	if _, _, err := svc.Lock(ctx, AdjustEscrowRequest{
		TenantID:     "ten_01",
		EscrowID:     first.EscrowID,
		Amount:       5,
		TransferType: TransferRelease,
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected transfer type mismatch error, got %v", err)
	}

	if _, _, err := svc.Lock(ctx, AdjustEscrowRequest{
		TenantID:     "ten_01",
		EscrowID:     first.EscrowID,
		Amount:       10,
		TransferType: TransferLock,
	}); err != nil {
		t.Fatalf("lock first escrow: %v", err)
	}

	transfers, err := svc.ListTransfers(ctx, "ten_01", first.EscrowID)
	if err != nil {
		t.Fatalf("list transfers: %v", err)
	}
	if len(transfers) != 1 {
		t.Fatalf("expected one transfer, got %d", len(transfers))
	}
	if _, err := svc.ListTransfers(ctx, "ten_01", "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing escrow list transfer error, got %v", err)
	}

	all := svc.ListEscrows(ctx, "ten_01")
	if len(all) != 2 {
		t.Fatalf("expected two escrows, got %d", len(all))
	}
	if all[0].EscrowID != second.EscrowID {
		t.Fatalf("expected most recent escrow first, got order %+v", all)
	}
	if len(svc.ListEscrows(ctx, "ten_02")) != 0 {
		t.Fatalf("expected tenant isolated escrow list")
	}
}
