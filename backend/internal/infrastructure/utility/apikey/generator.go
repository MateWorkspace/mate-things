package infrastructureutilityapikey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

const (
	keyPrefix    = "mate_"
	randomBytes  = 32
	lastFourSize = 4
)

type generatorImpl struct{}

func NewGeneratorImpl() domaincontractsutility.ApiKey {
	return &generatorImpl{}
}

func (g *generatorImpl) Generate() (raw string, hash string, lastFour string, err error) {
	buf := make([]byte, randomBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", "", "", domainmodels.NewError("failed to generate api key", domainmodels.ErrTypeFailure, err)
	}

	raw = keyPrefix + hex.EncodeToString(buf)
	hash = g.Hash(raw)
	lastFour = raw[len(raw)-lastFourSize:]

	return raw, hash, lastFour, nil
}

func (g *generatorImpl) Hash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
