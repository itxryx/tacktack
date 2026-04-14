package model

import (
	"fmt"

	"github.com/itxryx/tacktack/internal/ui"
)

func (m Model) viewDeleteConfirm() string {
	title := ""
	for _, t := range m.tasks {
		if t.ID == m.deleteTargetID {
			title = t.Title
			break
		}
	}

	content := fmt.Sprintf(
		"%s\n\n\"%s\"\n\n  [y/Enter] 削除する   [n/Esc] キャンセル",
		ui.StyleModalTitle.Render("タスクを削除しますか？"),
		title,
	)
	return "\n" + ui.StyleModal.Render(content)
}

func (m Model) viewTrackingAlert() string {
	currentTitle := ""
	if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
		currentTitle = m.tasks[m.cursor].Title
	}
	conflictTitle := ""
	for _, t := range m.tasks {
		if t.ID == m.conflictTaskID {
			conflictTitle = t.Title
			break
		}
	}

	content := fmt.Sprintf(
		"%s\n\n\"%s\" を計測停止して\n\"%s\" を開始しますか？\n\n  [y] 切り替える   [n/Esc] キャンセル",
		ui.StyleModalTitle.Render("別のタスクを計測中です"),
		conflictTitle,
		currentTitle,
	)
	return "\n" + ui.StyleModal.Render(content)
}
