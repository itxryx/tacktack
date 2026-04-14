package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/itxryx/tacktack/internal/db"
	"github.com/itxryx/tacktack/internal/domain"
	"github.com/itxryx/tacktack/internal/ui"
)

func (m Model) viewTimeline() string {
	date := m.timelineDate
	if date.IsZero() {
		date = time.Now()
	}

	dateStr := date.Format("2006-01-02 (Mon)")
	w := m.width
	if w <= 0 {
		w = 80 // フォールバック幅
	}
	separator := strings.Repeat("─", w)

	var sb strings.Builder
	fmt.Fprintf(&sb, "1日のタイムライン: %s\n%s\n", dateStr, separator)

	slots := domain.BuildTimelineSlots(m.tasks, date, m.activeLog)

	// 表示可能な行数。View() 全体の固定オーバーヘッドを除く:
	// global header(1) + timeline header(1) + separator(1) + blank前合計(1) + 合計(1) + blank(1) + footer(1) = 7行
	visibleLines := timelineVisibleLines(m.height)

	// タスクID → 色インデックスのマッピングを構築
	colorIdx := map[uint]int{}
	nextColor := 0
	for _, slot := range slots {
		for _, st := range slot.Tasks {
			if _, ok := colorIdx[st.TaskID]; !ok {
				colorIdx[st.TaskID] = nextColor % len(ui.TimelineColors)
				nextColor++
			}
		}
	}

	// スクロール範囲を計算（timelineScroll はフラット行インデックス）
	totalRows := countFlatRows(slots)
	maxStart := timelineMaxScroll(totalRows, m.height)
	start := m.timelineScroll
	if start > maxStart {
		start = maxStart
	}
	if start < 0 {
		start = 0
	}

	// スロットをフラット行インデックスで描画する。
	// rowIdx が start に達するまでスキップし、そこから visibleLines 行を描画する。
	// スキップ中も shownInSlot を更新し、showLabel の正確性を保つ。
	rowIdx := 0
	renderedLines := 0
	for _, slot := range slots {
		if renderedLines >= visibleLines {
			break
		}
		timeLabel := fmt.Sprintf("%02d:%02d", slot.Hour, slot.Minute)
		timePart := ui.StyleTimelineTime.Render(timeLabel + " │")

		if len(slot.Tasks) == 0 {
			if rowIdx >= start {
				sb.WriteString(timePart + "\n")
				renderedLines++
			}
			rowIdx++
		} else {
			shownInSlot := map[uint]bool{} // 現スロット内で既にラベルを表示したタスクID
			for subIdx, st := range slot.Tasks {
				if rowIdx >= start && renderedLines < visibleLines {
					showLabel := !shownInSlot[st.TaskID]
					if subIdx == 0 {
						// 先頭行: 時刻ラベルを付与
						taskLine := renderTimelineTask(st, colorIdx[st.TaskID], showLabel)
						sb.WriteString(timePart + " " + taskLine + "\n")
					} else {
						// 2行目以降: 字下げ
						indent := strings.Repeat(" ", len(timeLabel)+2) // "HH:MM │" の幅
						sb.WriteString(indent + " " + renderTimelineTask(st, colorIdx[st.TaskID], showLabel) + "\n")
					}
					renderedLines++
				}
				shownInSlot[st.TaskID] = true
				rowIdx++
			}
		}
	}

	// 合計時間（日をまたぐログは当日分のみ計上）
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	dayEnd := dayStart.Add(24 * time.Hour)
	totalSec := domain.CalcTotalSecondsInRange(collectTimeLogsForDate(m.tasks, date), dayStart, dayEnd)
	fmt.Fprintf(&sb, "\n合計: %s\n", formatSeconds(totalSec))

	return sb.String()
}

// renderTimelineTask は1スロット内の1タスク表示文字列を返す。
// showLabel が false の場合はバーのみ表示（連続スロットでの重複省略用）。
func renderTimelineTask(st domain.SlotTask, cIdx int, showLabel bool) string {
	color := ui.TimelineColors[cIdx]
	bar := lipgloss.NewStyle().Foreground(color).Render("████")

	if !showLabel {
		return bar
	}

	// タグ文字列
	tags := ""
	if len(st.Tags) > 0 {
		tags = " " + strings.Join(st.Tags, " ")
	}

	label := st.Title + tags
	if st.IsActive {
		label = ui.StyleTimelineActive.Render(label + " ◀ 計測中")
	}
	return bar + " " + label
}

// collectTimeLogsForDate は指定日と重なる TimeLogs を全タスクから収集する。
func collectTimeLogsForDate(tasks []db.Task, date time.Time) []db.TimeLog {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	dayEnd := dayStart.Add(24 * time.Hour)

	var logs []db.TimeLog
	for _, task := range tasks {
		for _, log := range task.TimeLogs {
			if log.EndAt == nil {
				continue
			}
			if log.EndAt.Before(dayStart) || log.StartAt.After(dayEnd) {
				continue
			}
			logs = append(logs, log)
		}
	}
	return logs
}

// timelineVisibleLines は viewTimeline() で描画可能なスロット行数を返す。
// オーバーヘッド7行（global header / timeline header / separator / blank前合計 / 合計 / blank / footer）を除く。
func timelineVisibleLines(height int) int {
	v := height - 7
	if v < 5 {
		return 48 // フォールバック: 全スロット表示
	}
	if v > 48 {
		return 48
	}
	return v
}

// slotRowCount はスロットが描画する行数を返す。
// タスクがない場合は時刻ラベル行のみで 1、タスクがある場合はタスク数分の行を返す。
func slotRowCount(slot domain.TimeSlot) int {
	if len(slot.Tasks) == 0 {
		return 1
	}
	return len(slot.Tasks)
}

// countFlatRows はスロットリスト全体の描画行数（フラット行数）を返す。
func countFlatRows(slots []domain.TimeSlot) int {
	total := 0
	for _, slot := range slots {
		total += slotRowCount(slot)
	}
	return total
}

// timelineMaxScroll はスクロール位置の上限（maxStart）を返す。
// totalRows はスロットリストの描画行数合計（countFlatRows で算出）。
func timelineMaxScroll(totalRows, height int) int {
	max := totalRows - timelineVisibleLines(height)
	if max < 0 {
		return 0
	}
	return max
}

// firstTaskScroll は指定日にタスクが存在する最初のフラット行インデックスを返す。
// termHeight を渡すことで viewTimeline() と同じオーバーヘッド計算を行い、
// クランプ後も最初のタスクが表示領域の先頭に来るよう調整する。
// タスクが存在しない場合は 12:00 相当のフラット行インデックスを返す。
func firstTaskScroll(tasks []db.Task, date time.Time, activeLog *db.TimeLog, termHeight int) int {
	slots := domain.BuildTimelineSlots(tasks, date, activeLog)
	totalRows := countFlatRows(slots)
	maxStart := timelineMaxScroll(totalRows, termHeight)

	rowIdx := 0
	for _, slot := range slots {
		if len(slot.Tasks) > 0 {
			if rowIdx > maxStart {
				return maxStart
			}
			return rowIdx
		}
		rowIdx += slotRowCount(slot)
	}

	// タスクなし: 12:00 = スロット24 のフラット行インデックスを計算
	// スロット0〜23 がすべて空の場合、各スロット1行なのでフラット行24 = 12:00
	rowIdx = 0
	for i, slot := range slots {
		if i == 24 {
			break
		}
		rowIdx += slotRowCount(slot)
	}
	if rowIdx > maxStart {
		return maxStart
	}
	return rowIdx
}
