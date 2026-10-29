package main

import (
	"sync"
	"time"
)

// smokeTimeout is how long a smoke start waits for the page to answer.
const smokeTimeout = 60 * time.Second

// smokeWatch backs KIROKIT_DESKTOP_SMOKE=1, which CI uses to prove a built app
// starts: main.go calls the returned function when the page sends smoke:ready
// (after its first State call answered), and it calls quit (the app then exits
// 0); when it has not run within timeout, fail runs instead.
func smokeWatch(timeout time.Duration, quit func(), fail func()) func() {
	var once sync.Once
	loaded := make(chan struct{})
	go func() {
		select {
		case <-loaded:
		case <-time.After(timeout):
			fail()
		}
	}()
	return func() {
		once.Do(func() {
			close(loaded)
			quit()
		})
	}
}
