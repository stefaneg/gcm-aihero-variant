package workerpool

import (
	"runtime"
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
				value, err := work(job.item)
				results[job.index] = Result[Output]{
					Value: value,
					Err:   err,
				}
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
