package concurrent

import (
	"iter"
	"sync"

	"github.com/sdsc-ordes/quitsh/pkg/errors"
)

// Map runs over all items in parallel and collects errors.
func Map[T any](items iter.Seq[T], run func(p T) error) error {
	wg := sync.WaitGroup{}
	errChan := make(chan error)

	// Process in parallel.
	for p := range items {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errChan <- run(p)
		}()
	}

	// Wait on a separate Go routine, for finishing.
	// The close the channel, that the below code, finishes.
	go func() {
		wg.Wait()
		close(errChan)
	}()

	var err error
	for e := range errChan {
		if e != nil {
			err = errors.Combine(err, e)
		}
	}

	return err
}

// Map runs over all items in parallel and collects errors.
func Map2[K any, T any](items iter.Seq2[K, T], run func(i K, p T) error) error {
	wg := sync.WaitGroup{}
	errChan := make(chan error)

	// Process in parallel.
	for i, p := range items {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errChan <- run(i, p)
		}()
	}

	// Wait on a separate Go routine, for finishing.
	// The close the channel, that the below code, finishes.
	go func() {
		wg.Wait()
		close(errChan)
	}()

	var err error
	for e := range errChan {
		if e != nil {
			err = errors.Combine(err, e)
		}
	}

	return err
}
