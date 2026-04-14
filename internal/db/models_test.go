package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := InitDB(":memory:")
	require.NoError(t, err)
	require.NoError(t, Migrate(db))
	return db
}

func TestMigrate_TablesExist(t *testing.T) {
	// arrange
	db := setupTestDB(t)

	// act
	m := db.Migrator()

	// assert
	assert.True(t, m.HasTable(&Task{}))
	assert.True(t, m.HasTable(&Tag{}))
	assert.True(t, m.HasTable(&TimeLog{}))
	assert.True(t, m.HasTable("task_tags"))
}

func TestTag_UniqueNameConstraint(t *testing.T) {
	// arrange
	db := setupTestDB(t)
	require.NoError(t, db.Create(&Tag{Name: "work", Type: "project"}).Error)

	// act
	err := db.Create(&Tag{Name: "work", Type: "context"}).Error

	// assert
	assert.Error(t, err, "重複タグ名はエラーになること")
}
