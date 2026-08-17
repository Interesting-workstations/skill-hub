package admin

import (
	"testing"
	"time"
)

// TestTaskDue 验证定时调度器的到期判断逻辑。
func TestTaskDue(t *testing.T) {
	// 固定一个基准时间：2026-08-15 03:00:00
	base := time.Date(2026, 8, 15, 3, 0, 0, 0, time.Local)
	yesterday := base.AddDate(0, 0, -1)
	_ = yesterday

	cases := []struct {
		name     string
		schedule string
		lastAuto time.Time
		now      time.Time
		want     bool
	}{
		// 每天 03:00
		{"daily-once-never", "每天 03:00", time.Time{}, base, true},
		{"daily-once-today-already", "每天 03:00", base, base, false},
		{"daily-once-yesterday", "每天 03:00", time.Date(2026, 8, 14, 3, 0, 0, 0, time.Local), base, true},
		{"daily-not-match-time", "每天 03:00", time.Time{}, base.Add(1 * time.Minute), false},
		// 每 6 小时
		{"every6h-never", "每 6 小时", time.Time{}, base, true},
		{"every6h-5h-ago", "每 6 小时", base.Add(-5 * time.Hour), base, false},
		{"every6h-6h-ago", "每 6 小时", base.Add(-6 * time.Hour), base, true},
		// 每小时
		{"hourly-once-never", "每小时", time.Time{}, base, true},
		{"hourly-not-on-hour", "每小时", time.Time{}, base.Add(15 * time.Minute), false},
		{"hourly-already-this-hour", "每小时", base, base, false},
		// 手动 / 空
		{"manual", "手动", time.Time{}, base, false},
		{"empty", "", time.Time{}, base, false},
	}
	for _, c := range cases {
		if got := taskDue(c.schedule, c.lastAuto, c.now); got != c.want {
			t.Fatalf("taskDue(%q, last=%v, now=%v) = %v，期望 %v", c.schedule, c.lastAuto.Format("15:04"), c.now.Format("15:04"), got, c.want)
		}
	}
}
