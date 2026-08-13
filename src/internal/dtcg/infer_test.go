package dtcg

import "testing"

func TestInferredTypeUsesSemanticIDForUntypedTokens(t *testing.T) {
	if got := inferredType("", "color.action.primary"); got != "color" {
		t.Fatalf("got %q", got)
	}
}
