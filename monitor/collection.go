package monitor

import "time"

type Collector interface {
	record(seq int, latency *time.Duration)
	output(output string, startTime *time.Time, records map[int]*time.Duration, report *Report) error
}
