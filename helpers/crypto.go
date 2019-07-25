package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
)

// GenerateRandomBytes uses the `crypto/rand` library to create a secure random `[]byte` at a given size `n`.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateSecureString uses the `crypto/rand` and `encoding/base64` libraries to create a random `string`
// of given length `n`.
func GenerateSecureString(n int) (string, error) {
	b, err := GenerateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b), err
}

// EncryptString encrypts a `string` with the `golang.org/x/crypto/bcrypt` library at a given cost.
func EncryptString(str string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(str), cost)
	return string(bytes), err
}

// CompareEncryptedData uses the `golang.org/x/crypto/bcrypt` library to compare a `string` to an
// encrypted `[]byte`. Returns true if the `string` matches the encrypted `[]byte`.
func CompareEncryptedData(str string, hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, []byte(str))
	return err == nil
}
