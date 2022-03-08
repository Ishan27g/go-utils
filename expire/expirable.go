package data

import (
	"context"
	"sync"
	"time"
)

var ExpiryTime = 2 * time.Second

type Expiry interface {
	Add(ids ...string) // Add ids that will expire. Re-adding same id before it expires resets the timer
	Reset(ids ...string)
	GetExpired() []string // GetExpired returns ids that have expired
	Check(id string) bool // check if id is present and
}

type expiry struct {
	lock  sync.Mutex
	ids   map[string]*time.Time
	reset map[string]*struct {
		ctx context.Context
		can context.CancelFunc
	}
	expired []string
}

// Check returns true if id is added but not yet expired
func (e *expiry) Check(id string) bool {
	return e.ids[id] != nil
}

// Add an id that will get appended to expired slice after no further calls to add/reset
func (e *expiry) Add(ids ...string) {
	e.lock.Lock()
	defer e.lock.Unlock()
	for _, id := range ids {
		e.add(id)
	}
}

// GetExpired returns ids that have expired
func (e *expiry) GetExpired() []string {
	e.lock.Lock()
	defer e.lock.Unlock()
	expired := e.expired
	e.expired = []string{}
	return expired
}
func (e *expiry) expire(id string) {
	ctx, _ := context.WithCancel(e.reset[id].ctx)
	go func() {
		select {
		case <-time.After(ExpiryTime):
			e.lock.Lock()
			e.expired = append(e.expired, id)
			delete(e.ids, id)
			e.lock.Unlock()
		case <-ctx.Done():
			return
		}
	}()

}
func (e *expiry) add(id string) {
	if e.reset[id] != nil {
		e.reset[id].can()
	}
	ctx, can := context.WithCancel(context.Background())
	e.reset[id] = &struct {
		ctx context.Context
		can context.CancelFunc
	}{ctx: ctx, can: can}
	now := time.Now()
	e.ids[id] = &now
	e.expire(id)
}

// Reset or add
func (e *expiry) Reset(ids ...string) {
	for _, id := range ids {
		if e.ids[id] == nil {
			e.lock.Lock()
			e.add(id)
			e.lock.Unlock()
		} else {
			e.reset[id].can()
			e.lock.Lock()
			e.add(id)
			e.lock.Unlock()
		}
	}
}

func NewExpiry(withTimeout time.Duration) Expiry {
	ExpiryTime = withTimeout
	return &expiry{
		lock: sync.Mutex{},
		ids:  make(map[string]*time.Time),
		reset: make(map[string]*struct {
			ctx context.Context
			can context.CancelFunc
		}),
		expired: []string{},
	}
}
