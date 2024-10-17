// internal/rps/library_test.go
package rps

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	original := "rock"
	encrypted := Encrypt(original)
	decrypted := Decrypt(encrypted)

	if original != decrypted {
		t.Errorf("Decrypt(Encrypt(%s)) = %s; want %s", original, decrypted, original)
	}
}
