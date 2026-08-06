package domaincontractsutility

type ApiKey interface {
	Generate() (raw string, hash string, lastFour string, err error)
	Hash(raw string) (hash string)
}
