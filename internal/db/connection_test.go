package db

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitDB_WALMode(t *testing.T) {
	// arrange
	tmp := filepath.Join(t.TempDir(), "test.db")

	// act
	db, err := InitDB(tmp)
	require.NoError(t, err)

	// assert
	var journalMode string
	require.NoError(t, db.Raw("PRAGMA journal_mode").Scan(&journalMode).Error)
	assert.Equal(t, "wal", journalMode)
}

func TestInitDB_InvalidPath(t *testing.T) {
	// act
	_, err := InitDB("/nonexistent_root_dir_abc123/tacktack/test.db")

	// assert
	assert.Error(t, err)
}
