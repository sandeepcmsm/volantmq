package types

import (
	"sync"
	"testing"
	"time"
)

// nolint: golint
type Provider interface {
	Name() string
	Run(t *testing.T)
}

// nolint: golint
func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
