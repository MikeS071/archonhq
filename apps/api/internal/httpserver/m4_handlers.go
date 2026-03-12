package httpserver

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/MikeS071/archonhq/pkg/apierrors"
	"github.com/MikeS071/archonhq/pkg/pricingengine"
	"github.com/MikeS071/archonhq/pkg/scoring"
	"github.com/MikeS071/archonhq/pkg/settlement"
)

type pricingQuoteRequest struct {
	QuoteID          string  `json:"quote_id,omitempty"`
	TaskID           string  `json:"task_id"`
	BaseRate         float64 `json:"base_rate"`
	BidAdjustment    float64 `json:"bid_adjustment"`
	PredictedRawJW   float64 `json:"predicted_raw_jw"`
	QualityFactor    float64 `json:"quality_factor"`
	ReliabilityRF    float64 `json:"reliability_rf"`
	ReserveRatio     float64 `json:"reserve_ratio"`
	StrategyOverride string  `json:"strategy_override,omitempty"`
}

func (s *Server) handlePricingQuoteV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "operator", "approver", "finance_viewer") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for pricing quote.", corrID, nil)
		return
	}

	var req pricingQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	req.TaskID = strings.TrimSpace(req.TaskID)
	if req.TaskID == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "task_id is required.", corrID, nil)
		return
	}

	const taskQ = "SELECT task_id, task_family FROM tasks WHERE task_id = $1 AND tenant_id = $2"
	var taskID, taskFamily string
	if err := s.postgres.DB.QueryRowContext(r.Context(), taskQ, req.TaskID, actor.TenantID).Scan(&taskID, &taskFamily); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "task_not_found", "Task not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "task_lookup_failed", "Failed to validate task.", corrID, nil)
		return
	}

	strategy := pricingengine.ResolveStrategy(req.StrategyOverride)
	quote, err := strategy.Quote(pricingengine.QuoteRequest{
		TaskID:         req.TaskID,
		BaseRate:       req.BaseRate,
		BidAdjustment:  req.BidAdjustment,
		PredictedRawJW: req.PredictedRawJW,
		QualityFactor:  req.QualityFactor,
		ReliabilityRF:  req.ReliabilityRF,
		ReserveRatio:   req.ReserveRatio,
	})
	if err != nil {
		apierrors.Write(w, http.StatusBadRequest, "rate_resolution_failed", "Failed to compute pricing quote.", corrID, map[string]any{"reason": err.Error()})
		return
	}

	quoteID := strings.TrimSpace(req.QuoteID)
	if quoteID == "" {
		quoteID = "quote_" + randomID(6)
	}
	quoteJSON, _ := json.Marshal(quote)

	const insertQ = "INSERT INTO price_quotes (quote_id, tenant_id, task_id, strategy_name, quote_json) VALUES ($1,$2,$3,$4,$5)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), insertQ, quoteID, actor.TenantID, req.TaskID, strategy.Name(), quoteJSON); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "pricing_quote_failed", "Failed to persist price quote.", corrID, nil)
		return
	}

	s.appendEvent(r, actor.TenantID, "pricing", quoteID, "pricing.quote_created", map[string]any{
		"task_id":       req.TaskID,
		"task_family":   taskFamily,
		"strategy_name": strategy.Name(),
		"rate_value":    quote.RateValue,
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"quote_id":       quoteID,
		"task_id":        req.TaskID,
		"task_family":    taskFamily,
		"strategy_name":  strategy.Name(),
		"quote":          quote,
		"correlation_id": corrID,
	})
}

func (s *Server) handlePricingRateCardsV2(w http.ResponseWriter, r *http.Request) {
	_, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"rate_cards":     pricingengine.DefaultRateCards(),
		"correlation_id": corrID,
	})
}

type pricingBidRequest struct {
	RateSnapshotID string  `json:"rate_snapshot_id,omitempty"`
	TaskID         string  `json:"task_id"`
	ResultID       string  `json:"result_id,omitempty"`
	BaseRate       float64 `json:"base_rate"`
	BidAdjustment  float64 `json:"bid_adjustment"`
	QualityFactor  float64 `json:"quality_factor"`
	ReliabilityRF  float64 `json:"reliability_rf"`
	ReserveRatio   float64 `json:"reserve_ratio"`
}

func (s *Server) handlePricingBidsV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "operator", "approver", "finance_viewer") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for pricing bids.", corrID, nil)
		return
	}

	var req pricingBidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	req.TaskID = strings.TrimSpace(req.TaskID)
	if req.TaskID == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "task_id is required.", corrID, nil)
		return
	}

	const taskQ = "SELECT task_id, task_family FROM tasks WHERE task_id = $1 AND tenant_id = $2"
	var taskID, taskFamily string
	if err := s.postgres.DB.QueryRowContext(r.Context(), taskQ, req.TaskID, actor.TenantID).Scan(&taskID, &taskFamily); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "task_not_found", "Task not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "task_lookup_failed", "Failed to validate task.", corrID, nil)
		return
	}

	quote, err := pricingengine.FixedPlusBid{}.Quote(pricingengine.QuoteRequest{
		TaskID:         req.TaskID,
		BaseRate:       req.BaseRate,
		BidAdjustment:  req.BidAdjustment,
		PredictedRawJW: 1.0,
		QualityFactor:  req.QualityFactor,
		ReliabilityRF:  req.ReliabilityRF,
		ReserveRatio:   req.ReserveRatio,
	})
	if err != nil {
		apierrors.Write(w, http.StatusBadRequest, "rate_resolution_failed", "Failed to resolve rate snapshot.", corrID, map[string]any{"reason": err.Error()})
		return
	}

	rateSnapshotID := strings.TrimSpace(req.RateSnapshotID)
	if rateSnapshotID == "" {
		rateSnapshotID = "rate_" + randomID(6)
	}

	metadataJSON, _ := json.Marshal(map[string]any{
		"task_family":       taskFamily,
		"base_rate":         req.BaseRate,
		"bid_adjustment":    req.BidAdjustment,
		"quality_factor":    req.QualityFactor,
		"reliability_rf":    req.ReliabilityRF,
		"estimated_gross":   quote.EstimatedGross,
		"estimated_reserve": quote.EstimatedReserve,
		"estimated_net":     quote.EstimatedNet,
		"reward_multiplier": quote.RewardMultiplier,
	})

	const insertQ = "INSERT INTO rate_snapshots (rate_snapshot_id, tenant_id, task_id, result_id, strategy_name, rate_value, metadata_json) VALUES ($1,$2,$3,$4,$5,$6,$7)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), insertQ, rateSnapshotID, actor.TenantID, req.TaskID, strings.TrimSpace(req.ResultID), quote.StrategyName, quote.RateValue, metadataJSON); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "rate_resolution_failed", "Failed to persist rate snapshot.", corrID, nil)
		return
	}

	s.appendEvent(r, actor.TenantID, "pricing", rateSnapshotID, "pricing.rate_resolved", map[string]any{
		"task_id":       req.TaskID,
		"task_family":   taskFamily,
		"strategy_name": quote.StrategyName,
		"rate_value":    quote.RateValue,
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"rate_snapshot_id":  rateSnapshotID,
		"task_id":           req.TaskID,
		"result_id":         strings.TrimSpace(req.ResultID),
		"strategy_name":     quote.StrategyName,
		"rate_value":        quote.RateValue,
		"estimated_gross":   quote.EstimatedGross,
		"estimated_reserve": quote.EstimatedReserve,
		"estimated_net":     quote.EstimatedNet,
		"correlation_id":    corrID,
	})
}

func (s *Server) handleReliabilitySubjectV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	subjectType := strings.TrimSpace(r.PathValue("subject_type"))
	subjectID := strings.TrimSpace(r.PathValue("subject_id"))
	s.handleReliabilityLookup(w, r, actor.TenantID, subjectType, subjectID, corrID)
}

func (s *Server) handleOperatorReliabilityV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	s.handleReliabilityLookup(w, r, actor.TenantID, "operator", strings.TrimSpace(r.PathValue("operator_id")), corrID)
}

func (s *Server) handleReliabilityLookup(w http.ResponseWriter, r *http.Request, tenantID, subjectType, subjectID, corrID string) {
	const q = "SELECT snapshot_id, tenant_id, subject_type, subject_id, family, window_name, rf_value, components_json, created_at FROM reliability_snapshots WHERE tenant_id = $1 AND subject_type = $2 AND subject_id = $3 ORDER BY created_at DESC LIMIT 1"
	var snapshotID, dbTenantID, dbSubjectType, dbSubjectID, family, windowName string
	var rfValue float64
	var componentsJSON []byte
	var createdAt time.Time
	if err := s.postgres.DB.QueryRowContext(r.Context(), q, tenantID, subjectType, subjectID).Scan(&snapshotID, &dbTenantID, &dbSubjectType, &dbSubjectID, &family, &windowName, &rfValue, &componentsJSON, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "reliability_not_found", "Reliability snapshot not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "reliability_lookup_failed", "Failed to fetch reliability snapshot.", corrID, nil)
		return
	}

	components := map[string]any{}
	_ = json.Unmarshal(componentsJSON, &components)
	writeJSON(w, http.StatusOK, map[string]any{
		"snapshot_id":    snapshotID,
		"tenant_id":      dbTenantID,
		"subject_type":   dbSubjectType,
		"subject_id":     dbSubjectID,
		"family":         family,
		"window_name":    windowName,
		"rf_value":       rfValue,
		"components":     components,
		"created_at":     createdAt,
		"correlation_id": corrID,
	})
}

func (s *Server) handleGetLedgerAccountV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	const q = "SELECT account_id, tenant_id, owner_type, owner_id, currency, status, created_at FROM ledger_accounts WHERE account_id = $1"
	var accountID, tenantID, ownerType, ownerID, currency, status string
	var createdAt time.Time
	if err := s.postgres.DB.QueryRowContext(r.Context(), q, r.PathValue("account_id")).Scan(&accountID, &tenantID, &ownerType, &ownerID, &currency, &status, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "ledger_account_not_found", "Ledger account not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "ledger_lookup_failed", "Failed to fetch ledger account.", corrID, nil)
		return
	}
	if !s.ensureTenantAccess(actor, tenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant read is not allowed.", corrID, nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"account_id":     accountID,
		"tenant_id":      tenantID,
		"owner_type":     ownerType,
		"owner_id":       ownerID,
		"currency":       currency,
		"status":         status,
		"created_at":     createdAt,
		"correlation_id": corrID,
	})
}

func (s *Server) handleGetLedgerAccountEntriesV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	accountID := strings.TrimSpace(r.PathValue("account_id"))
	const q = "SELECT entry_id, event_type, result_id, raw_jw, credited_jw, rate, gross_amount, reserve_amount, net_amount, status, created_at FROM ledger_entries WHERE tenant_id = $1 AND account_id = $2 ORDER BY created_at DESC"
	rows, err := s.postgres.DB.QueryContext(r.Context(), q, actor.TenantID, accountID)
	if err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "ledger_lookup_failed", "Failed to fetch ledger entries.", corrID, nil)
		return
	}
	defer rows.Close()

	entries := make([]map[string]any, 0)
	for rows.Next() {
		var entryID, eventType, resultID, status string
		var rawJW, creditedJW, rateValue, grossAmount, reserveAmount, netAmount sql.NullFloat64
		var createdAt time.Time
		if err := rows.Scan(&entryID, &eventType, &resultID, &rawJW, &creditedJW, &rateValue, &grossAmount, &reserveAmount, &netAmount, &status, &createdAt); err != nil {
			apierrors.Write(w, http.StatusInternalServerError, "ledger_lookup_failed", "Failed to scan ledger entries.", corrID, nil)
			return
		}
		entries = append(entries, map[string]any{
			"entry_id":       entryID,
			"event_type":     eventType,
			"result_id":      resultID,
			"raw_jw":         rawJW.Float64,
			"credited_jw":    creditedJW.Float64,
			"rate":           rateValue.Float64,
			"gross_amount":   grossAmount.Float64,
			"reserve_amount": reserveAmount.Float64,
			"net_amount":     netAmount.Float64,
			"status":         status,
			"created_at":     createdAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"entries": entries, "correlation_id": corrID})
}

func (s *Server) handleOperatorEarningsSummaryV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	operatorID := strings.TrimSpace(r.PathValue("operator_id"))

	const accountQ = "SELECT account_id FROM ledger_accounts WHERE tenant_id = $1 AND owner_type = $2 AND owner_id = $3"
	var accountID string
	if err := s.postgres.DB.QueryRowContext(r.Context(), accountQ, actor.TenantID, "operator", operatorID).Scan(&accountID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "ledger_account_not_found", "Operator ledger account not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "ledger_lookup_failed", "Failed to resolve operator account.", corrID, nil)
		return
	}

	const sumQ = "SELECT COALESCE(SUM(credited_jw),0), COALESCE(SUM(net_amount),0), COALESCE(SUM(reserve_amount),0) FROM ledger_entries WHERE tenant_id = $1 AND account_id = $2"
	var creditedTotal, netTotal, reserveTotal float64
	if err := s.postgres.DB.QueryRowContext(r.Context(), sumQ, actor.TenantID, accountID).Scan(&creditedTotal, &netTotal, &reserveTotal); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "ledger_lookup_failed", "Failed to compute earnings summary.", corrID, nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"operator_id":        operatorID,
		"account_id":         accountID,
		"credited_jw_total":  creditedTotal,
		"net_amount_total":   netTotal,
		"reserve_held_total": reserveTotal,
		"correlation_id":     corrID,
	})
}

func (s *Server) handleOperatorReserveHoldsV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	operatorID := strings.TrimSpace(r.PathValue("operator_id"))
	const q = "SELECT reserve_hold_id, ledger_entry_id, status, release_after, released_at FROM reserve_holds WHERE tenant_id = $1 AND ledger_entry_id IN (SELECT entry_id FROM ledger_entries WHERE account_id = (SELECT account_id FROM ledger_accounts WHERE tenant_id = $1 AND owner_type = $2 AND owner_id = $3)) ORDER BY release_after DESC"
	rows, err := s.postgres.DB.QueryContext(r.Context(), q, actor.TenantID, "operator", operatorID)
	if err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "ledger_lookup_failed", "Failed to fetch reserve holds.", corrID, nil)
		return
	}
	defer rows.Close()

	holds := make([]map[string]any, 0)
	for rows.Next() {
		var reserveHoldID, ledgerEntryID, status string
		var releaseAfter time.Time
		var releasedAt sql.NullTime
		if err := rows.Scan(&reserveHoldID, &ledgerEntryID, &status, &releaseAfter, &releasedAt); err != nil {
			apierrors.Write(w, http.StatusInternalServerError, "ledger_lookup_failed", "Failed to scan reserve holds.", corrID, nil)
			return
		}
		holds = append(holds, map[string]any{
			"reserve_hold_id": reserveHoldID,
			"ledger_entry_id": ledgerEntryID,
			"status":          status,
			"release_after":   releaseAfter,
			"released_at":     releasedAt.Time,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"reserve_holds": holds, "correlation_id": corrID})
}

type settlementRequest struct {
	EntryID             string                `json:"entry_id,omitempty"`
	ReserveHoldID       string                `json:"reserve_hold_id,omitempty"`
	ResultID            string                `json:"result_id"`
	Rate                float64               `json:"rate"`
	TaskDifficulty      string                `json:"task_difficulty"`
	ReserveRatio        float64               `json:"reserve_ratio"`
	NodeSnapshotID      string                `json:"node_snapshot_id,omitempty"`
	OpSnapshotID        string                `json:"op_snapshot_id,omitempty"`
	Metering            scoring.Metering      `json:"metering"`
	QualityInputs       scoring.QualityInputs `json:"quality_inputs"`
	RFLast100           float64               `json:"rf_last_100"`
	RFLast30d           float64               `json:"rf_last_30d"`
	RFLifetime          float64               `json:"rf_lifetime"`
	ReleaseAfterSeconds int                   `json:"release_after_seconds"`
}

func (s *Server) handlePostSettlementV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "finance_viewer", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for settlement posting.", corrID, nil)
		return
	}

	var req settlementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	req.ResultID = strings.TrimSpace(req.ResultID)
	if req.ResultID == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "result_id is required.", corrID, nil)
		return
	}

	const resultQ = "SELECT result_id, tenant_id, task_id, node_id FROM results WHERE result_id = $1"
	var resultID, tenantID, taskID, nodeID string
	if err := s.postgres.DB.QueryRowContext(r.Context(), resultQ, req.ResultID).Scan(&resultID, &tenantID, &taskID, &nodeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "result_not_found", "Result not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "settlement_failed", "Failed to load result for settlement.", corrID, nil)
		return
	}
	if !s.ensureTenantAccess(actor, tenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant settlement is not allowed.", corrID, nil)
		return
	}

	const nodeQ = "SELECT operator_id FROM nodes WHERE node_id = $1 AND tenant_id = $2"
	var operatorID string
	if err := s.postgres.DB.QueryRowContext(r.Context(), nodeQ, nodeID, tenantID).Scan(&operatorID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "node_not_found", "Result node not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "settlement_failed", "Failed to resolve operator for result.", corrID, nil)
		return
	}

	accountID, err := s.getOrCreateOperatorAccount(r, tenantID, operatorID)
	if err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "settlement_failed", "Failed to resolve operator account.", corrID, nil)
		return
	}

	engine := settlement.DefaultEngine{}
	payout, err := engine.Settle(settlement.ScoreInput{
		Metering:       req.Metering,
		TaskDifficulty: req.TaskDifficulty,
		Quality:        req.QualityInputs,
		RFLast100:      req.RFLast100,
		RFLast30d:      req.RFLast30d,
		RFLifetime:     req.RFLifetime,
		Rate:           req.Rate,
		ReserveRatio:   req.ReserveRatio,
	})
	if err != nil {
		apierrors.Write(w, http.StatusBadRequest, "settlement_failed", "Settlement computation failed.", corrID, map[string]any{"reason": err.Error()})
		return
	}

	entryID := strings.TrimSpace(req.EntryID)
	if entryID == "" {
		entryID = "entry_" + randomID(6)
	}
	metadataJSON, _ := json.Marshal(map[string]any{
		"task_id":       taskID,
		"node_id":       nodeID,
		"operator_id":   operatorID,
		"rf_last_100":   req.RFLast100,
		"rf_last_30d":   req.RFLast30d,
		"rf_lifetime":   req.RFLifetime,
		"quality_input": req.QualityInputs,
	})
	const insertEntryQ = "INSERT INTO ledger_entries (entry_id, tenant_id, account_id, event_type, result_id, raw_jw, credited_jw, rate, gross_amount, reserve_amount, net_amount, status, metadata_json) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), insertEntryQ, entryID, tenantID, accountID, "ledger.settlement_posted", req.ResultID, payout.RawJW, payout.CreditedJW, payout.Rate, payout.GrossAmount, payout.ReserveAmount, payout.NetAmount, "posted", metadataJSON); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "settlement_failed", "Failed to persist ledger entry.", corrID, nil)
		return
	}

	releaseAfterSeconds := req.ReleaseAfterSeconds
	if releaseAfterSeconds <= 0 {
		releaseAfterSeconds = 24 * 3600
	}
	releaseAfter := time.Now().UTC().Add(time.Duration(releaseAfterSeconds) * time.Second)
	reserveHoldID := strings.TrimSpace(req.ReserveHoldID)
	if reserveHoldID == "" {
		reserveHoldID = "hold_" + randomID(6)
	}
	const insertHoldQ = "INSERT INTO reserve_holds (reserve_hold_id, tenant_id, ledger_entry_id, status, release_after) VALUES ($1,$2,$3,$4,$5)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), insertHoldQ, reserveHoldID, tenantID, entryID, "held", releaseAfter); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "settlement_failed", "Failed to persist reserve hold.", corrID, nil)
		return
	}

	nodeSnapshotID := strings.TrimSpace(req.NodeSnapshotID)
	if nodeSnapshotID == "" {
		nodeSnapshotID = "rel_" + randomID(6)
	}
	opSnapshotID := strings.TrimSpace(req.OpSnapshotID)
	if opSnapshotID == "" {
		opSnapshotID = "rel_" + randomID(6)
	}
	nodeComponentsJSON, _ := json.Marshal(map[string]any{"quality_factor": payout.QualityFactor, "reward_multiplier": payout.RewardMultiplier})
	opComponentsJSON, _ := json.Marshal(map[string]any{"settled_result_id": req.ResultID, "credited_jw": payout.CreditedJW})

	const insertReliabilityQ = "INSERT INTO reliability_snapshots (snapshot_id, tenant_id, subject_type, subject_id, family, window_name, rf_value, components_json) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), insertReliabilityQ, nodeSnapshotID, tenantID, "node", nodeID, "task", "last_30d", payout.RFFinal, nodeComponentsJSON); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "settlement_failed", "Failed to persist node reliability snapshot.", corrID, nil)
		return
	}
	operatorRF := payout.RFFinal + 0.02
	if operatorRF > 1 {
		operatorRF = 1
	}
	if _, err := s.postgres.DB.ExecContext(r.Context(), insertReliabilityQ, opSnapshotID, tenantID, "operator", operatorID, "fleet", "last_30d", operatorRF, opComponentsJSON); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "settlement_failed", "Failed to persist operator reliability snapshot.", corrID, nil)
		return
	}

	s.appendEvent(r, tenantID, "ledger", entryID, "ledger.settlement_posted", map[string]any{"result_id": req.ResultID, "account_id": accountID})
	s.appendEvent(r, tenantID, "ledger", reserveHoldID, "ledger.reserve_hold_created", map[string]any{"entry_id": entryID, "release_after": releaseAfter})
	s.appendEvent(r, tenantID, "reliability", nodeSnapshotID, "reliability.snapshot_updated", map[string]any{"subject_type": "node", "subject_id": nodeID})
	s.appendEvent(r, tenantID, "reliability", opSnapshotID, "reliability.snapshot_updated", map[string]any{"subject_type": "operator", "subject_id": operatorID})

	writeJSON(w, http.StatusOK, map[string]any{
		"entry_id":        entryID,
		"reserve_hold_id": reserveHoldID,
		"account_id":      accountID,
		"operator_id":     operatorID,
		"payout":          payout,
		"release_after":   releaseAfter,
		"correlation_id":  corrID,
	})
}

func (s *Server) getOrCreateOperatorAccount(r *http.Request, tenantID, operatorID string) (string, error) {
	const lookupQ = "SELECT account_id FROM ledger_accounts WHERE tenant_id = $1 AND owner_type = $2 AND owner_id = $3"
	var accountID string
	if err := s.postgres.DB.QueryRowContext(r.Context(), lookupQ, tenantID, "operator", operatorID).Scan(&accountID); err == nil {
		return accountID, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	accountID = "acct_" + randomID(6)
	const insertQ = "INSERT INTO ledger_accounts (account_id, tenant_id, owner_type, owner_id, currency, status) VALUES ($1,$2,$3,$4,$5,$6)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), insertQ, accountID, tenantID, "operator", operatorID, "JWUSD", "active"); err != nil {
		return "", err
	}
	return accountID, nil
}

func (s *Server) handleReleaseReserveHoldV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "finance_viewer", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for reserve release.", corrID, nil)
		return
	}

	reserveHoldID := strings.TrimSpace(r.PathValue("reserve_hold_id"))
	if reserveHoldID == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "reserve_hold_id is required.", corrID, nil)
		return
	}
	releasedAt := time.Now().UTC()
	const q = "UPDATE reserve_holds SET status = $1, released_at = $2 WHERE reserve_hold_id = $3 AND tenant_id = $4"
	result, err := s.postgres.DB.ExecContext(r.Context(), q, "released", releasedAt, reserveHoldID, actor.TenantID)
	if err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "settlement_failed", "Failed to release reserve hold.", corrID, nil)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		apierrors.Write(w, http.StatusNotFound, "reserve_hold_not_found", "Reserve hold not found.", corrID, nil)
		return
	}

	s.appendEvent(r, actor.TenantID, "ledger", reserveHoldID, "ledger.reserve_hold_released", map[string]any{"released_at": releasedAt})
	writeJSON(w, http.StatusOK, map[string]any{
		"reserve_hold_id": reserveHoldID,
		"status":          "released",
		"released_at":     releasedAt,
		"correlation_id":  corrID,
	})
}
