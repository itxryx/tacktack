package db

import (
	"time"

	"gorm.io/gorm"
)

// Task はタスク管理の中心となるモデル。
type Task struct {
	ID          uint           `gorm:"primaryKey"`
	Priority    string         `gorm:"size:1"`           // "A"〜"Z"、空文字列は優先度なし
	Title       string         `gorm:"not null"`
	DueDate     *time.Time                               // NULL = 締切なし
	IsCompleted bool           `gorm:"default:false"`
	CompletedAt *time.Time                               // NULL = 未完了
	CreatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	Tags        []Tag          `gorm:"many2many:task_tags;"`
	TimeLogs    []TimeLog      `gorm:"foreignKey:TaskID"`
}

// Tag はプロジェクト（+）またはコンテキスト（@）のタグ。
type Tag struct {
	ID        uint           `gorm:"primaryKey"`
	Name      string         `gorm:"uniqueIndex;not null"`
	Type      string         `gorm:"not null"` // "project" or "context"
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TimeLog はタスクへの時間計測ログ。EndAt が NULL の場合は計測中または異常終了。
type TimeLog struct {
	ID        uint           `gorm:"primaryKey"`
	TaskID    uint           `gorm:"not null;index"`
	StartAt   time.Time      `gorm:"not null"`
	EndAt     *time.Time     // NULL = 計測中または異常終了
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
