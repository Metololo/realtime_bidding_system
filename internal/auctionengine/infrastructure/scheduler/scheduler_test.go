package scheduler

import (
	"sync"
	"testing"
	"time"
)

func TestGoTimerScheduler_Execution(t *testing.T) {
	scheduler := NewTimerScheduler()
	executed := make(chan bool)

	delay := 10 * time.Millisecond

	err := scheduler.Schedule(time.Now().Add(delay), func() {
		close(executed)
	})

	if err != nil {
		t.Fatalf("Failed to schedule: %v", err)
	}

	select {
	case <-executed:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Task was never executed by the real timer")
	}
}

func TestTimerScheduler_ConcurrentTasks(t *testing.T) {
	s := NewTimerScheduler()
	const taskCount = 100
	var wg sync.WaitGroup
	wg.Add(taskCount)

	for i := 0; i < taskCount; i++ {
		delay := time.Duration(i%10) * time.Millisecond
		err := s.Schedule(time.Now().Add(delay), func() {
			wg.Done()
		})
		if err != nil {
			t.Errorf("Failed to schedule task %d: %v", i, err)
		}
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Not all concurrent tasks were executed in time")
	}
}
