package timex_test

import (
	"testing"
	"time"

	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/goutil/timex"
)

func TestUtil_basic(t *testing.T) {
	sec := timex.NowUnix()

	assert.NotEmpty(t, timex.FormatUnix(sec))
	assert.NotEmpty(t, timex.FormatUnixBy(sec, time.RFC3339))

	tt := timex.TodayStart()
	assert.Eq(t, "00:00:00", timex.DateFormat(tt, "H:I:S"))

	tt = timex.TodayEnd()
	assert.Eq(t, "23:59:59", timex.DateFormat(tt, "H:I:S"))

	tt = timex.NowHourStart()
	assert.Eq(t, "00:00", timex.DateFormat(tt, "I:S"))

	tt = timex.NowHourEnd()
	assert.Eq(t, "59:59", timex.DateFormat(tt, "I:S"))
}

func TestNowAddDay(t *testing.T) {
	now := timex.Now()
	tt := timex.NowAddDay(1)
	assert.True(t, tt.Unix() > now.Unix())

	tt = timex.NowAddHour(-3)
	assert.Neq(t, tt.Hour(), now.Hour())
	assert.Eq(t, "3 hrs", timex.HowLongAgo(int64(now.DiffSec(tt)+2)))

	tt = timex.NowAddMinutes(5)
	assert.Neq(t, tt.Minute(), now.Minute())
}

func TestDateFormat(t *testing.T) {
	now := time.Now()

	tests := []struct{ layout, template string }{
		{"20060102 15:04:05", "Ymd H:I:S"},
		{"2006-01-02 15:04:05", "Y-m-d H:I:S"},
		{"2006-01-02 15:04", "Y-m-d H:I"},
		{"01/02 15:04:05", "m/d H:I:S"},
		{"06/01/02 15:04:05", "y/m/d H:I:S"},
		{"06/01/02 15:04:05.000", "y/m/d H:I:Sv"},
	}

	for i, item := range tests {
		date := now.Format(item.layout)
		assert.Eq(t, date, timex.DateFormat(now, item.template))
		if i%2 == 0 {
			assert.Eq(t, date, timex.Date(now, item.template))
		}
	}

	assert.Eq(t, now.Format("01/02 15:04:05.000000"), timex.Date(now, "m/d H:I:Su"))
}

func TestFormatUnix(t *testing.T) {
	now := time.Now()
	want := now.Format("2006-01-02 15:04:05")

	assert.Eq(t, want, timex.FormatUnix(now.Unix()))
	assert.Eq(t, want, timex.FormatUnixBy(now.Unix(), timex.DefaultLayout))
	assert.Eq(t, want, timex.FormatUnixByTpl(now.Unix(), "Y-m-d H:I:S"))
	// dump.P(want)

	assert.Eq(t, want, timex.Format(now))
	assert.Eq(t, want, timex.FormatBy(now, timex.DefaultLayout))
}

func TestToLayout(t *testing.T) {
	assert.Eq(t, timex.DefaultLayout, timex.ToLayout(""))
	assert.Eq(t, time.RFC3339, timex.ToLayout("c"))
	assert.Eq(t, time.RFC3339, timex.ToLayout("Y-m-dTH:I:SP"))
}

func TestToDur(t *testing.T) {
	dur, err := timex.ToDur("3s")
	assert.NoErr(t, err)
	assert.Eq(t, 3*timex.Second, dur)

	dur, err = timex.ToDur("now")
	assert.NoErr(t, err)
	assert.Eq(t, time.Duration(0), dur)

	dur, err = timex.ToDuration("0")
	assert.NoErr(t, err)
	assert.Eq(t, time.Duration(0), dur)

	assert.True(t, timex.IsDuration("3s"))
	assert.True(t, timex.IsDuration("3m"))
	assert.True(t, timex.IsDuration("-3h"))
	assert.True(t, timex.IsDuration("0"))
}

func TestTryToTime(t *testing.T) {
	tn := timex.Now()

	// duration string
	durTests := []struct {
		in  string
		out string
		ok  bool
	}{
		{"now", tn.Datetime(), true},
		{"0", tn.Datetime(), true},
		{"3s", tn.AddSeconds(3).Datetime(), true},
		{"3m", tn.AddMinutes(3).Datetime(), true},
	}

	for _, item := range durTests {
		tt, err := timex.TryToTime(item.in, tn.T())
		if item.ok {
			assert.NoErr(t, err)
		} else {
			assert.Err(t, err)
		}

		assert.Eq(t, item.out, timex.Format(tt))
	}

	bt := timex.ZeroTime
	assert.True(t, bt.IsZero())
	assert.Neq(t, 0, bt.Unix())

	noErrTests := []struct {
		in  string
		out string
	}{
		// date string
		{"2020-01-02 15:04:05", "2020-01-02 15:04:05"},
		{"2020-01-02", "2020-01-02 00:00:00"},
		{"2020-01-02 15:04", "2020-01-02 15:04:00"},
		{"2020-01-02 15", "2020-01-02 15:00:00"},
		{"2020-01-02 15:04:05.123", "2020-01-02 15:04:05"},
		{"2020-01-02 15:04:05.123456", "2020-01-02 15:04:05"},
		{"2020-01-02 15:04:05.123456789", "2020-01-02 15:04:05"},
		{"2020-01-02T15:04:05.123456789+08:00", "2020-01-02 15:04:05"},
	}

	for _, item := range noErrTests {
		tt, err := timex.TryToTime(item.in, bt)
		assert.NoErr(t, err)
		assert.Eq(t, item.out, timex.Format(tt))
	}
}

func TestParseRange(t *testing.T) {
	tests := []struct {
		input string
		start int64
		end   int64
	}{
		// {"2020-01-02 15:04:05", 1577977445, 0},
		// {"2020-01-02 15:04:05~2020-01-03 15:04:05", 1577977445, 1578063845},
		{"2020-01-02 15:04:06~", 1577977446, 0},
		{"~2020-01-02 15:04:07", 0, 1577946245},
		{"~", 0, 0},
	}

	for _, item := range tests {
		start, end, err := timex.ParseRange(item.input, nil)
		assert.NoErr(t, err)
		assert.Eq(t, item.start, start.Unix(), "start for %q", item.input)
		assert.Eq(t, item.end, end.Unix(), "end for %q", item.input)
	}
}
