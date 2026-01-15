package clock

import "time"

type Clock interface {
	Now() time.Time
}

// 実際の時刻を返す
type Real struct{}

func (r *Real) Now() time.Time {
	return time.Now()
}

type Fixed struct {
	Time time.Time
}

// 固定時刻を返す（テスト用）
func (f *Fixed) Now() time.Time {
	return f.Time
}
