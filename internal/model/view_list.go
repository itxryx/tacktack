package model

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/itxryx/tacktack/internal/db"
	"github.com/itxryx/tacktack/internal/domain"
	"github.com/itxryx/tacktack/internal/ui"
)

func (m Model) viewList() string {
	if len(m.tasks) == 0 {
		msg := "タスクがありません。i でタスクを追加してください。"
		return "\n" + centerText(msg, m.width) + "\n"
	}

	var sb strings.Builder
	for i, task := range m.tasks {
		sb.WriteString(m.renderTaskRow(i, task))
		sb.WriteString("\n")
	}
	return sb.String()
}

func (m Model) renderTaskRow(i int, task db.Task) string {
	// カーソルマーカー
	cursor := "  "
	if i == m.cursor {
		cursor = "▶ "
	}

	// タグ
	var tags []string
	for _, tag := range task.Tags {
		if tag.Type == "project" {
			tags = append(tags, "+"+tag.Name)
		}
	}
	for _, tag := range task.Tags {
		if tag.Type == "context" {
			tags = append(tags, "@"+tag.Name)
		}
	}
	tagStr := ""
	if len(tags) > 0 {
		tagStr = " " + strings.Join(tags, " ")
	}

	// 締切日
	dueStr := ""
	if task.DueDate != nil {
		due := fmt.Sprintf("due:%d/%d", task.DueDate.Month(), task.DueDate.Day())
		if task.DueDate.Before(truncateDay(time.Now())) {
			dueStr = " " + ui.StyleOverdue.Render(due)
		} else {
			dueStr = " " + due
		}
	}

	// 累積トラッキング時間
	timerStr := ""
	if m.activeLog != nil && m.activeLog.TaskID == task.ID {
		pastSec := domain.CalcTotalSeconds(task.TimeLogs) // end_at=nil の現セッションは除外済み
		currentSec := int(time.Since(m.activeLog.StartAt).Seconds())
		totalSec := pastSec + currentSec
		h := totalSec / 3600
		mins := (totalSec % 3600) / 60
		secs := totalSec % 60
		totalElapsed := fmt.Sprintf("%02d:%02d:%02d", h, mins, secs)
		timerStr = " " + ui.StyleTracking.Render(fmt.Sprintf("[計測中 %s]", totalElapsed))
	} else {
		totalSec := domain.CalcTotalSeconds(task.TimeLogs)
		if totalSec > 0 {
			timerStr = " " + formatSeconds(totalSec)
		}
	}

	// 作成日（ゼロ値の場合は省略）
	createdAt := ""
	if !task.CreatedAt.IsZero() {
		createdAt = task.CreatedAt.Format("2006-01-02") + " "
	}

	var line string
	if task.IsCompleted {
		// todo.txt 完了形式: x 完了日 作成日 タイトル +tags @contexts due:date
		// 優先度は todo.txt 仕様に従い完了時は表示しない
		completionMark := "x"
		if task.CompletedAt != nil {
			completionMark += " " + task.CompletedAt.Format("2006-01-02")
		}
		title := ui.StyleCompleted.Render(task.Title)
		line = cursor + completionMark + " " + createdAt + title + tagStr + dueStr + timerStr
	} else {
		// todo.txt 未完了形式: (優先度) 作成日 タイトル +tags @contexts due:date
		priority := ""
		if task.Priority != "" {
			priority = ui.StylePriority.Render(fmt.Sprintf("(%s)", task.Priority)) + " "
		}
		line = cursor + priority + createdAt + task.Title + tagStr + dueStr + timerStr
	}

	// 端末幅に合わせて切り詰め
	if m.width > 0 {
		line = truncateString(line, m.width)
	}

	if i == m.cursor && !task.IsCompleted {
		line = ui.StyleSelected.Render(line)
	}
	if m.activeLog != nil && m.activeLog.TaskID == task.ID {
		line = ui.StyleTracking.Render(line)
	}

	return line
}

func centerText(s string, width int) string {
	if width <= 0 {
		return s
	}
	slen := utf8.RuneCountInString(s)
	if slen >= width {
		return s
	}
	pad := (width - slen) / 2
	return strings.Repeat(" ", pad) + s
}

// truncateString はターミナルの表示幅単位で文字列を指定幅に切り詰める。
// ANSIエスケープシーケンスは表示幅に含めず、CJK文字は2列として計算する。
func truncateString(s string, maxWidth int) string {
	if ansi.StringWidth(s) <= maxWidth {
		return s
	}
	return ansi.Truncate(s, maxWidth, "…")
}
