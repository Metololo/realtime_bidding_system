package infrastructure

import "time"

type SystemClock struct {
}

func NewSystemClock() *SystemClock {
	return &SystemClock{}
}

func (s *SystemClock) Now() time.Time {
	return time.Now()
}
