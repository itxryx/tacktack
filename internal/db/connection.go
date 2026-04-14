package db

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB は指定パスで SQLite DB を開き、接続を返す。
// dbPath に ":memory:" を渡すとインメモリ DB になる（テスト用）。
func InitDB(dbPath string) (*gorm.DB, error) {
	if dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create db directory: %w", err)
		}
	}

	dsn := dbPath
	if dbPath != ":memory:" {
		dsn = dbPath + "?_journal_mode=WAL&_busy_timeout=5000"
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if dbPath == ":memory:" {
		// インメモリDBでもPRAGMAを設定
		if err := db.Exec("PRAGMA journal_mode=WAL").Error; err != nil {
			return nil, fmt.Errorf("failed to set journal_mode: %w", err)
		}
	}

	// H6: FK制約を有効化（SQLiteはデフォルト無効）
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return db, nil
}
