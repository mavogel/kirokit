package main

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestSmoke_PageLoadedQuitsWithSuccess(t *testing.T) {
	var quits, fails atomic.Int32
	ready := smokeWatch(time.Second, func(context.Context) { quits.Add(1) }, func() { fails.Add(1) })

	ready(context.Background())
	ready(context.Background()) // a second DOM-ready quits only once
	time.Sleep(1200 * time.Millisecond)

	if quits.Load() != 1 || fails.Load() != 0 {
		t.Errorf("quits = %d, fails = %d; want one quit and no failure", quits.Load(), fails.Load())
	}
}

func TestSmoke_NoPageWithinTheLimitFails(t *testing.T) {
	failed := make(chan struct{})
	var quits atomic.Int32
	smokeWatch(50*time.Millisecond, func(context.Context) { quits.Add(1) }, func() { close(failed) })

	select {
	case <-failed:
	case <-time.After(5 * time.Second):
		t.Fatal("no failure after the limit")
	}
	if quits.Load() != 0 {
		t.Error("quit called although the page never loaded")
	}
}
