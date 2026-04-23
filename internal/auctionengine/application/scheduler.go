package application

import "time"

type Scheduler interface {
	Schedule(at time.Time, job func()) error
}
