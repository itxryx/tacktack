package model

import (
	"fmt"
	"time"

	"github.com/itxryx/tacktack/internal/db"
	"github.com/itxryx/tacktack/internal/repository"
)

// refreshTasks はリポジトリからタスク一覧を再取得して m.tasks を更新する。
// 直近1ヶ月以内に完了したタスクも含めて取得する。
// C6: statsTasks キャッシュも無効化する。
func (m *Model) refreshTasks() error {
	from := time.Now().AddDate(0, -1, 0)
	tasks, err := m.taskRepo.FindAll(repository.WithRecentCompleted(from))
	if err != nil {
		return fmt.Errorf("refreshTasks: %w", err)
	}
	m.tasks = tasks
	m.statsTasks = nil // C6: 統計キャッシュを無効化して次回遷移時に再取得させる
	return nil
}

// refreshNullLogs は DB から end_at IS NULL のログを再取得して m.nullLogs を更新する (C7)。
// アクティブなセッション（m.activeLog）は除外する。
func (m *Model) refreshNullLogs() error {
	nulls, err := m.timeLogRepo.FindNullEndAt()
	if err != nil {
		return fmt.Errorf("refreshNullLogs: %w", err)
	}
	var filtered []db.TimeLog
	for _, log := range nulls {
		if m.activeLog == nil || log.ID != m.activeLog.ID {
			filtered = append(filtered, log)
		}
	}
	m.nullLogs = filtered
	return nil
}

// dedupTags は同一 ID のタグを除いた新しいスライスを返す。
// 同名でも Type が異なるタグ（例: +task と @task）は別タグとして扱われる。
func dedupTags(tags []db.Tag) []db.Tag {
	seen := make(map[uint]bool, len(tags))
	result := make([]db.Tag, 0, len(tags))
	for _, t := range tags {
		if !seen[t.ID] {
			seen[t.ID] = true
			result = append(result, t)
		}
	}
	return result
}

// containsTagID は tags の中に指定 ID のタグが存在するかを返す。
func containsTagID(tags []db.Tag, id uint) bool {
	for _, t := range tags {
		if t.ID == id {
			return true
		}
	}
	return false
}

// tagTypeLabel は tag.Type（"project" / "context"）を日本語ラベルに変換する。
func tagTypeLabel(tagType string) string {
	if tagType == "context" {
		return "コンテキスト"
	}
	return "プロジェクト"
}

// clampCursor はカーソル位置をタスク数の範囲内に収める。
func (m *Model) clampCursor() {
	if len(m.tasks) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(m.tasks) {
		m.cursor = len(m.tasks) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}
