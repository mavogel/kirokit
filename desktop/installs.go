package main

import (
	"sync"
	"time"

	"github.com/mavogel/kirokit/internal/install"
)

// installsPollInterval is how often the install records are read.
const installsPollInterval = time.Second

// installPoller reads the install records under the Kiro home and emits
// installs:changed when they differ from the last read.
type installPoller struct {
	home func() string
	emit func(event string, data any)

	mu   sync.Mutex // guards last
	last []install.Record

	stop     chan struct{}
	done     chan struct{}
	stopOnce sync.Once
}

// newInstallPoller takes the first reading as the baseline, so the pages'
// own first load is not announced again, and starts checking every interval.
// home is asked at each check, so a Settings change applies to the next one.
func newInstallPoller(home func() string, emit func(event string, data any), interval time.Duration) *installPoller {
	p := &installPoller{home: home, emit: emit, stop: make(chan struct{}), done: make(chan struct{})}
	p.last, _ = install.Installed(home())
	go p.run(interval)
	return p
}

func (p *installPoller) run(interval time.Duration) {
	defer close(p.done)
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-p.stop:
			return
		case <-t.C:
			p.Check()
		}
	}
}

// Check reads the records now and emits installs:changed when they differ
// from the previous read. A failed read keeps the previous state.
func (p *installPoller) Check() {
	records, err := install.Installed(p.home())
	if err != nil {
		return
	}
	p.mu.Lock()
	changed := !sameRecords(p.last, records)
	p.last = records
	p.mu.Unlock()
	if changed {
		if records == nil {
			records = []install.Record{}
		}
		p.emit("installs:changed", records)
	}
}

// Stop ends the polling goroutine and waits for it. It is safe to call twice.
func (p *installPoller) Stop() {
	p.stopOnce.Do(func() { close(p.stop) })
	<-p.done
}

func sameRecords(a, b []install.Record) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Team != b[i].Team || a[i].Files != b[i].Files || !a[i].Modified.Equal(b[i].Modified) {
			return false
		}
	}
	return true
}

func (a *App) startInstallPoller() {
	p := newInstallPoller(a.kiroHome, a.emit, installsPollInterval)
	a.mu.Lock()
	a.poller = p
	a.mu.Unlock()
}

// Installs lists the teams installed under the Kiro home. The pages call it
// once at start; installs:changed keeps them current after that.
func (a *App) Installs() ([]install.Record, error) {
	records, err := install.Installed(a.kiroHome())
	if records == nil {
		records = []install.Record{}
	}
	return records, err
}
