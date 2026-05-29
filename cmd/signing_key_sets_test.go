package cmd

import (
	"testing"
)

func TestParseSignJWTInput_ValidPayload(t *testing.T) {
	in, err := parseSignJWTInput(`{"sub":"alice","aud":"api://prod"}`, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.Payload["sub"] != "alice" {
		t.Fatalf("Payload[sub] = %v, want alice", in.Payload["sub"])
	}
	if in.Payload["aud"] != "api://prod" {
		t.Fatalf("Payload[aud] = %v, want api://prod", in.Payload["aud"])
	}
	if in.KID != "" {
		t.Fatalf("KID = %q, want empty", in.KID)
	}
	if in.Header != nil {
		t.Fatalf("Header = %v, want nil", in.Header)
	}
}

func TestParseSignJWTInput_WithKIDAndHeader(t *testing.T) {
	in, err := parseSignJWTInput(`{"sub":"bob"}`, `{"typ":"JWT"}`, "key-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.KID != "key-123" {
		t.Fatalf("KID = %q, want key-123", in.KID)
	}
	if in.Header["typ"] != "JWT" {
		t.Fatalf("Header[typ] = %v, want JWT", in.Header["typ"])
	}
}

func TestParseSignJWTInput_InvalidPayloadJSON(t *testing.T) {
	if _, err := parseSignJWTInput(`{not valid json}`, "", ""); err == nil {
		t.Fatal("expected error for invalid payload JSON, got nil")
	}
}

func TestParseSignJWTInput_InvalidHeaderJSON(t *testing.T) {
	if _, err := parseSignJWTInput(`{"sub":"alice"}`, `{bad}`, ""); err == nil {
		t.Fatal("expected error for invalid header JSON, got nil")
	}
}

func TestParseSignJWTInput_EmptyPayloadParsesNilMap(t *testing.T) {
	in, err := parseSignJWTInput("", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.Payload != nil {
		t.Fatalf("Payload = %v, want nil for empty string", in.Payload)
	}
}
