package repository

import (
	"time"

	"github.com/itxryx/tacktack/internal/db"
)

// TaskRepository はタスクのCRUD操作を抽象化する。
type TaskRepository interface {
	Create(task *db.Task) error
	FindAll(opts ...QueryOption) ([]db.Task, error)
	FindByID(id uint) (*db.Task, error) // Tags と TimeLogs を Preload する
	Update(task *db.Task) error
	Delete(id uint) error                                               // ソフトデリート（関連 TimeLog も論理削除、task_tags は物理削除）
	ToggleComplete(id uint) error                                       // is_completed と completed_at をトランザクションで更新
	ReplaceTagsForTask(taskID uint, tags []db.Tag) error                // タグの全置換
	SaveWithTags(task *db.Task, tags []db.Tag) error                    // H1: タスク保存とタグ置換をトランザクションでアトミックに実行
	StopAndToggleComplete(logID, taskID uint) error                     // TimeLog停止とToggleCompleteを1トランザクションで実行
	StopAndDelete(logID, taskID uint) error                             // TimeLog停止とタスク削除を1トランザクションで実行
	StopAndSaveWithTags(logID uint, task *db.Task, tags []db.Tag) error // TimeLog停止とタスク保存/タグ置換を1トランザクションで実行
	FindAllWithTimeLogs() ([]db.Task, error)                            // 統計用: 全タスク（完了含む）を Tags/TimeLogs と共に取得
}

// TagRepository はタグのCRUD操作を抽象化する。
type TagRepository interface {
	FindAll() ([]db.Tag, error)
	FindByType(tagType string) ([]db.Tag, error) // "project" or "context"
	FindOrCreate(name string, tagType string) (*db.Tag, error)
	Delete(id uint) error // ソフトデリート（task_tags の紐付けは物理削除）
}

// TimeLogRepository はタイムログのCRUD操作を抽象化する。
type TimeLogRepository interface {
	Start(taskID uint) (*db.TimeLog, error)                     // 新規セッション開始
	Stop(logID uint) error                                      // end_at を現在時刻で更新
	StopAndStart(stopID, startTaskID uint) (*db.TimeLog, error) // Stop と Start を1トランザクションで実行
	FindActive() (*db.TimeLog, error)                           // end_at IS NULL かつ最新1件を返す
	FindNullEndAt() ([]db.TimeLog, error)                       // end_at IS NULL の全件
	FindByTaskID(taskID uint) ([]db.TimeLog, error)
	Update(log *db.TimeLog) error
	Delete(id uint) error
}

// QueryOption は FindAll のクエリ条件を設定する関数型。
type QueryOption func(*queryConfig)

type queryConfig struct {
	recentCompletedFrom *time.Time // nil = 完了済みを含めない
}

// WithRecentCompleted は指定日時以降に完了したタスクも含める（一覧表示用）。
func WithRecentCompleted(from time.Time) QueryOption {
	return func(c *queryConfig) {
		c.recentCompletedFrom = &from
	}
}
