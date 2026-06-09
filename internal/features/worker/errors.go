package core_worker

import "errors"

var (
	ErrFatalDBAction     = errors.New("fatal db action")
	ErrRetryableDBAction = errors.New("retryable db action")
)
