package acl

import (
	"database/sql"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsInvalidUUID(t *testing.T) {
	t.Parallel()

	t.Run("true for uuid invalid syntax", func(t *testing.T) {
		err := &pgconn.PgError{
			Code:    "22P02",
			Message: `invalid input syntax for type uuid: "not-a-uuid"`,
		}
		if !isInvalidUUID(err) {
			t.Fatal("isInvalidUUID() = false, want true")
		}
	})

	t.Run("false for non-uuid 22P02", func(t *testing.T) {
		err := &pgconn.PgError{
			Code:    "22P02",
			Message: `invalid input syntax for type integer: "x"`,
		}
		if isInvalidUUID(err) {
			t.Fatal("isInvalidUUID() = true, want false")
		}
	})
}

func TestIsNotFoundLookup(t *testing.T) {
	t.Parallel()

	if !isNotFoundLookup(sql.ErrNoRows) {
		t.Fatal("isNotFoundLookup(sql.ErrNoRows) = false, want true")
	}

	err := &pgconn.PgError{
		Code:    "22P02",
		Message: `invalid input syntax for type uuid: "not-a-uuid"`,
	}
	if !isNotFoundLookup(err) {
		t.Fatal("isNotFoundLookup(invalid uuid) = false, want true")
	}
}
