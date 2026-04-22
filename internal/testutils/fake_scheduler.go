package testutils

import "time"

type FakeManualScheduler struct {
	ScheduledAt   time.Time
	ScheduledJobs []func()
}

func (f *FakeManualScheduler) Schedule(at time.Time, job func()) error {
	f.ScheduledJobs = append(f.ScheduledJobs, job)
	return nil
}

func (f *FakeManualScheduler) ExecuteLastScheduledTask() {
	if len(f.ScheduledJobs) == 0 {
		return
	}
	lastJob := f.ScheduledJobs[len(f.ScheduledJobs)-1]
	lastJob()
}
