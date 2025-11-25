package postgres

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestIdentityRepository_WithSQLMockPlaceholder(t *testing.T) {
	_, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Skip("placeholder to keep sqlmock dependency until real tests are added")
}
