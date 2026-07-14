package gh

import (
    "context"
    "errors"
    "time"

    "github.com/google/go-github/v47/github"
)

const maxAttempts = 3

// sleepBetweenRetries waits attempt*5s before the next retry, returning early
// if ctx is canceled or its deadline passes so a canceled call does not block
// through the backoff. It is a variable so tests can eliminate the delay.
var sleepBetweenRetries = func(ctx context.Context, attempt int) {

    timer := time.NewTimer(time.Duration(attempt) * 5 * time.Second)
    defer timer.Stop()

    select {
    case <-timer.C:
    case <-ctx.Done():
    }
}

// retryableAPIFailure reports whether a GitHub API call failed in a way worth
// retrying: a transport error (no HTTP response at all) or a 5xx from the API.
// 4xx responses (404, 410, auth) are permanent, and context cancellation or
// deadline errors are terminal — neither is ever retried.
func retryableAPIFailure(res *github.Response, err error) bool {

    if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
        return false
    }

    if res == nil {
        return err != nil
    }

    return res.StatusCode >= 500
}
