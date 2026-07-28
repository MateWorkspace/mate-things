package domaincontractsutility

type Password interface {
	Hash(password string) (hashed string, err error)
	Compare(storedHash string, password string) (err error)
}
