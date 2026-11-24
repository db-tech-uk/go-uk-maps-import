package rates

import (
	"fmt"
	"sort"
	"time"

	"github.com/db-tech-uk/go-utils/timeutils"
)

type RateInfo struct {
	ShapeFile string
	Records   int
	Rows      map[string]int
	Duration  time.Duration
}

func (r RateInfo) String() string {
	return fmt.Sprintf("\n[%v] Records %v Rows %v Duration %v, Avg %v/s",
		r.ShapeFile,
		r.Records,
		r.Rows,
		r.Durn(),
		r.Rate(),
	)
}

func (r RateInfo) Avg() float64 {
	return calcRate(r)
}

func (r RateInfo) Rate() string {
	rate := r.Avg()
	rateStr := fmt.Sprintf("%v", int64(rate))

	if rate < 1 {
		rateStr = fmt.Sprintf("%.2f", rate)
	}

	return rateStr
}

func (r RateInfo) Durn() string {
	precision := 2

	if r.Duration.Seconds() < 1 {
		precision = 3
	}

	return timeutils.FormatDuration(timeutils.Round(r.Duration, precision), precision)
}

type RatesInfo []RateInfo

func (r RatesInfo) String() string {
	var s string

	sort.Slice(r, func(i, j int) bool {
		return r[i].ShapeFile < r[j].ShapeFile
	})

	for _, k := range r {
		s += fmt.Sprintf("%v", k)
	}

	return s
}
