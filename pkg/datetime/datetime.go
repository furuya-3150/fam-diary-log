package datetime

import (
	"time"
)

// GetWeekRange は指定された日付を含む週の月曜日から日曜日までの範囲を返す
// 月曜日を週の開始、日曜日を週の終了とする
func GetWeekRange(date time.Time) (time.Time, time.Time) {
	weekday := date.Weekday()

	daysToMonday := int(weekday) - 1
	if weekday == 0 { // 日曜日の場合
		daysToMonday = 6
	}
	monday := date.AddDate(0, 0, -daysToMonday)

	// 日曜日の日付を計算
	sunday := monday.AddDate(0, 0, 6)

	// 時刻を設定（月曜日は 00:00:00、日曜日は 23:59:59）
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
	sunday = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 999999999, sunday.Location())

	return monday, sunday
}
