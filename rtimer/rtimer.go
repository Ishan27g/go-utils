package rtimer

import (
	"math/rand"
	"sync"
	"time"
)

var timers int
var once sync.Once
var mutex sync.Mutex

func init() {
	once.Do(func() {
		timers = 0
		mutex = sync.Mutex{}
	})
}

// NewTimer closes returned channel on timeout. Use quit channel to exit timer
// without closing returned channel. Thread safe
func NewTimer(timeout time.Duration) (timedOut <-chan bool, quits chan<- bool) {
	return newTimer(timeout)
}

// NewTimerRandomDelay closes returned channel on timeout. Use quit channel to exit timer
// without closing returned channel. Thread safe
func NewTimerRandomDelay(timeout time.Duration) (timedOut <-chan bool, quits chan<- bool) {
	randomDelay := timeout + time.Duration(randomInt())
	return newTimer(randomDelay)
}

func newTimer(duration time.Duration) (chan bool, chan bool) {
	quit := make(chan bool)
	timed := make(chan bool)
	mutex.Lock()
	timers++
	mutex.Unlock()
	go func() {
		select {
		case <-quit:
			mutex.Lock()
			timers--
			mutex.Unlock()
			return
		case <-time.After(duration * time.Millisecond):
			mutex.Lock()
			timers--
			mutex.Unlock()
			close(timed)
		}
	}()
	return timed, quit
}

var randomInt = func() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(1000)
}
