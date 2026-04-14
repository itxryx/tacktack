package domain

import "time"

// PeriodType は集計期間の種類を表す。
type PeriodType int

const (
	PeriodDay PeriodType = iota
	PeriodWeek
	PeriodMonth
	PeriodHalfYear
	PeriodYear
	PeriodAll
)

// CalcDateRange は期間タイプと基準日時から検索範囲の開始・終了を返す。
// PeriodAll の場合は from = time.Time{} (zero value)、to = now を返す。
func CalcDateRange(period PeriodType, now time.Time) (from, to time.Time) {
	today := truncateToDay(now)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.Local)

	switch period {
	case PeriodDay:
		return today, endOfDay
	case PeriodWeek:
		// 週の月曜日を計算
		weekday := int(today.Weekday())
		if weekday == 0 {
			weekday = 7 // 日曜日を7に
		}
		monday := today.AddDate(0, 0, -(weekday - 1))
		return monday, endOfDay
	case PeriodMonth:
		firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
		return firstDay, endOfDay
	case PeriodHalfYear:
		halfYearAgo := now.AddDate(0, -6, 0)
		firstDay := truncateToDay(halfYearAgo)
		return firstDay, endOfDay
	case PeriodYear:
		firstDay := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.Local)
		return firstDay, endOfDay
	case PeriodAll:
		return time.Time{}, now
	default:
		return time.Time{}, now
	}
}
