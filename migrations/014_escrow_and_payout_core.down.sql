DROP INDEX IF EXISTS idx_payout_requests_status;
DROP INDEX IF EXISTS idx_task_escrows_listing_status;
DROP INDEX IF EXISTS idx_funding_accounts_tenant_owner;

DROP TABLE IF EXISTS payout_transfers;
DROP TABLE IF EXISTS payout_requests;
DROP TABLE IF EXISTS payout_accounts;
DROP TABLE IF EXISTS escrow_transfers;
DROP TABLE IF EXISTS task_escrows;
DROP TABLE IF EXISTS funding_accounts;
