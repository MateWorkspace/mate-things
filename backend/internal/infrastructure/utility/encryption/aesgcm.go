package infrastructureutilityencryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
)

type aesGCMImpl struct {
	gcm cipher.AEAD
}

// NewAESGCMImpl builds an AES-256-GCM encryptor. key must be exactly 32 bytes
// (AES-256); pass it via the BE_LLM_ENCRYPTION_KEY environment variable.
func NewAESGCMImpl(key string) (domaincontractsutility.Encryptor, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("failed to construct AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to construct GCM mode: %w", err)
	}
	return &aesGCMImpl{gcm: gcm}, nil
}

func (a *aesGCMImpl) Encrypt(plaintext string) ([]byte, error) {
	nonce := make([]byte, a.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	return a.gcm.Seal(nonce, nonce, []byte(plaintext), nil), nil
}

func (a *aesGCMImpl) Decrypt(ciphertext []byte) (string, error) {
	nonceSize := a.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext shorter than nonce size")
	}
	nonce, sealed := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := a.gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	return string(plaintext), nil
}
