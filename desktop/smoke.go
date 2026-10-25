package main

import (
	"context"
	"sync"
	"time"
)

// smokeTimeout is how long a smoke start waits for the page to load.
const smokeTimeout = 60 * time.Second

// smokeWatch backs KIROKIT_DESKTOP_SMOKE=1, which CI uses to prove a built app
// starts: the returned function is Wails' OnDomReady and calls quit (the app
// then exits 0); when it has not run within timeout, fail runs instead.
func smokeWatch(timeout time.Duration, quit func(context.Context), fail func()) func(context.Context) {
	var once sync.Once
	loaded := make(chan struct{})
	go func() {
		select {
		case <-loaded:
		case <-time.After(timeout):
			fail()
		}
	}()
	return func(ctx context.Context) {
		once.Do(func() {
			close(loaded)
			quit(ctx)
		})
	}
}
