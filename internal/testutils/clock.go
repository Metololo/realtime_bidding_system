package testutils

import "time"

type FakeClock struct {
	currentTime time.Time
}

func NewFakeClock(now time.Time) *FakeClock {
	return &FakeClock{currentTime: now}
}

func (f *FakeClock) Now() time.Time {
	return f.currentTime
}

func (f *FakeClock) Advance(d time.Duration) {
	f.currentTime = f.currentTime.Add(d)
}
