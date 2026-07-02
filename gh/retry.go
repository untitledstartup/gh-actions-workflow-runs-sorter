package gh

import (
    "time"

    "github.com/google/go-github/v47/github"
)

const maxAttempts = 3

// sleepBetweenRetries is a variable so tests can eliminate the backoff delay.
var sleepBetweenRetries = func(attempt int) {
    time.Sleep(time.Duration(attempt) * 5 * time.Second)
}

// retryableAPIFailure reports whether a GitHub API call failed in a way worth
// retrying: a transport error (no HTTP response at all) or a 5xx from the API.
// 4xx responses (404, 410, auth) are permanent and must not be retried.
func retryableAPIFailure(res *github.Response, err error) bool {

    if res == nil {
        return err != nil
    }

    return res.StatusCode >= 500
}
