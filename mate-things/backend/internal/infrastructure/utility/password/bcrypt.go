package infrastructureutilitypassword

import (
	"errors"

	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	"golang.org/x/crypto/bcrypt"
)

type bcryptImpl struct {
	cost int
}

func NewBcryptImpl(cost int) domaincontractsutility.Password {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}

	return &bcryptImpl{
		cost: cost,
	}
}

func (b *bcryptImpl) Hash(password string) (hashed string, err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", domainmodels.NewError("failed to hash password", domainmodels.ErrTypeFailure, err)
	}

	return string(hash), nil
}

func (b *bcryptImpl) Compare(storedHash string, password string) (err error) {
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return domainmodels.NewError("password does not match", domainmodels.ErrTypeUnauthorized, err)
		}

		return domainmodels.NewError("failed to compare password", domainmodels.ErrTypeFailure, err)
	}

	return nil
}
