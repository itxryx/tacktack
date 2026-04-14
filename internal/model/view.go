package model

import (
	"fmt"

	"github.com/itxryx/tacktack/internal/ui"
)

// View はモードに応じてサブビュー関数を呼び出す。
func (m Model) View() string {
	header := m.viewHeader()
	content := m.viewContent()
	footer := m.viewFooter()
	return header + "\n" + content + "\n" + footer
}

func (m Model) viewHeader() string {
	title := ui.StyleHeader.Render("tacktack")
	if len(m.nullLogs) > 0 {
		warning := ui.StyleWarning.Render(
			fmt.Sprintf("⚠ 異常な計測セッションが %d 件あります (Tab で確認)", len(m.nullLogs)),
		)
		return title + "  " + warning
	}
	return title
}

func (m Model) viewContent() string {
	switch m.mode {
	case modeList:
		return m.viewList()
	case modeInput:
		return m.viewInput()
	case modeTagSelect:
		return m.viewTagSelect()
	case modeEditDetail:
		return m.viewEdit()
	case modeDeleteConfirm:
		return m.viewDeleteConfirm()
	case modeTrackingAlert:
		return m.viewTrackingAlert()
	case modeStats:
		return m.viewStats()
	case modeTimeline:
		return m.viewTimeline()
	}
	return ""
}

func (m Model) viewFooter() string {
	var hint string
	switch m.mode {
	case modeList:
		spaceHint := "  Space: 計測"
		if len(m.tasks) > 0 && m.cursor < len(m.tasks) && m.tasks[m.cursor].IsCompleted {
			spaceHint = ""
		}
		hint = "j/k: 移動  x: 完了  d/BS: 削除  i/a: 新規  e: 編集" + spaceHint + "  Tab: タイムライン  q: 終了"
	case modeInput:
		hint = "Tab/Shift+Tab: フィールド移動  Enter/Ctrl+S: 保存  Esc: キャンセル"
	case modeTagSelect:
		hint = "j/k: 移動  Enter: 選択  D: 削除  BS: フィルタ削除  Esc: 戻る  文字: フィルタ"
	case modeEditDetail:
		if m.editLogFocus {
			if m.editLogEditing >= 0 {
				hint = "Tab: 項目切替  Enter: 保存  Esc: キャンセル"
			} else {
				hint = "j/k: 移動  e: 編集  d: 削除  Shift+Tab: タスクに戻る  Esc: 一覧に戻る"
			}
		} else {
			hint = "Tab/Shift+Tab: フィールド移動  Enter/Ctrl+S: 保存  Esc: キャンセル"
		}
	case modeDeleteConfirm:
		hint = "y/Enter: 削除  n/Esc: キャンセル"
	case modeTrackingAlert:
		hint = "y/Enter: 切り替える  n/Esc: キャンセル"
	case modeStats:
		hint = "j/k: 異常セッション移動  h/l: 期間切替  e: 詳細編集  Tab/Esc: 一覧に戻る  q: 終了"
	case modeTimeline:
		hint = "h/l: 日付移動  j/k: スクロール  t: 今日  Tab: 統計  Esc: 一覧  q: 終了"
	}
	if m.lastErr != nil {
		errMsg := ui.StyleError.Render("エラー: " + m.lastErr.Error())
		return errMsg + "\n" + ui.StyleFooter.Render(hint)
	}
	return ui.StyleFooter.Render(hint)
}
