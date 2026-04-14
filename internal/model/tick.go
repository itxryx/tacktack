package model

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/itxryx/tacktack/internal/db"
)

// tickMsg は毎秒発火するメッセージ。計測中タスクの経過時間更新に使用する。
type tickMsg time.Time

// tickCmd は1秒後に tickMsg を送信するコマンドを返す。
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// initDoneMsg はアプリ起動時の初期化完了を通知するメッセージ。
type initDoneMsg struct {
	tasks     []db.Task
	activeLog *db.TimeLog
	nullLogs  []db.TimeLog
	err       error
}
