package repository

import (
	"testing"

	"github.com/itxryx/tacktack/internal/db"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := db.InitDB(":memory:")
	require.NoError(t, err)
	require.NoError(t, db.Migrate(database))
	return database
}
