package data

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewExpiry(t *testing.T) {
	for i := 1; i <= 3; i++ {
		exp := NewExpiry(30 * time.Millisecond)
		exp.Add("1", "2", "3")
		<-time.After(10 * time.Millisecond)
		exp.Reset("1")
		<-time.After(25 * time.Millisecond)
		assert.Len(t, exp.GetExpired(), 2)
		<-time.After(10 * time.Millisecond)
		assert.Len(t, exp.GetExpired(), 1)
		<-time.After(10 * time.Millisecond)
		assert.Len(t, exp.GetExpired(), 0)
	}

}
func TestConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	exp := NewExpiry(30 * time.Millisecond)
	exp.Add("1", "2", "3")
	ctx, cancel := context.WithCancel(context.Background())
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ctx.Done()
			exp.Reset("1")
		}()
	}
	cancel()
	wg.Wait()
}
