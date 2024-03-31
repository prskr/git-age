package testx

import (
	"context"
	"time"
)

type DeadlineHelper interface {
	Helper()
	Deadline() (deadline time.Time, ok bool)
	Cleanup(f func())
}

func Context(t DeadlineHelper) context.Context {
	t.Helper()

	deadline, hasDeadline := t.Deadline()
	if hasDeadline {
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		t.Cleanup(cancel)

		return ctx
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	return ctx
}
