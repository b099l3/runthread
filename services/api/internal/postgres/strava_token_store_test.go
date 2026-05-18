package postgres

import (
	"encoding/base64"
	"testing"
)

func TestProviderTokenKeyBytesAcceptsBase64Encoded32Bytes(t *testing.T) {
	key := base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))

	decoded, err := providerTokenKeyBytes(key)
	if err != nil {
		t.Fatalf("providerTokenKeyBytes returned error: %v", err)
	}
	if string(decoded) != "12345678901234567890123456789012" {
		t.Fatalf("decoded key = %q, want raw bytes", string(decoded))
	}
}

func TestProviderTokenKeyBytesRejectsShortKey(t *testing.T) {
	_, err := providerTokenKeyBytes("too-short")
	if err == nil {
		t.Fatal("expected short key error")
	}
}
