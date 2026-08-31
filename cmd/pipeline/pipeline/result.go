package pipeline

import "time"

// Result is the generic outcome of any job type J.
type Result[J any] struct {
	Job      J
	Err      error
	Duration time.Duration
	// TODO: add outcome metadata (e.g. accounts processed count)
}

func (r Result[J]) succeeded() bool {
	return r.Err == nil
}
