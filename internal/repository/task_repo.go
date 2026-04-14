package repository

import (
	"fmt"
	"time"

	"github.com/itxryx/tacktack/internal/db"
	"gorm.io/gorm"
)

// GormTaskRepository は TaskRepository の GORM 実装。
type GormTaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(database *gorm.DB) TaskRepository {
	return &GormTaskRepository{db: database}
}

func (r *GormTaskRepository) Create(task *db.Task) error {
	return r.db.Create(task).Error
}

func (r *GormTaskRepository) FindAll(opts ...QueryOption) ([]db.Task, error) {
	cfg := &queryConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	q := r.db.Preload("Tags").Preload("TimeLogs")

	if cfg.recentCompletedFrom != nil {
		q = q.Where("is_completed = ? OR (is_completed = ? AND completed_at >= ?)", false, true, cfg.recentCompletedFrom)
	} else {
		q = q.Where("is_completed = ?", false)
	}

	q = q.Order(`
		is_completed ASC,
		CASE WHEN is_completed = 0 AND priority = '' THEN 1 ELSE 0 END ASC,
		CASE WHEN is_completed = 0 THEN priority ELSE NULL END ASC,
		CASE WHEN is_completed = 0 AND due_date IS NULL THEN 1 ELSE 0 END ASC,
		CASE WHEN is_completed = 0 THEN due_date ELSE NULL END ASC,
		CASE WHEN is_completed = 0 THEN created_at ELSE NULL END ASC,
		CASE WHEN is_completed = 1 THEN completed_at ELSE NULL END DESC
	`)

	var tasks []db.Task
	if err := q.Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("FindAll: %w", err)
	}
	return tasks, nil
}

func (r *GormTaskRepository) FindByID(id uint) (*db.Task, error) {
	var task db.Task
	if err := r.db.Preload("Tags").Preload("TimeLogs").First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *GormTaskRepository) Update(task *db.Task) error {
	return r.db.Save(task).Error
}

func (r *GormTaskRepository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. task_tags 中間テーブルを物理削除（中間テーブルはソフトデリート対象外）
		var task db.Task
		if err := tx.First(&task, id).Error; err != nil {
			return err
		}
		if err := tx.Model(&task).Association("Tags").Unscoped().Clear(); err != nil {
			return fmt.Errorf("clear task_tags: %w", err)
		}
		// 2. 関連 TimeLogs をソフトデリート
		if err := tx.Where("task_id = ?", id).Delete(&db.TimeLog{}).Error; err != nil {
			return fmt.Errorf("soft delete timelogs: %w", err)
		}
		// 3. Task 本体をソフトデリート
		if err := tx.Delete(&db.Task{}, id).Error; err != nil {
			return fmt.Errorf("soft delete task: %w", err)
		}
		return nil
	})
}

func (r *GormTaskRepository) ToggleComplete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var task db.Task
		if err := tx.First(&task, id).Error; err != nil {
			return err
		}
		if task.IsCompleted {
			task.IsCompleted = false
			task.CompletedAt = nil
		} else {
			task.IsCompleted = true
			now := time.Now()
			task.CompletedAt = &now
		}
		return tx.Save(&task).Error
	})
}

func (r *GormTaskRepository) FindAllWithTimeLogs() ([]db.Task, error) {
	var tasks []db.Task
	if err := r.db.Preload("Tags").Preload("TimeLogs").Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("FindAllWithTimeLogs: %w", err)
	}
	return tasks, nil
}

func (r *GormTaskRepository) ReplaceTagsForTask(taskID uint, tags []db.Tag) error {
	var task db.Task
	if err := r.db.First(&task, taskID).Error; err != nil {
		return err
	}
	return r.db.Model(&task).Association("Tags").Replace(tags)
}

// SaveWithTags はタスク保存とタグ置換をひとつのトランザクションでアトミックに実行する (H1)。
func (r *GormTaskRepository) SaveWithTags(task *db.Task, tags []db.Tag) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(task).Error; err != nil {
			return fmt.Errorf("save task: %w", err)
		}
		if err := tx.Model(task).Association("Tags").Replace(tags); err != nil {
			return fmt.Errorf("replace tags: %w", err)
		}
		return nil
	})
}

// StopAndToggleComplete は TimeLog の停止とタスク完了トグルを1トランザクションで実行する。
func (r *GormTaskRepository) StopAndToggleComplete(logID, taskID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		result := tx.Model(&db.TimeLog{}).
			Where("id = ? AND end_at IS NULL", logID).
			Update("end_at", now)
		if result.Error != nil {
			return fmt.Errorf("stop timelog: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("stop timelog: no active session found (id=%d)", logID)
		}
		var task db.Task
		if err := tx.First(&task, taskID).Error; err != nil {
			return err
		}
		if task.IsCompleted {
			task.IsCompleted = false
			task.CompletedAt = nil
		} else {
			task.IsCompleted = true
			t := time.Now()
			task.CompletedAt = &t
		}
		return tx.Save(&task).Error
	})
}

// StopAndDelete は TimeLog の停止とタスク削除を1トランザクションで実行する。
func (r *GormTaskRepository) StopAndDelete(logID, taskID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		result := tx.Model(&db.TimeLog{}).
			Where("id = ? AND end_at IS NULL", logID).
			Update("end_at", now)
		if result.Error != nil {
			return fmt.Errorf("stop timelog: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("stop timelog: no active session found (id=%d)", logID)
		}
		var task db.Task
		if err := tx.First(&task, taskID).Error; err != nil {
			return err
		}
		if err := tx.Model(&task).Association("Tags").Unscoped().Clear(); err != nil {
			return fmt.Errorf("clear task_tags: %w", err)
		}
		if err := tx.Where("task_id = ?", taskID).Delete(&db.TimeLog{}).Error; err != nil {
			return fmt.Errorf("soft delete timelogs: %w", err)
		}
		if err := tx.Delete(&db.Task{}, taskID).Error; err != nil {
			return fmt.Errorf("soft delete task: %w", err)
		}
		return nil
	})
}

// StopAndSaveWithTags は TimeLog の停止・タスク保存・タグ置換を1トランザクションで実行する (C3+H1)。
func (r *GormTaskRepository) StopAndSaveWithTags(logID uint, task *db.Task, tags []db.Tag) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		result := tx.Model(&db.TimeLog{}).
			Where("id = ? AND end_at IS NULL", logID).
			Update("end_at", now)
		if result.Error != nil {
			return fmt.Errorf("stop timelog: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("stop timelog: no active session found (id=%d)", logID)
		}
		if err := tx.Save(task).Error; err != nil {
			return fmt.Errorf("save task: %w", err)
		}
		if err := tx.Model(task).Association("Tags").Replace(tags); err != nil {
			return fmt.Errorf("replace tags: %w", err)
		}
		return nil
	})
}
