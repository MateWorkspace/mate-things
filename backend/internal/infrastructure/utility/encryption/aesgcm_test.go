package infrastructureutilityencryption

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := "01234567890123456789012345678901" // 32 bytes for AES-256
	encryptor, err := NewAESGCMImpl(key)
	if err != nil {
		t.Fatalf("NewAESGCMImpl() error = %v, want nil", err)
	}

	plaintext := "sk-ant-super-secret-key"
	ciphertext, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v, want nil", err)
	}
	if bytes.Contains(ciphertext, []byte(plaintext)) {
		t.Fatalf("Encrypt() output contains the plaintext verbatim")
	}

	decrypted, err := encryptor.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v, want nil", err)
	}
	if decrypted != plaintext {
		t.Fatalf("Decrypt() = %q, want %q", decrypted, plaintext)
	}
}

func TestEncryptProducesDistinctCiphertextsForSamePlaintext(t *testing.T) {
	key := "01234567890123456789012345678901"
	encryptor, err := NewAESGCMImpl(key)
	if err != nil {
		t.Fatalf("NewAESGCMImpl() error = %v, want nil", err)
	}

	first, err := encryptor.Encrypt("same-plaintext")
	if err != nil {
		t.Fatalf("Encrypt() error = %v, want nil", err)
	}
	second, err := encryptor.Encrypt("same-plaintext")
	if err != nil {
		t.Fatalf("Encrypt() error = %v, want nil", err)
	}
	if bytes.Equal(first, second) {
		t.Fatalf("Encrypt() produced identical ciphertexts for the same plaintext across two calls (nonce reuse)")
	}
}

func TestDecryptRejectsCorruptedCiphertext(t *testing.T) {
	key := "01234567890123456789012345678901"
	encryptor, err := NewAESGCMImpl(key)
	if err != nil {
		t.Fatalf("NewAESGCMImpl() error = %v, want nil", err)
	}

	ciphertext, err := encryptor.Encrypt("plaintext")
	if err != nil {
		t.Fatalf("Encrypt() error = %v, want nil", err)
	}
	ciphertext[len(ciphertext)-1] ^= 0xFF // flip the last byte

	if _, err := encryptor.Decrypt(ciphertext); err == nil {
		t.Fatalf("Decrypt() error = nil, want an authentication error on corrupted ciphertext")
	}
}
