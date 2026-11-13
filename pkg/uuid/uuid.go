package uuid

import "github.com/google/uuid"

// NewV7 returns a random UUID version 7 string.
func NewV7() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
