package workerpool_test

import (
	"errors"
	"runtime"
	"testing"
	"time"

	"git-clone-manager/internal/workerpool"
)

func TestRunReturnsResultsInInputOrder(t *testing.T) {
	items := []int{1, 2, 3}

	results := workerpool.Run(items, func(item int) (string, error) {
		if item == 1 {
			time.Sleep(30 * time.Millisecond)
		}
		if item == 2 {
			time.Sleep(10 * time.Millisecond)
		}

		return string(rune('0' + item)), nil
	})

	if len(results) != len(items) {
		t.Fatalf("len(results) = %d, want %d", len(results), len(items))
	}

	for index, want := range []string{"1", "2", "3"} {
		if results[index].Value != want {
			t.Fatalf("results[%d].Value = %q, want %q", index, results[index].Value, want)
		}
		if results[index].Err != nil {
			t.Fatalf("results[%d].Err = %v, want nil", index, results[index].Err)
		}
	}
}

func TestRunPreservesPerItemErrorsWithoutAbortingOtherItems(t *testing.T) {
	items := []int{1, 2, 3}
	wantErr := errors.New("boom")

	results := workerpool.Run(items, func(item int) (int, error) {
		if item == 2 {
			return 0, wantErr
		}

		return item * 10, nil
	})

	if results[0].Value != 10 || results[0].Err != nil {
		t.Fatalf("results[0] = %#v, want value 10 with nil error", results[0])
	}

	if !errors.Is(results[1].Err, wantErr) {
		t.Fatalf("results[1].Err = %v, want %v", results[1].Err, wantErr)
	}

	if results[2].Value != 30 || results[2].Err != nil {
		t.Fatalf("results[2] = %#v, want value 30 with nil error", results[2])
	}
}

func TestRunProcessesItemsInParallel(t *testing.T) {
	items := make([]int, min(2*runtime.NumCPU(), 4))
	sleepPerItem := 40 * time.Millisecond

	start := time.Now()
	results := workerpool.Run(items, func(item int) (int, error) {
		time.Sleep(sleepPerItem)
		return item, nil
	})
	elapsed := time.Since(start)

	if len(results) != len(items) {
		t.Fatalf("len(results) = %d, want %d", len(results), len(items))
	}

	sequentialDuration := time.Duration(len(items)) * sleepPerItem
	if elapsed >= sequentialDuration-(sleepPerItem/2) {
		t.Fatalf("Run took %v, want meaningfully less than sequential duration %v", elapsed, sequentialDuration)
	}
}
