package repository

import (
	"fmt"
	"time"

	"github.com/itxryx/tacktack/internal/db"
	"gorm.io/gorm"
)

// GormTimeLogRepository は TimeLogRepository の GORM 実装。
type GormTimeLogRepository struct {
	db *gorm.DB
}

func NewTimeLogRepository(database *gorm.DB) TimeLogRepository {
	return &GormTimeLogRepository{db: database}
}

func (r *GormTimeLogRepository) Start(taskID uint) (*db.TimeLog, error) {
	log := &db.TimeLog{
		TaskID:  taskID,
		StartAt: time.Now(),
		EndAt:   nil,
	}
	if err := r.db.Create(log).Error; err != nil {
		return nil, fmt.Errorf("Start timelog: %w", err)
	}
	return log, nil
}

func (r *GormTimeLogRepository) Stop(logID uint) error {
	now := time.Now()
	result := r.db.Model(&db.TimeLog{}).
		Where("id = ? AND end_at IS NULL", logID).
		Update("end_at", now)
	if result.Error != nil {
		return fmt.Errorf("Stop timelog: %w", result.Error)
	}
	// 0行更新はセッションが存在しないか既に停止済みを意味する
	if result.RowsAffected == 0 {
		return fmt.Errorf("Stop timelog: no active session found (id=%d)", logID)
	}
	return nil
}

func (r *GormTimeLogRepository) StopAndStart(stopID, startTaskID uint) (*db.TimeLog, error) {
	var newLog *db.TimeLog
	err := r.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		result := tx.Model(&db.TimeLog{}).
			Where("id = ? AND end_at IS NULL", stopID).
			Update("end_at", now)
		if result.Error != nil {
			return fmt.Errorf("stop timelog: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("stop timelog: no active session found (id=%d)", stopID)
		}
		log := &db.TimeLog{TaskID: startTaskID, StartAt: time.Now(), EndAt: nil}
		if err := tx.Create(log).Error; err != nil {
			return fmt.Errorf("start timelog: %w", err)
		}
		newLog = log
		return nil
	})
	if err != nil {
		return nil, err
	}
	return newLog, nil
}

func (r *GormTimeLogRepository) FindActive() (*db.TimeLog, error) {
	var log db.TimeLog
	err := r.db.Where("end_at IS NULL").Order("start_at DESC").First(&log).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("FindActive: %w", err)
	}
	return &log, nil
}

func (r *GormTimeLogRepository) FindNullEndAt() ([]db.TimeLog, error) {
	var logs []db.TimeLog
	if err := r.db.Where("end_at IS NULL").Order("start_at ASC").Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("FindNullEndAt: %w", err)
	}
	return logs, nil
}

func (r *GormTimeLogRepository) FindByTaskID(taskID uint) ([]db.TimeLog, error) {
	var logs []db.TimeLog
	if err := r.db.Where("task_id = ?", taskID).Order("start_at ASC").Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("FindByTaskID: %w", err)
	}
	return logs, nil
}

func (r *GormTimeLogRepository) Update(log *db.TimeLog) error {
	return r.db.Save(log).Error
}

func (r *GormTimeLogRepository) Delete(id uint) error {
	return r.db.Delete(&db.TimeLog{}, id).Error
}
