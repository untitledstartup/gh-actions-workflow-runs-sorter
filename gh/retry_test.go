package gh

import (
    "context"
    "errors"
    "fmt"
    "io/ioutil"
    "net/http"
    "testing"

    "github.com/google/go-github/v47/github"
    log "github.com/sirupsen/logrus"
)

func TestRetryableAPIFailure(t *testing.T) {

    tests := []struct {
        name string
        res  *github.Response
        err  error
        want bool
    }{
        {"transport error, no response", nil, fmt.Errorf("connection reset by peer"), true},
        {"5xx response", &github.Response{Response: &http.Response{StatusCode: 503}}, nil, true},
        {"4xx response", &github.Response{Response: &http.Response{StatusCode: 404}}, nil, false},
        {"200 response", &github.Response{Response: &http.Response{StatusCode: 200}}, nil, false},
        {"bare context canceled", nil, context.Canceled, false},
        {"bare deadline exceeded", nil, context.DeadlineExceeded, false},
        {"wrapped context canceled", nil, fmt.Errorf("Get \"...\": %w", context.Canceled), false},
        {"wrapped deadline exceeded", nil, fmt.Errorf("Get \"...\": %w", context.DeadlineExceeded), false},
        {"no response, no error", nil, nil, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := retryableAPIFailure(tt.res, tt.err); got != tt.want {
                t.Errorf("retryableAPIFailure() = %v, want %v", got, tt.want)
            }
        })
    }
}

// A canceled context must fail fast: the wrapper should not classify the
// resulting transport error as retryable, so it never enters the backoff.
func TestReturnWorkflowRunsDoesNotRetryOnCanceledContext(t *testing.T) {

    // supress logrus
    log.SetOutput(ioutil.Discard)

    sleeps := 0
    sleepBetweenRetries = func(context.Context, int) { sleeps++ }

    client, mux, _, teardown := Setup()
    defer teardown()

    mux.HandleFunc("/repos/testowner/testrepo/actions/workflows/testfile.yaml/runs", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprint(w, `{"total_count":0,"workflow_runs":[]}`)
    })

    ctx, cancel := context.WithCancel(context.Background())
    cancel()

    _, gotErr := ReturnWorkflowRuns("ft/test-branch", ctx, client, "testowner", "testrepo", "testfile.yaml", 20)

    if gotErr == nil {
        t.Fatalf("expected an error on a canceled context but got nil")
    }

    if !errors.Is(gotErr, context.Canceled) {
        t.Errorf("expected context.Canceled but got '%v'", gotErr)
    }

    if sleeps != 0 {
        t.Errorf("expected no backoff sleeps on a canceled context but got %d", sleeps)
    }
}
