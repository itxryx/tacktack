package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/itxryx/tacktack/internal/db"
	"github.com/itxryx/tacktack/internal/domain"
	"github.com/itxryx/tacktack/internal/ui"
)

var statsTagPeriods = []struct {
	period domain.PeriodType
	label  string
}{
	{domain.PeriodDay, "今日"},
	{domain.PeriodWeek, "今週"},
	{domain.PeriodMonth, "今月"},
	{domain.PeriodHalfYear, "半年"},
	{domain.PeriodYear, "1年"},
}

func (m Model) viewStats() string {
	now := time.Now()
	todayStart, todayEnd := domain.CalcDateRange(domain.PeriodDay, now)
	weekStart, weekEnd := domain.CalcDateRange(domain.PeriodWeek, now)
	monthStart, monthEnd := domain.CalcDateRange(domain.PeriodMonth, now)

	var todaySec, weekSec, monthSec int
	var todayCount, weekCount, monthCount int

	for _, task := range m.tasks {
		for _, log := range task.TimeLogs {
			if log.EndAt == nil {
				continue
			}
			logs := []db.TimeLog{log}
			if s := domain.CalcTotalSecondsInRange(logs, todayStart, todayEnd); s > 0 {
				todaySec += s
				todayCount++
			}
			if s := domain.CalcTotalSecondsInRange(logs, weekStart, weekEnd); s > 0 {
				weekSec += s
				weekCount++
			}
			if s := domain.CalcTotalSecondsInRange(logs, monthStart, monthEnd); s > 0 {
				monthSec += s
				monthCount++
			}
		}
	}

	var sb strings.Builder
	sb.WriteString("統計・セッション情報（Tab: タイムラインに戻る）\n\n")

	sb.WriteString("── タイムトラッキング ─────────────────────\n")
	fmt.Fprintf(&sb, "  今日    : %s\n", formatSeconds(todaySec))
	fmt.Fprintf(&sb, "  今週    : %s\n", formatSeconds(weekSec))
	fmt.Fprintf(&sb, "  今月    : %s\n", formatSeconds(monthSec))

	sb.WriteString("\n── セッション数 ───────────────────────────\n")
	fmt.Fprintf(&sb, "  今日    : %d 件\n", todayCount)
	fmt.Fprintf(&sb, "  今週    : %d 件\n", weekCount)
	fmt.Fprintf(&sb, "  今月    : %d 件\n", monthCount)

	// タグ別計測時間割合（期間選択付き）
	p := statsTagPeriods[m.statsTagPeriodIdx]
	from, to := domain.CalcDateRange(p.period, now)

	// statsTasks が取得済みなら使用、なければ m.tasks にフォールバック
	statsSrc := m.statsTasks
	if statsSrc == nil {
		statsSrc = m.tasks
	}
	tagStats := domain.CalcTagTimeStats(statsSrc, from, to)

	fmt.Fprintf(&sb, "\n── プロジェクト/コンテキスト別計測時間 [%s] (h/l: 期間変更) ─────\n", p.label)
	if len(tagStats) == 0 {
		sb.WriteString("  この期間の計測データはありません\n")
	} else {
		const barWidth = 20
		for _, ts := range tagStats {
			prefix := "+"
			if ts.TagType == "context" {
				prefix = "@"
			}
			name := fmt.Sprintf("%-14s", prefix+ts.TagName)
			filled := int(ts.Percentage * float64(barWidth))
			if filled > barWidth {
				filled = barWidth
			}
			bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
			fmt.Fprintf(&sb, "  %s %s %3.0f%%  %s\n", name, bar, ts.Percentage*100, formatSeconds(ts.TotalSec))
		}
		sb.WriteString("  ※複数タグのタスクは各タグに計上\n")
	}

	// 異常セッション
	sb.WriteString("\n── 異常なセッション ───────────────────────\n")
	if len(m.nullLogs) == 0 {
		sb.WriteString("  異常なセッションはありません\n")
	} else {
		sb.WriteString(ui.StyleWarning.Render(fmt.Sprintf("  ⚠ %d件の異常なセッションがあります", len(m.nullLogs))) + "\n")
		for i, log := range m.nullLogs {
			cursor := "  "
			if i == m.statsCursor {
				cursor = "▶ "
			}
			// タスクタイトルを探す
			taskTitle := fmt.Sprintf("タスクID:%d", log.TaskID)
			for _, t := range m.tasks {
				if t.ID == log.TaskID {
					taskTitle = t.Title
					break
				}
			}
			line := fmt.Sprintf("%s%s %s〜 [未終了]  \"%s\"",
				cursor,
				log.StartAt.In(time.Local).Format("2006-01-02"),
				log.StartAt.In(time.Local).Format("15:04"),
				taskTitle,
			)
			sb.WriteString(line + "\n")
		}
		sb.WriteString("  e: 該当タスクの詳細編集へ\n")
	}

	return sb.String()
}

func formatSeconds(sec int) string {
	if sec == 0 {
		return "0秒"
	}
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	if h == 0 && m == 0 {
		return fmt.Sprintf("%d秒", s)
	}
	if h == 0 {
		return fmt.Sprintf("%d分 %02d秒", m, s)
	}
	return fmt.Sprintf("%d時間 %02d分 %02d秒", h, m, s)
}
