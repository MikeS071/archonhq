package policy

import (
	"fmt"
	"strings"
)

var supportedBackends = map[string]struct{}{
	"docker": {},
	"ssh":    {},
	"modal":  {},
}

var supportedNetworkPolicies = map[string]struct{}{
	"restricted":       {},
	"deny_all":         {},
	"tenant_allowlist": {},
}

type RuntimeExecutionPolicy struct {
	AllowedBackends       []string `json:"allowed_backends"`
	AllowedToolsets       []string `json:"allowed_toolsets"`
	NetworkPolicy         string   `json:"network_policy"`
	WorkspaceMode         string   `json:"workspace_mode"`
	HermesMemoryWriteback bool     `json:"hermes_memory_writeback"`
	InferenceMode         string   `json:"inference_mode"`
	ByokProviderRef       string   `json:"byok_provider_ref"`
}

func NormalizeExecutionPolicy(raw map[string]any) (RuntimeExecutionPolicy, error) {
	p := RuntimeExecutionPolicy{
		AllowedBackends:       []string{"docker"},
		AllowedToolsets:       []string{"file", "terminal"},
		NetworkPolicy:         "restricted",
		WorkspaceMode:         "ephemeral",
		HermesMemoryWriteback: false,
		InferenceMode:         "byok_only",
		ByokProviderRef:       "tenant_default",
	}
	if raw == nil {
		return p, nil
	}

	if backends, ok := raw["allowed_backends"]; ok {
		normalized, err := asStringSlice(backends)
		if err != nil {
			return RuntimeExecutionPolicy{}, fmt.Errorf("allowed_backends: %w", err)
		}
		if len(normalized) == 0 {
			return RuntimeExecutionPolicy{}, fmt.Errorf("allowed_backends must not be empty")
		}
		uniq := make([]string, 0, len(normalized))
		seen := map[string]struct{}{}
		for _, backend := range normalized {
			backend = strings.ToLower(strings.TrimSpace(backend))
			if _, ok := supportedBackends[backend]; !ok {
				return RuntimeExecutionPolicy{}, fmt.Errorf("unsupported backend: %s", backend)
			}
			if _, ok := seen[backend]; ok {
				continue
			}
			seen[backend] = struct{}{}
			uniq = append(uniq, backend)
		}
		p.AllowedBackends = uniq
	}

	if toolsets, ok := raw["allowed_toolsets"]; ok {
		normalized, err := asStringSlice(toolsets)
		if err != nil {
			return RuntimeExecutionPolicy{}, fmt.Errorf("allowed_toolsets: %w", err)
		}
		if len(normalized) > 0 {
			p.AllowedToolsets = normalized
		}
	}

	if network, ok := raw["network_policy"]; ok {
		networkStr, ok := network.(string)
		if !ok {
			return RuntimeExecutionPolicy{}, fmt.Errorf("network_policy must be a string")
		}
		networkStr = strings.TrimSpace(strings.ToLower(networkStr))
		if _, ok := supportedNetworkPolicies[networkStr]; !ok {
			return RuntimeExecutionPolicy{}, fmt.Errorf("unsupported network_policy: %s", networkStr)
		}
		p.NetworkPolicy = networkStr
	}

	if workspaceMode, ok := raw["workspace_mode"]; ok {
		modeStr, ok := workspaceMode.(string)
		if !ok {
			return RuntimeExecutionPolicy{}, fmt.Errorf("workspace_mode must be a string")
		}
		modeStr = strings.TrimSpace(strings.ToLower(modeStr))
		if modeStr != "" && modeStr != "ephemeral" {
			return RuntimeExecutionPolicy{}, fmt.Errorf("workspace_mode must be ephemeral")
		}
	}

	if writeback, ok := raw["hermes_memory_writeback"]; ok {
		writebackBool, ok := writeback.(bool)
		if !ok {
			return RuntimeExecutionPolicy{}, fmt.Errorf("hermes_memory_writeback must be a boolean")
		}
		if writebackBool {
			return RuntimeExecutionPolicy{}, fmt.Errorf("hermes_memory_writeback must be false")
		}
	}

	if inferenceMode, ok := raw["inference_mode"]; ok {
		modeStr, ok := inferenceMode.(string)
		if !ok {
			return RuntimeExecutionPolicy{}, fmt.Errorf("inference_mode must be a string")
		}
		modeStr = strings.TrimSpace(strings.ToLower(modeStr))
		if modeStr != "" && modeStr != "byok_only" {
			return RuntimeExecutionPolicy{}, fmt.Errorf("inference_mode must be byok_only")
		}
	}

	if providerRef, ok := raw["byok_provider_ref"]; ok {
		refStr, ok := providerRef.(string)
		if !ok {
			return RuntimeExecutionPolicy{}, fmt.Errorf("byok_provider_ref must be a string")
		}
		refStr = strings.TrimSpace(refStr)
		if refStr != "" {
			p.ByokProviderRef = refStr
		}
	}

	return p, nil
}

func (p RuntimeExecutionPolicy) Snapshot() map[string]any {
	return map[string]any{
		"allowed_backends":        p.AllowedBackends,
		"allowed_toolsets":        p.AllowedToolsets,
		"network_policy":          p.NetworkPolicy,
		"workspace_mode":          p.WorkspaceMode,
		"hermes_memory_writeback": p.HermesMemoryWriteback,
		"inference_mode":          p.InferenceMode,
		"byok_provider_ref":       p.ByokProviderRef,
	}
}

func asStringSlice(v any) ([]string, error) {
	switch typed := v.(type) {
	case []string:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			item = strings.TrimSpace(item)
			if item != "" {
				out = append(out, item)
			}
		}
		return out, nil
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			str, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("must contain only strings")
			}
			str = strings.TrimSpace(str)
			if str != "" {
				out = append(out, str)
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("must be an array of strings")
	}
}
