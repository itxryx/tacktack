package db

import (
	"fmt"

	"gorm.io/gorm"
)

// Migrate は Task, Tag, TimeLog テーブルを AutoMigrate で作成・更新する。
// task_tags 中間テーブルは GORM が Task の Tags フィールドから自動生成する。
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&Task{}, &Tag{}, &TimeLog{}); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	return nil
}
