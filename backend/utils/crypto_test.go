package utils

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestEncryptDecryptWithBase64Key(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32))))

	encrypted, err := Encrypt("github-token")
	if err != nil {
		t.Fatalf("Encrypt returned an error: %v", err)
	}
	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt returned an error: %v", err)
	}
	if decrypted != "github-token" {
		t.Fatalf("Decrypt returned %q", decrypted)
	}
}

func TestEncryptDecryptWithHexKey(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", strings.Repeat("ab", 32))

	encrypted, err := Encrypt("provider-token")
	if err != nil {
		t.Fatalf("Encrypt returned an error: %v", err)
	}
	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt returned an error: %v", err)
	}
	if decrypted != "provider-token" {
		t.Fatalf("Decrypt returned %q", decrypted)
	}
}
