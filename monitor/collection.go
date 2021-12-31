package monitor

import "time"

type Collection interface {
	record(time *time.Time, delay int)
	out(target Output) error
}

type Output interface {
}