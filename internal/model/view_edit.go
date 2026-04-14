package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/itxryx/tacktack/internal/domain"
	"github.com/itxryx/tacktack/internal/ui"
)

func (m Model) viewEdit() string {
	if m.editTask == nil {
		return "読み込み中..."
	}

	var sb strings.Builder
	sb.WriteString("タスク詳細編集（Enter: 保存 / Esc: キャンセル）\n\n")

	// タスク属性エリア
	fields := []struct {
		label string
		idx   int
	}{
		{"タイトル ", editFieldTitle},
		{"優先度  ", editFieldPriority},
		{"締切日  ", editFieldDueDate},
	}

	for _, f := range fields {
		prefix := "  "
		if !m.editLogFocus && m.editField == f.idx {
			prefix = "▶ "
		}
		fmt.Fprintf(&sb, "%s%s: %s\n", prefix, f.label, m.editInputs[f.idx].View())
	}

	// タグフィールド
	tagPrefix := "  "
	tagFocused := !m.editLogFocus && m.editField == editFieldTags
	if tagFocused {
		tagPrefix = "▶ "
	}
	var tagStrs []string
	for i, tag := range m.selectedTags {
		p := "+"
		if tag.Type == "context" {
			p = "@"
		}
		name := p + tag.Name
		if tagFocused && i == m.editTagCursor {
			name = "[" + name + "]"
		}
		tagStrs = append(tagStrs, name)
	}
	tagDisplay := strings.Join(tagStrs, " ")
	if tagDisplay == "" {
		tagDisplay = "(なし)"
	}
	tagHint := "[+: プロジェクト  @: コンテキスト]"
	if tagFocused && len(m.selectedTags) > 0 {
		tagHint = "[+: プロジェクト  @: コンテキスト  j/k: 移動  d: 削除]"
	}
	fmt.Fprintf(&sb, "%sタグ      : %s  %s\n", tagPrefix, tagDisplay, tagHint)

	// 完了状態
	completePrefix := "  "
	if !m.editLogFocus && m.editField == editFieldComplete {
		completePrefix = "▶ "
	}
	completeMark := "[ ]"
	if m.editTask.IsCompleted {
		completeMark = "[x]"
	}
	fmt.Fprintf(&sb, "%s完了済み  : %s （Space でトグル）\n", completePrefix, completeMark)

	// TimeLog エリア
	sb.WriteString("\n──────────── 計測ログ ─────────────── Tab で移動\n")

	logs := m.editTask.TimeLogs
	if len(logs) == 0 {
		sb.WriteString("  （ログなし）\n")
	} else {
		for i, log := range logs {
			cursor := "  "
			if m.editLogFocus && i == m.editLogCursor {
				cursor = "▶ "
			}

			var logLine string
			if m.editLogEditing == i {
				// 編集中
				startView := m.editLogInputs[0].View()
				endView := m.editLogInputs[1].View()
				logLine = fmt.Sprintf("%s開始: %s  終了: %s", cursor, startView, endView)
			} else if log.EndAt == nil {
				logLine = fmt.Sprintf("%s%s 〜 [未終了]",
					cursor,
					log.StartAt.In(time.Local).Format("2006-01-02 15:04:05"),
				)
				logLine = ui.StyleWarning.Render(logLine)
			} else {
				seconds := int(log.EndAt.Sub(log.StartAt).Seconds())
				logLine = fmt.Sprintf("%s%s 〜 %s  (%s)",
					cursor,
					log.StartAt.In(time.Local).Format("2006-01-02 15:04:05"),
					log.EndAt.In(time.Local).Format("15:04:05"),
					formatSeconds(seconds),
				)
			}
			sb.WriteString(logLine + "\n")
		}
	}

	totalSec := domain.CalcTotalSeconds(logs)
	sb.WriteString("                               [e: 編集 / d: 削除]\n")
	fmt.Fprintf(&sb, "合計: %s\n", formatSeconds(totalSec))

	if m.lastErr != nil {
		sb.WriteString("\n" + ui.StyleError.Render("✗ "+m.lastErr.Error()) + "\n")
	}

	return sb.String()
}
