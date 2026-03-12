package settlement

import "testing"

func TestDefaultEngine(t *testing.T) {
	e := DefaultEngine{}
	if _, err := e.Settle(ScoreInput{}); err == nil {
		t.Fatalf("expected not implemented error")
	}
}
