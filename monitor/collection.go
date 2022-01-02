package monitor

import "time"

type Collector interface {
	record(seq int, time *time.Time, latency *time.Duration)
	output(host string, startTime *time.Time, records *map[int]*time.Duration, report *Report, output string) error
}
