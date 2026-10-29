package main

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestSmoke_PageReadyQuitsWithSuccess(t *testing.T) {
	var quits, fails atomic.Int32
	ready := smokeWatch(time.Second, func() { quits.Add(1) }, func() { fails.Add(1) })

	ready()
	ready() // a second smoke:ready (a page reload) quits only once
	time.Sleep(1200 * time.Millisecond)

	if quits.Load() != 1 || fails.Load() != 0 {
		t.Errorf("quits = %d, fails = %d; want one quit and no failure", quits.Load(), fails.Load())
	}
}

func TestSmoke_NoPageWithinTheLimitFails(t *testing.T) {
	failed := make(chan struct{})
	var quits atomic.Int32
	smokeWatch(50*time.Millisecond, func() { quits.Add(1) }, func() { close(failed) })

	select {
	case <-failed:
	case <-time.After(5 * time.Second):
		t.Fatal("no failure after the limit")
	}
	if quits.Load() != 0 {
		t.Error("quit called although the page never answered")
	}
}
