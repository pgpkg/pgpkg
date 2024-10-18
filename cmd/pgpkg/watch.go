package main

import (
	"fmt"
	"github.com/rjeczalik/notify"
	"path/filepath"
	"time"
)

type Watch struct {
	c chan notify.EventInfo
}

func NewWatch(pkgPath string) (*Watch, error) {
	// Make the channel buffered to ensure no event is dropped. Notify will drop
	// an event if the receiver is not able to keep up the sending pace.
	c := make(chan notify.EventInfo, 100)

	// Set up a watchpoint listening for events within a directory tree rooted
	// at current working directory. Dispatch remove events to c.
	if err := notify.Watch(pkgPath+"/...", c, notify.Create, notify.Write, notify.Rename, notify.Remove); err != nil {
		return nil, fmt.Errorf("unable to create watcher on %s: %w", pkgPath, err)
	}
	return &Watch{c: c}, nil
}

// Watch watches the filesystem for changes to ".sql" files.
func (w *Watch) Watch(then func([]notify.EventInfo)) {
	defer notify.Stop(w.c)

	delay := 100 * time.Millisecond
	var timer *time.Timer
	var events []notify.EventInfo

	for ei := range w.c {
		ext := filepath.Ext(ei.Path())
		if ext != ".sql" && ext != ".toml" {
			continue
		}

		events = append(events, ei)

		if timer != nil {
			timer.Reset(delay)
		} else {
			timer = time.AfterFunc(delay, func() {
				fmt.Print("\r")
				then(events)
				fmt.Print("pgpkg> ")
			})
		}
	}

	return
}
