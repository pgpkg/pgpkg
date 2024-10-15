package main

import (
	"fmt"
	"github.com/rjeczalik/notify"
)

func watch(pkgPath string, then func(notify.EventInfo)) error {
	// Make the channel buffered to ensure no event is dropped. Notify will drop
	// an event if the receiver is not able to keep up the sending pace.
	c := make(chan notify.EventInfo, 100)

	// Set up a watchpoint listening for events within a directory tree rooted
	// at current working directory. Dispatch remove events to c.
	if err := notify.Watch(pkgPath+"/...", c, notify.Create, notify.Write, notify.Rename, notify.Remove); err != nil {
		return fmt.Errorf("unable to create watcher on %s: %w", pkgPath, err)
	}
	defer notify.Stop(c)

	// Block until an event is received.
	for ei := range c {
		then(ei)
	}

	return nil
}
