package infrastructurecacheshared

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func StringPart(value string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(value))
}

func Int32Part(value int32) string {
	return fmt.Sprintf("%d", value)
}

func HashPart(values ...any) (string, error) {
	raw, err := json.Marshal(values)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}
