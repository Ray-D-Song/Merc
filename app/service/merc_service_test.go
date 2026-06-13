package service

import (
	"strings"
	"testing"
)

func TestRedactTaskPayload(t *testing.T) {
	payload := `{"registrationToken":"secret-token","name":"runner-1"}`
	got := RedactTaskPayload(payload)
	if strings.Contains(got, "secret-token") {
		t.Fatalf("RedactTaskPayload() leaked token: %s", got)
	}
	if !strings.Contains(got, "[redacted]") {
		t.Fatalf("RedactTaskPayload() did not redact token: %s", got)
	}
}
