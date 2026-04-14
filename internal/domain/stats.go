package domain

import (
	"sort"
	"time"

	"github.com/itxryx/tacktack/internal/db"
)

// CalcTotalSeconds は TimeLogs の合計時間を秒単位で返す。
// end_at が nil のログは計算から除外する。
func CalcTotalSeconds(logs []db.TimeLog) int {
	total := 0
	for _, log := range logs {
		if log.EndAt == nil {
			continue
		}
		total += int(log.EndAt.Sub(log.StartAt).Seconds())
	}
	return total
}

// CalcTotalSecondsInRange は from〜to 期間内の合計時間を秒単位で返す。
// 期間をまたぐログは from/to でクランプした重なり部分のみ計上する。
// end_at が nil のログは計算から除外する。
func CalcTotalSecondsInRange(logs []db.TimeLog, from, to time.Time) int {
	total := 0
	for _, log := range logs {
		if log.EndAt == nil {
			continue
		}
		logStart := log.StartAt
		logEnd := *log.EndAt
		if logEnd.Before(from) || logStart.After(to) {
			continue
		}
		if logStart.Before(from) {
			logStart = from
		}
		if logEnd.After(to) {
			logEnd = to
		}
		total += int(logEnd.Sub(logStart).Seconds())
	}
	return total
}

// HasNullEndAt は end_at が nil のログが1件以上あるかを返す。
func HasNullEndAt(logs []db.TimeLog) bool {
	for _, log := range logs {
		if log.EndAt == nil {
			return true
		}
	}
	return false
}

// DailyStat は1日の統計情報を保持する。
type DailyStat struct {
	Date         time.Time
	TotalSeconds int
	TaskCount    int
}

// TagTimeStat はタグごとの計測時間統計。
type TagTimeStat struct {
	TagType    string  // "project" or "context"
	TagName    string  // タグ名
	TotalSec   int     // 期間内の合計秒数
	Percentage float64 // 全タグ合計に対する割合 (0.0〜1.0)
}

// CalcTagTimeStats は指定期間内のタグ別計測時間割合を計算して返す。
// 複数タグを持つタスクの計測時間は各タグにそれぞれ計上する。
// タグなしのタスクは集計から除外する。
// 結果は TotalSec 降順でソートされる。
func CalcTagTimeStats(tasks []db.Task, from, to time.Time) []TagTimeStat {
	tagSec := map[string]int{}     // "project:work" -> 秒数
	tagMeta := map[string][2]string{} // "project:work" -> [TagType, TagName]

	for _, task := range tasks {
		if len(task.Tags) == 0 {
			continue
		}
		// この期間に該当するログの秒数を集計（期間との重なり部分をクランプ）
		taskPeriodSec := CalcTotalSecondsInRange(task.TimeLogs, from, to)
		if taskPeriodSec == 0 {
			continue
		}
		for _, tag := range task.Tags {
			key := tag.Type + ":" + tag.Name
			tagSec[key] += taskPeriodSec
			tagMeta[key] = [2]string{tag.Type, tag.Name}
		}
	}

	var totalSec int
	for _, sec := range tagSec {
		totalSec += sec
	}

	result := make([]TagTimeStat, 0, len(tagSec))
	for key, sec := range tagSec {
		meta := tagMeta[key]
		pct := 0.0
		if totalSec > 0 {
			pct = float64(sec) / float64(totalSec)
		}
		result = append(result, TagTimeStat{
			TagType:    meta[0],
			TagName:    meta[1],
			TotalSec:   sec,
			Percentage: pct,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].TotalSec != result[j].TotalSec {
			return result[i].TotalSec > result[j].TotalSec
		}
		if result[i].TagType != result[j].TagType {
			return result[i].TagType < result[j].TagType
		}
		return result[i].TagName < result[j].TagName
	})
	return result
}

// CalcDailyStats は TimeLogs を日別に集計して返す。
// 日をまたぐログは各日の境界でクランプして分割計上する。
// end_at が nil のログは除外する。
func CalcDailyStats(logs []db.TimeLog) []DailyStat {
	type key struct{ y, m, d int }
	statMap := map[key]*DailyStat{}
	var order []key

	for _, log := range logs {
		if log.EndAt == nil {
			continue
		}
		startLocal := log.StartAt.In(time.Local)
		endLocal := log.EndAt.In(time.Local)
		startDay := time.Date(startLocal.Year(), startLocal.Month(), startLocal.Day(), 0, 0, 0, 0, time.Local)
		endDay := time.Date(endLocal.Year(), endLocal.Month(), endLocal.Day(), 0, 0, 0, 0, time.Local)

		// ログが複数日にまたがる場合、各日の境界でクランプして分割計上する
		for d := startDay; !d.After(endDay); d = d.AddDate(0, 0, 1) {
			dayStart := d
			dayEnd := d.Add(24 * time.Hour)

			logStart := log.StartAt
			logEnd := *log.EndAt
			if logStart.Before(dayStart) {
				logStart = dayStart
			}
			if logEnd.After(dayEnd) {
				logEnd = dayEnd
			}
			sec := int(logEnd.Sub(logStart).Seconds())
			if sec <= 0 {
				continue
			}

			k := key{d.Year(), int(d.Month()), d.Day()}
			if _, exists := statMap[k]; !exists {
				statMap[k] = &DailyStat{
					Date: d,
				}
				order = append(order, k)
			}
			statMap[k].TotalSeconds += sec
			statMap[k].TaskCount++
		}
	}

	result := make([]DailyStat, 0, len(order))
	for _, k := range order {
		result = append(result, *statMap[k])
	}
	return result
}
