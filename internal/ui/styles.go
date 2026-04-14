package ui

import "github.com/charmbracelet/lipgloss"

var (
	// タスクライン
	StyleSelected  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4"))
	StyleCompleted = lipgloss.NewStyle().Strikethrough(true).Faint(true)
	StyleTracking  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	StylePriority  = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	StyleOverdue   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	// ヘッダー・フッター
	StyleHeader  = lipgloss.NewStyle().Bold(true).Padding(0, 1)
	StyleFooter  = lipgloss.NewStyle().Faint(true).Padding(0, 1)
	StyleWarning = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
	StyleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	// モーダル
	StyleModal      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	StyleModalTitle = lipgloss.NewStyle().Bold(true)

	// タイムライン
	StyleTimelineActive = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	StyleTimelineTime   = lipgloss.NewStyle().Faint(true)

	// タイムライン タスク色パレット
	TimelineColors = []lipgloss.Color{"4", "5", "6", "3", "13", "14", "9", "10"}
)
