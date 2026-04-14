package domain

import (
	"time"

	"github.com/itxryx/tacktack/internal/db"
)

// SlotTask は30分スロット内に表示するタスク情報。
type SlotTask struct {
	TaskID   uint
	Title    string
	Tags     []string // "+work", "@office" 形式
	IsActive bool     // 現在計測中か
}

// TimeSlot は30分単位のタイムスロット。
type TimeSlot struct {
	Hour   int
	Minute int // 0 or 30
	Tasks  []SlotTask
}

// BuildTimelineSlots は指定日の TimeLogs を30分スロット（48個、0:00〜23:30）に分割して返す。
// 計測中のタスク（activeLog != nil）は time.Now() を EndAt として仮計算する。
func BuildTimelineSlots(tasks []db.Task, date time.Time, activeLog *db.TimeLog) []TimeSlot {
	// 指定日のローカル時間範囲（date はローカル時間で渡される）
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	dayEnd := dayStart.Add(24 * time.Hour)

	// 全48スロットを初期化
	slots := make([]TimeSlot, 48)
	for i := range slots {
		slots[i] = TimeSlot{
			Hour:   i / 2,
			Minute: (i % 2) * 30,
		}
	}

	for _, task := range tasks {
		// タグ文字列を構築
		var tagStrs []string
		for _, tag := range task.Tags {
			prefix := "+"
			if tag.Type == "context" {
				prefix = "@"
			}
			tagStrs = append(tagStrs, prefix+tag.Name)
		}

		for _, log := range task.TimeLogs {
			logStart := log.StartAt
			var logEnd time.Time
			if log.EndAt == nil {
				// 計測中: activeLog と TaskID が一致する場合のみ現在時刻を使用
				if activeLog != nil && activeLog.ID == log.ID {
					logEnd = time.Now()
				} else {
					continue
				}
			} else {
				logEnd = *log.EndAt
			}

			// 指定日と重ならないログはスキップ
			if logEnd.Before(dayStart) || logStart.After(dayEnd) {
				continue
			}
			// 指定日範囲にクランプ
			if logStart.Before(dayStart) {
				logStart = dayStart
			}
			if logEnd.After(dayEnd) {
				logEnd = dayEnd
			}

			isActive := activeLog != nil && activeLog.ID == log.ID

			// ログが重なるスロットを特定
			for i := range slots {
				slotStart := dayStart.Add(time.Duration(i*30) * time.Minute)
				slotEnd := slotStart.Add(30 * time.Minute)
				if logEnd.After(slotStart) && logStart.Before(slotEnd) {
					slots[i].Tasks = append(slots[i].Tasks, SlotTask{
						TaskID:   task.ID,
						Title:    task.Title,
						Tags:     tagStrs,
						IsActive: isActive,
					})
				}
			}
		}
	}
	return slots
}
