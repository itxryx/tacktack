package repository

import (
	"testing"

	"github.com/itxryx/tacktack/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagRepo_FindOrCreate_NoDuplicate(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	repo := NewTagRepository(database)

	tag1, err := repo.FindOrCreate("work", "project")
	require.NoError(t, err)

	// act
	tag2, err := repo.FindOrCreate("work", "project")
	require.NoError(t, err)

	// assert
	assert.Equal(t, tag1.ID, tag2.ID, "同名タグは1件しか作成されないこと")

	all, err := repo.FindAll()
	require.NoError(t, err)
	assert.Len(t, all, 1)
}

func TestTagRepo_FindByType(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	repo := NewTagRepository(database)

	_, err := repo.FindOrCreate("work", "project")
	require.NoError(t, err)
	_, err = repo.FindOrCreate("office", "context")
	require.NoError(t, err)

	// act
	projects, err := repo.FindByType("project")
	require.NoError(t, err)

	contexts, err := repo.FindByType("context")
	require.NoError(t, err)

	// assert
	assert.Len(t, projects, 1)
	assert.Equal(t, "work", projects[0].Name)
	assert.Len(t, contexts, 1)
	assert.Equal(t, "office", contexts[0].Name)
}

func TestTagRepo_FindAll_Empty(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	repo := NewTagRepository(database)

	// act
	tags, err := repo.FindAll()
	require.NoError(t, err)

	// assert
	assert.Empty(t, tags)
}

func TestTagRepo_Delete(t *testing.T) {
	t.Run("通常削除", func(t *testing.T) {
		// arrange
		database := setupTestDB(t)
		repo := NewTagRepository(database)

		tag, err := repo.FindOrCreate("work", "project")
		require.NoError(t, err)

		// act
		require.NoError(t, repo.Delete(tag.ID))

		// assert
		all, err := repo.FindAll()
		require.NoError(t, err)
		assert.Empty(t, all, "FindAll から消えていること")
	})

	t.Run("SoftDelete確認", func(t *testing.T) {
		// arrange
		database := setupTestDB(t)
		repo := NewTagRepository(database)

		tag, err := repo.FindOrCreate("work", "project")
		require.NoError(t, err)
		require.NoError(t, repo.Delete(tag.ID))

		// act + assert
		all, err := repo.FindAll()
		require.NoError(t, err)
		assert.Empty(t, all, "通常クエリでは見えないこと")

		var got db.Tag
		require.NoError(t, database.Unscoped().First(&got, tag.ID).Error)
		assert.True(t, got.DeletedAt.Valid, "deleted_at が設定されていること")
	})
}

func TestTagRepo_FindOrCreate_RestoreSoftDeleted(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	repo := NewTagRepository(database)

	tag, err := repo.FindOrCreate("work", "project")
	require.NoError(t, err)
	originalID := tag.ID
	require.NoError(t, repo.Delete(tag.ID))

	// act
	restored, err := repo.FindOrCreate("work", "project")
	require.NoError(t, err)

	// assert
	assert.Equal(t, originalID, restored.ID, "同じIDのレコードが復活すること")
	assert.False(t, restored.DeletedAt.Valid, "deleted_at が NULL に戻っていること")

	all, err := repo.FindAll()
	require.NoError(t, err)
	assert.Len(t, all, 1)
}

func TestTagRepo_FindOrCreate_RestoreSoftDeleted_TypeOverwrite(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	repo := NewTagRepository(database)

	tag, err := repo.FindOrCreate("work", "project")
	require.NoError(t, err)
	require.NoError(t, repo.Delete(tag.ID))

	// act
	restored, err := repo.FindOrCreate("work", "context")
	require.NoError(t, err)

	// assert
	assert.Equal(t, tag.ID, restored.ID, "同じIDのレコードが復活すること")
	assert.Equal(t, "context", restored.Type, "Type が context に上書きされること")
	assert.False(t, restored.DeletedAt.Valid, "deleted_at が NULL に戻っていること")
}

func TestTagRepo_Delete_ClearsTaskTags(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	tagRepo := NewTagRepository(database)
	taskRepo := NewTaskRepository(database)

	task := &db.Task{Title: "タスク"}
	require.NoError(t, taskRepo.Create(task))

	tag, err := tagRepo.FindOrCreate("work", "project")
	require.NoError(t, err)
	require.NoError(t, taskRepo.ReplaceTagsForTask(task.ID, []db.Tag{*tag}))

	// act
	require.NoError(t, tagRepo.Delete(tag.ID))

	// assert
	var count int64
	database.Table("task_tags").Where("tag_id = ?", tag.ID).Count(&count)
	assert.Zero(t, count)

	all, err := tagRepo.FindAll()
	require.NoError(t, err)
	assert.Empty(t, all)
}
