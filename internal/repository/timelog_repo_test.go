package repository

import (
	"testing"
	"time"

	"github.com/itxryx/tacktack/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeLogRepo_Start(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	taskRepo := NewTaskRepository(database)
	logRepo := NewTimeLogRepository(database)

	task := &db.Task{Title: "タスク"}
	require.NoError(t, taskRepo.Create(task))

	// act
	log, err := logRepo.Start(task.ID)
	require.NoError(t, err)

	// assert
	assert.NotZero(t, log.ID)
	assert.Nil(t, log.EndAt)
	assert.Equal(t, task.ID, log.TaskID)
}

func TestTimeLogRepo_Stop(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	taskRepo := NewTaskRepository(database)
	logRepo := NewTimeLogRepository(database)

	task := &db.Task{Title: "タスク"}
	require.NoError(t, taskRepo.Create(task))

	log, err := logRepo.Start(task.ID)
	require.NoError(t, err)

	// act
	require.NoError(t, logRepo.Stop(log.ID))

	// assert
	logs, err := logRepo.FindByTaskID(task.ID)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.NotNil(t, logs[0].EndAt)
}

func TestTimeLogRepo_Stop_AlreadyStopped(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	taskRepo := NewTaskRepository(database)
	logRepo := NewTimeLogRepository(database)

	task := &db.Task{Title: "タスク"}
	require.NoError(t, taskRepo.Create(task))

	log, err := logRepo.Start(task.ID)
	require.NoError(t, err)
	require.NoError(t, logRepo.Stop(log.ID))

	// act + assert
	assert.Error(t, logRepo.Stop(log.ID), "停止済みセッションへのStopはエラーであること")
}

func TestTimeLogRepo_FindActive(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	taskRepo := NewTaskRepository(database)
	logRepo := NewTimeLogRepository(database)

	// act
	active, err := logRepo.FindActive()
	require.NoError(t, err)

	// assert
	assert.Nil(t, active)

	// arrange
	task := &db.Task{Title: "タスク"}
	require.NoError(t, taskRepo.Create(task))

	log, err := logRepo.Start(task.ID)
	require.NoError(t, err)

	// act
	active, err = logRepo.FindActive()
	require.NoError(t, err)

	// assert
	require.NotNil(t, active)
	assert.Equal(t, log.ID, active.ID)
}

func TestTimeLogRepo_Update(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	taskRepo := NewTaskRepository(database)
	logRepo := NewTimeLogRepository(database)

	task := &db.Task{Title: "タスク"}
	require.NoError(t, taskRepo.Create(task))

	log, err := logRepo.Start(task.ID)
	require.NoError(t, err)
	assert.Nil(t, log.EndAt)

	newEnd := time.Now()
	log.EndAt = &newEnd

	// act
	require.NoError(t, logRepo.Update(log))

	// assert
	logs, err := logRepo.FindByTaskID(task.ID)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.NotNil(t, logs[0].EndAt)
}

func TestTimeLogRepo_Delete(t *testing.T) {
	t.Run("通常削除", func(t *testing.T) {
		// arrange
		database := setupTestDB(t)
		taskRepo := NewTaskRepository(database)
		logRepo := NewTimeLogRepository(database)

		task := &db.Task{Title: "タスク"}
		require.NoError(t, taskRepo.Create(task))

		log, err := logRepo.Start(task.ID)
		require.NoError(t, err)

		// act
		require.NoError(t, logRepo.Delete(log.ID))

		// assert
		logs, err := logRepo.FindByTaskID(task.ID)
		require.NoError(t, err)
		assert.Empty(t, logs, "FindByTaskID から消えていること")
	})

	t.Run("SoftDelete確認", func(t *testing.T) {
		// arrange
		database := setupTestDB(t)
		taskRepo := NewTaskRepository(database)
		logRepo := NewTimeLogRepository(database)

		task := &db.Task{Title: "タスク"}
		require.NoError(t, taskRepo.Create(task))
		log, err := logRepo.Start(task.ID)
		require.NoError(t, err)
		require.NoError(t, logRepo.Delete(log.ID))

		// act + assert
		logs, err := logRepo.FindByTaskID(task.ID)
		require.NoError(t, err)
		assert.Empty(t, logs, "通常クエリでは見えないこと")

		var got db.TimeLog
		require.NoError(t, database.Unscoped().First(&got, log.ID).Error)
		assert.True(t, got.DeletedAt.Valid, "deleted_at が設定されていること")
	})
}

func TestTimeLogRepo_StopAndStart(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	taskRepo := NewTaskRepository(database)
	logRepo := NewTimeLogRepository(database)

	task1 := &db.Task{Title: "タスク1"}
	task2 := &db.Task{Title: "タスク2"}
	require.NoError(t, taskRepo.Create(task1))
	require.NoError(t, taskRepo.Create(task2))

	log1, err := logRepo.Start(task1.ID)
	require.NoError(t, err)

	// act
	log2, err := logRepo.StopAndStart(log1.ID, task2.ID)
	require.NoError(t, err)

	// assert
	require.NotNil(t, log2)
	assert.Equal(t, task2.ID, log2.TaskID)
	assert.Nil(t, log2.EndAt, "新しいセッションは計測中であること")

	logs1, err := logRepo.FindByTaskID(task1.ID)
	require.NoError(t, err)
	require.Len(t, logs1, 1)
	assert.NotNil(t, logs1[0].EndAt, "タスク1の計測が停止していること")

	active, err := logRepo.FindActive()
	require.NoError(t, err)
	require.NotNil(t, active)
	assert.Equal(t, log2.ID, active.ID)
}

func TestTimeLogRepo_StopAndStart_AlreadyStopped(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	taskRepo := NewTaskRepository(database)
	logRepo := NewTimeLogRepository(database)

	task1 := &db.Task{Title: "タスク1"}
	task2 := &db.Task{Title: "タスク2"}
	require.NoError(t, taskRepo.Create(task1))
	require.NoError(t, taskRepo.Create(task2))

	log1, err := logRepo.Start(task1.ID)
	require.NoError(t, err)
	require.NoError(t, logRepo.Stop(log1.ID))

	// act
	_, err = logRepo.StopAndStart(log1.ID, task2.ID)

	// assert
	assert.Error(t, err, "停止済みセッションへのStopAndStartはエラーであること")

	logs2, err := logRepo.FindByTaskID(task2.ID)
	require.NoError(t, err)
	assert.Empty(t, logs2, "ロールバックによりタスク2のログは作成されないこと")
}

func TestTimeLogRepo_FindByTaskID_WithMultipleLogs(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	taskRepo := NewTaskRepository(database)
	logRepo := NewTimeLogRepository(database)

	task := &db.Task{Title: "タスク"}
	require.NoError(t, taskRepo.Create(task))

	log1, err := logRepo.Start(task.ID)
	require.NoError(t, err)
	require.NoError(t, logRepo.Stop(log1.ID))

	log2, err := logRepo.Start(task.ID)
	require.NoError(t, err)
	require.NoError(t, logRepo.Stop(log2.ID))

	log3, err := logRepo.Start(task.ID)
	require.NoError(t, err)

	// act
	logs, err := logRepo.FindByTaskID(task.ID)
	require.NoError(t, err)

	// assert
	assert.Len(t, logs, 3, "3件のログが返ること")
	assert.Equal(t, log1.ID, logs[0].ID, "最初のログが先頭であること")
	assert.Equal(t, log2.ID, logs[1].ID, "2番目のログが2番目であること")
	assert.Equal(t, log3.ID, logs[2].ID, "3番目のログが末尾であること")
}

func TestTimeLogRepo_Start_NonexistentTask(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	logRepo := NewTimeLogRepository(database)

	// act
	_, err := logRepo.Start(9999)

	// assert
	assert.Error(t, err, "存在しないTaskIDへのStartはFKエラーになること")
}

func TestTimeLogRepo_FindNullEndAt(t *testing.T) {
	// arrange
	database := setupTestDB(t)
	taskRepo := NewTaskRepository(database)
	logRepo := NewTimeLogRepository(database)

	task := &db.Task{Title: "タスク"}
	require.NoError(t, taskRepo.Create(task))

	_, err := logRepo.Start(task.ID)
	require.NoError(t, err)
	log2, err := logRepo.Start(task.ID)
	require.NoError(t, err)
	require.NoError(t, logRepo.Stop(log2.ID))

	// act
	nullLogs, err := logRepo.FindNullEndAt()
	require.NoError(t, err)

	// assert
	assert.Len(t, nullLogs, 1)
}
