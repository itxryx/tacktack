package repository

import (
	"testing"
	"time"

	"github.com/itxryx/tacktack/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestTaskRepo_Create(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	repo := NewTaskRepository(database)
	task := &db.Task{Title: "テストタスク"}

	// act
	require.NoError(t, repo.Create(task))

	// assert
	assert.NotZero(t, task.ID)
}

func TestTaskRepo_FindAll_DefaultExcludesCompleted(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	repo := NewTaskRepository(database)

	require.NoError(t, repo.Create(&db.Task{Title: "未完了"}))
	completed := &db.Task{Title: "完了済み"}
	require.NoError(t, repo.Create(completed))
	require.NoError(t, repo.ToggleComplete(completed.ID))

	// act
	tasks, err := repo.FindAll()
	require.NoError(t, err)

	// assert
	assert.Len(t, tasks, 1)
	assert.Equal(t, "未完了", tasks[0].Title)
}

func TestTaskRepo_FindAll_SortOrder(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	repo := NewTaskRepository(database)

	due := time.Now().Add(24 * time.Hour)

	noPriority := &db.Task{Title: "優先度なし"}
	require.NoError(t, repo.Create(noPriority))

	priorityB := &db.Task{Title: "優先度B", Priority: "B"}
	require.NoError(t, repo.Create(priorityB))

	priorityAWithDue := &db.Task{Title: "優先度A・締切あり", Priority: "A", DueDate: &due}
	require.NoError(t, repo.Create(priorityAWithDue))

	// act
	tasks, err := repo.FindAll()
	require.NoError(t, err)

	// assert
	require.Len(t, tasks, 3)
	assert.Equal(t, "優先度A・締切あり", tasks[0].Title)
	assert.Equal(t, "優先度B", tasks[1].Title)
	assert.Equal(t, "優先度なし", tasks[2].Title)
}

func TestTaskRepo_ToggleComplete(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	repo := NewTaskRepository(database)

	task := &db.Task{Title: "タスク"}
	require.NoError(t, repo.Create(task))

	// act
	require.NoError(t, repo.ToggleComplete(task.ID))
	got, err := repo.FindByID(task.ID)
	require.NoError(t, err)

	// assert
	assert.True(t, got.IsCompleted)
	assert.NotNil(t, got.CompletedAt)

	// act
	require.NoError(t, repo.ToggleComplete(task.ID))
	got, err = repo.FindByID(task.ID)
	require.NoError(t, err)

	// assert
	assert.False(t, got.IsCompleted)
	assert.Nil(t, got.CompletedAt)
}

func TestTaskRepo_Delete(t *testing.T) {
	t.Run("Cascades", func(t *testing.T) {
		// arrange
		database := setupTestDB(t)
		taskRepo := NewTaskRepository(database)
		logRepo := NewTimeLogRepository(database)

		task := &db.Task{Title: "削除対象"}
		require.NoError(t, taskRepo.Create(task))

		_, err := logRepo.Start(task.ID)
		require.NoError(t, err)

		tagRepo := NewTagRepository(database)
		tag, err := tagRepo.FindOrCreate("work", "project")
		require.NoError(t, err)
		require.NoError(t, taskRepo.ReplaceTagsForTask(task.ID, []db.Tag{*tag}))

		// act
		require.NoError(t, taskRepo.Delete(task.ID))

		// assert
		_, err = taskRepo.FindByID(task.ID)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

		logs, err := logRepo.FindByTaskID(task.ID)
		require.NoError(t, err)
		assert.Empty(t, logs, "TimeLogs が消えていること")

		var count int64
		database.Table("task_tags").Where("task_id = ?", task.ID).Count(&count)
		assert.Zero(t, count, "task_tags が消えていること")
	})

	t.Run("SoftDelete", func(t *testing.T) {
		// arrange
		database := setupTestDB(t)
		repo := NewTaskRepository(database)

		task := &db.Task{Title: "ソフトデリート対象"}
		require.NoError(t, repo.Create(task))

		// act
		require.NoError(t, repo.Delete(task.ID))

		// assert
		_, err := repo.FindByID(task.ID)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "通常クエリでは見えないこと")

		var got db.Task
		require.NoError(t, database.Unscoped().First(&got, task.ID).Error)
		assert.True(t, got.DeletedAt.Valid, "deleted_at が設定されていること")
	})

	t.Run("SoftDelete_CascadesTimeLogs", func(t *testing.T) {
		// arrange
		database := setupTestDB(t)
		taskRepo := NewTaskRepository(database)
		logRepo := NewTimeLogRepository(database)

		task := &db.Task{Title: "削除対象"}
		require.NoError(t, taskRepo.Create(task))
		log, err := logRepo.Start(task.ID)
		require.NoError(t, err)

		// act
		require.NoError(t, taskRepo.Delete(task.ID))

		// assert
		logs, err := logRepo.FindByTaskID(task.ID)
		require.NoError(t, err)
		assert.Empty(t, logs, "通常クエリでは TimeLogs も見えないこと")

		var got db.TimeLog
		require.NoError(t, database.Unscoped().First(&got, log.ID).Error)
		assert.True(t, got.DeletedAt.Valid, "TimeLog の deleted_at が設定されていること")
	})
}

func TestTaskRepo_FindByID_NotFound(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	repo := NewTaskRepository(database)

	// act
	_, err := repo.FindByID(9999)

	// assert
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestTaskRepo_Update(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	repo := NewTaskRepository(database)

	task := &db.Task{Title: "元タイトル"}
	require.NoError(t, repo.Create(task))

	task.Title = "新タイトル"
	task.Priority = "A"

	// act
	require.NoError(t, repo.Update(task))

	// assert
	got, err := repo.FindByID(task.ID)
	require.NoError(t, err)
	assert.Equal(t, "新タイトル", got.Title)
	assert.Equal(t, "A", got.Priority)
}

func TestTaskRepo_FindAllWithTimeLogs(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	taskRepo := NewTaskRepository(database)
	logRepo := NewTimeLogRepository(database)

	active := &db.Task{Title: "未完了"}
	require.NoError(t, taskRepo.Create(active))
	startedLog, err := logRepo.Start(active.ID)
	require.NoError(t, err)
	require.NoError(t, logRepo.Stop(startedLog.ID))

	completed := &db.Task{Title: "完了済み"}
	require.NoError(t, taskRepo.Create(completed))
	require.NoError(t, taskRepo.ToggleComplete(completed.ID))

	deleted := &db.Task{Title: "削除済み"}
	require.NoError(t, taskRepo.Create(deleted))
	require.NoError(t, taskRepo.Delete(deleted.ID))

	// act
	tasks, err := taskRepo.FindAllWithTimeLogs()
	require.NoError(t, err)

	// assert
	assert.Len(t, tasks, 2, "未完了と完了済みの2件（削除済みは除外）")

	for _, task := range tasks {
		if task.Title == "未完了" {
			assert.Len(t, task.TimeLogs, 1)
		}
	}
}

func TestTaskRepo_FindAll_WithRecentCompleted(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	repo := NewTaskRepository(database)

	active := &db.Task{Title: "未完了"}
	require.NoError(t, repo.Create(active))

	recent := &db.Task{Title: "直近完了"}
	require.NoError(t, repo.Create(recent))
	require.NoError(t, repo.ToggleComplete(recent.ID))

	old := &db.Task{Title: "古い完了"}
	require.NoError(t, repo.Create(old))
	require.NoError(t, repo.ToggleComplete(old.ID))
	oldTime := time.Now().AddDate(0, -2, 0)
	database.Model(old).Update("completed_at", oldTime)

	from := time.Now().AddDate(0, -1, 0)

	// act
	tasks, err := repo.FindAll(WithRecentCompleted(from))
	require.NoError(t, err)

	// assert
	assert.Len(t, tasks, 2, "未完了と直近完了の2件が返ること")
	assert.Equal(t, "未完了", tasks[0].Title, "未完了タスクが先頭に来ること")
	assert.Equal(t, "直近完了", tasks[1].Title, "完了済みが末尾に来ること")
}

func TestTaskRepo_FindAll_WithRecentCompleted_SortOrder(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	repo := NewTaskRepository(database)

	older := &db.Task{Title: "先に完了"}
	require.NoError(t, repo.Create(older))
	require.NoError(t, repo.ToggleComplete(older.ID))
	olderTime := time.Now().Add(-2 * time.Hour)
	database.Model(older).Update("completed_at", olderTime)

	newer := &db.Task{Title: "後に完了"}
	require.NoError(t, repo.Create(newer))
	require.NoError(t, repo.ToggleComplete(newer.ID))

	from := time.Now().AddDate(0, -1, 0)

	// act
	tasks, err := repo.FindAll(WithRecentCompleted(from))
	require.NoError(t, err)

	// assert
	require.Len(t, tasks, 2)
	assert.Equal(t, "後に完了", tasks[0].Title, "直近完了が先頭に来ること")
	assert.Equal(t, "先に完了", tasks[1].Title)
}

func TestTaskRepo_StopAndToggleComplete(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	taskRepo := NewTaskRepository(database)
	logRepo := NewTimeLogRepository(database)

	task := &db.Task{Title: "タスク"}
	require.NoError(t, taskRepo.Create(task))

	log, err := logRepo.Start(task.ID)
	require.NoError(t, err)

	// act
	require.NoError(t, taskRepo.StopAndToggleComplete(log.ID, task.ID))

	// assert
	logs, err := logRepo.FindByTaskID(task.ID)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.NotNil(t, logs[0].EndAt, "計測が停止していること")

	updated, err := taskRepo.FindByID(task.ID)
	require.NoError(t, err)
	assert.True(t, updated.IsCompleted, "タスクが完了済みになっていること")
}

func TestTaskRepo_StopAndDelete(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	taskRepo := NewTaskRepository(database)
	logRepo := NewTimeLogRepository(database)

	task := &db.Task{Title: "タスク"}
	require.NoError(t, taskRepo.Create(task))

	log, err := logRepo.Start(task.ID)
	require.NoError(t, err)

	// act
	require.NoError(t, taskRepo.StopAndDelete(log.ID, task.ID))

	// assert
	tasks, err := taskRepo.FindAll()
	require.NoError(t, err)
	assert.Empty(t, tasks, "タスクが削除されていること")

	active, err := logRepo.FindActive()
	require.NoError(t, err)
	assert.Nil(t, active, "アクティブセッションがないこと")
}

func TestTaskRepo_StopAndSaveWithTags(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	taskRepo := NewTaskRepository(database)
	tagRepo := NewTagRepository(database)
	logRepo := NewTimeLogRepository(database)

	task := &db.Task{Title: "タスク"}
	require.NoError(t, taskRepo.Create(task))

	log, err := logRepo.Start(task.ID)
	require.NoError(t, err)

	tag, err := tagRepo.FindOrCreate("work", "project")
	require.NoError(t, err)

	task.Title = "更新後タイトル"
	task.IsCompleted = true
	now := time.Now()
	task.CompletedAt = &now

	// act
	require.NoError(t, taskRepo.StopAndSaveWithTags(log.ID, task, []db.Tag{*tag}))

	// assert
	logs, err := logRepo.FindByTaskID(task.ID)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.NotNil(t, logs[0].EndAt, "計測が停止していること")

	updated, err := taskRepo.FindByID(task.ID)
	require.NoError(t, err)
	assert.Equal(t, "更新後タイトル", updated.Title)
	assert.True(t, updated.IsCompleted)
	assert.Len(t, updated.Tags, 1, "タグが設定されていること")
}

func TestTaskRepo_SaveWithTags(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	taskRepo := NewTaskRepository(database)
	tagRepo := NewTagRepository(database)

	task := &db.Task{Title: "元タイトル"}
	require.NoError(t, taskRepo.Create(task))

	tag, err := tagRepo.FindOrCreate("work", "project")
	require.NoError(t, err)

	task.Title = "更新後タイトル"

	// act
	require.NoError(t, taskRepo.SaveWithTags(task, []db.Tag{*tag}))

	// assert
	updated, err := taskRepo.FindByID(task.ID)
	require.NoError(t, err)
	assert.Equal(t, "更新後タイトル", updated.Title, "タイトルが更新されていること")
	require.Len(t, updated.Tags, 1, "タグが1件設定されていること")
	assert.Equal(t, "work", updated.Tags[0].Name, "タグが正しいこと")

	// act
	require.NoError(t, taskRepo.SaveWithTags(task, []db.Tag{}))

	// assert
	updated2, err := taskRepo.FindByID(task.ID)
	require.NoError(t, err)
	assert.Empty(t, updated2.Tags, "タグが削除されていること")
}

func TestTaskRepo_ToggleComplete_NotFound(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	repo := NewTaskRepository(database)

	// act
	err := repo.ToggleComplete(9999)

	// assert
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "存在しないタスクIDの完了切替はエラーになること")
}

func TestTaskRepo_ReplaceTagsForTask_NotFound(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	repo := NewTaskRepository(database)

	// act
	err := repo.ReplaceTagsForTask(9999, nil)

	// assert
	assert.Error(t, err, "存在しないタスクでエラーになること")
}

func TestTaskRepo_Delete_NotFound(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	repo := NewTaskRepository(database)

	// act
	err := repo.Delete(9999)

	// assert
	assert.Error(t, err, "存在しないタスクIDの削除はエラーになること")
}

func TestTaskRepo_StopAndToggleComplete_TogglesBackToIncomplete(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	taskRepo := NewTaskRepository(database)
	logRepo := NewTimeLogRepository(database)

	task := &db.Task{Title: "タスク"}
	require.NoError(t, taskRepo.Create(task))
	require.NoError(t, taskRepo.ToggleComplete(task.ID))

	log, err := logRepo.Start(task.ID)
	require.NoError(t, err)

	// act
	require.NoError(t, taskRepo.StopAndToggleComplete(log.ID, task.ID))

	// assert
	logs, err := logRepo.FindByTaskID(task.ID)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.NotNil(t, logs[0].EndAt, "計測が停止していること")

	updated, err := taskRepo.FindByID(task.ID)
	require.NoError(t, err)
	assert.False(t, updated.IsCompleted, "タスクが未完了に戻っていること")
	assert.Nil(t, updated.CompletedAt, "CompletedAt がクリアされていること")
}

func TestTaskRepo_StopAndToggleComplete_AlreadyStopped(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	taskRepo := NewTaskRepository(database)
	logRepo := NewTimeLogRepository(database)

	task := &db.Task{Title: "タスク"}
	require.NoError(t, taskRepo.Create(task))

	log, err := logRepo.Start(task.ID)
	require.NoError(t, err)
	require.NoError(t, logRepo.Stop(log.ID))

	// act
	err = taskRepo.StopAndToggleComplete(log.ID, task.ID)

	// assert
	assert.Error(t, err, "停止済みセッションへのStopAndToggleCompleteはエラーになること")
}

func TestTaskRepo_StopAndDelete_AlreadyStopped(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	taskRepo := NewTaskRepository(database)
	logRepo := NewTimeLogRepository(database)

	task := &db.Task{Title: "タスク"}
	require.NoError(t, taskRepo.Create(task))

	log, err := logRepo.Start(task.ID)
	require.NoError(t, err)
	require.NoError(t, logRepo.Stop(log.ID))

	// act
	err = taskRepo.StopAndDelete(log.ID, task.ID)

	// assert
	assert.Error(t, err, "停止済みセッションへのStopAndDeleteはエラーになること")

	_, findErr := taskRepo.FindByID(task.ID)
	assert.NoError(t, findErr, "ロールバックによりタスクは削除されないこと")
}

func TestTaskRepo_StopAndSaveWithTags_AlreadyStopped(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	taskRepo := NewTaskRepository(database)
	logRepo := NewTimeLogRepository(database)

	task := &db.Task{Title: "タスク"}
	require.NoError(t, taskRepo.Create(task))

	log, err := logRepo.Start(task.ID)
	require.NoError(t, err)
	require.NoError(t, logRepo.Stop(log.ID))

	task.Title = "更新後タイトル"

	// act
	err = taskRepo.StopAndSaveWithTags(log.ID, task, []db.Tag{})

	// assert
	assert.Error(t, err, "停止済みセッションへのStopAndSaveWithTagsはエラーになること")

	original, findErr := taskRepo.FindByID(task.ID)
	require.NoError(t, findErr)
	assert.Equal(t, "タスク", original.Title, "ロールバックによりタイトルは更新されないこと")
}
