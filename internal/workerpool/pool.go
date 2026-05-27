package workerpool

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
)

type Result[T any] struct {
	Value T
	Err   error
}

type indexedItem[T any] struct {
	index int
	item  T
}

func Run[Item any, Output any](items []Item, work func(Item) (Output, error)) []Result[Output] {
	results := make([]Result[Output], len(items))
	if len(items) == 0 {
		return results
	}

	jobs := make(chan indexedItem[Item])

	var waitGroup sync.WaitGroup
	workerCount := 2 * runtime.NumCPU()
	for range workerCount {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			for job := range jobs {
				results[job.index] = runWork(job.item, work)
			}
		}()
	}

	for index, item := range items {
		jobs <- indexedItem[Item]{
			index: index,
			item:  item,
		}
	}
	close(jobs)

	waitGroup.Wait()
	return results
}

func runWork[Item any, Output any](item Item, work func(Item) (Output, error)) (result Result[Output]) {
	defer func() {
		if recovered := recover(); recovered != nil {
			result.Err = fmt.Errorf("worker panic for item %v: %v\n%s", item, recovered, debug.Stack())
		}
	}()

	value, err := work(item)
	return Result[Output]{
		Value: value,
		Err:   err,
	}
}
