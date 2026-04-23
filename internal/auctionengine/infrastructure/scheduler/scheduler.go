package scheduler

import "time"

type TimerScheduler struct {
}

func NewTimerScheduler() *TimerScheduler {
	return &TimerScheduler{}
}

func (f *TimerScheduler) Schedule(at time.Time, job func()) error {
	delay := time.Until(at)

	if delay <= 0 {
		go job()
		return nil
	}

	time.AfterFunc(delay, job)
	return nil
}
