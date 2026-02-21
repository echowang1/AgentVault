package tss

import "testing"

func TestKeyGenAndSignIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration test in short mode")
	}

	t.Skip("waiting for Task 003 signing module")
}
