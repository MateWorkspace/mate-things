package domaincontractsutility

type Encryptor interface {
	Encrypt(plaintext string) (ciphertext []byte, err error)
	Decrypt(ciphertext []byte) (plaintext string, err error)
}
