package domain

import (
	"fmt"
	"strings"
	"time"
)

// ParseDueDate は入力文字列から日付を解析して返す。
// 認識できるキーワードが含まれない場合は nil, nil を返す（締切なし）。
// now を引数で受け取ることでテスト時の日付固定を可能にする。
func ParseDueDate(input string, now time.Time) (*time.Time, error) {
	s := strings.ToLower(strings.TrimSpace(input))

	var result time.Time
	switch {
	case s == "today":
		result = truncateToDay(now)
	case s == "tomorrow":
		result = truncateToDay(now).AddDate(0, 0, 1)
	case isWeekday(s):
		result = nextWeekday(now, parseWeekday(s), false)
	case strings.HasPrefix(s, "next "):
		wd := strings.TrimPrefix(s, "next ")
		if !isWeekday(wd) {
			return nil, fmt.Errorf("invalid weekday: %q", wd)
		}
		result = nextWeekday(now, parseWeekday(wd), true)
	default:
		// YYYY-MM-DD 形式を試みる
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			// 認識できないキーワードはエラー（due: プレフィックスが付いた場合に呼ばれる想定）
			return nil, fmt.Errorf("cannot parse date: %q", input)
		}
		result = truncateToDay(t)
	}

	return &result, nil
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

var weekdayNames = map[string]time.Weekday{
	"mon": time.Monday,
	"tue": time.Tuesday,
	"wed": time.Wednesday,
	"thu": time.Thursday,
	"fri": time.Friday,
	"sat": time.Saturday,
	"sun": time.Sunday,
}

func isWeekday(s string) bool {
	_, ok := weekdayNames[s]
	return ok
}

func parseWeekday(s string) time.Weekday {
	return weekdayNames[s]
}

// nextWeekday は now から見て次の wd 曜日の日付を返す。
// forceNext=false: 当日なら当日を返し、過去なら次回（7日以内）の発生日を返す。
// forceNext=true : 当日・未来曜日は来週の発生日（+7）、過去曜日は次の発生日（7日以内）を返す。
func nextWeekday(now time.Time, wd time.Weekday, forceNext bool) time.Time {
	today := truncateToDay(now)
	diff := int(wd) - int(today.Weekday())
	if diff < 0 {
		// 今週すでに過ぎた曜日 → 次回発生日（7日以内）
		diff += 7
	} else if diff == 0 && forceNext {
		// 当日かつ強制翌週 → 来週同曜日
		diff += 7
	} else if diff == 0 {
		// 当日（forceNext=false）→ 当日
		return today
	} else if forceNext {
		// 今週まだ来ていない曜日 + forceNext → 来週の発生日
		diff += 7
	}
	return today.AddDate(0, 0, diff)
}
