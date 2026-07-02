package gh

import (
    "context"
    "fmt"
    "time"

    "github.com/google/go-github/v47/github"

    log "github.com/sirupsen/logrus"
)

func ReturnWorkflowRunStatus(ctx context.Context, client *github.Client, owner string, repo string, workflowRunId int) (string, *github.Timestamp, error) {

    log.WithFields(log.Fields{
        "repo":         repo,
        "owner":        owner,
        "workflowRunId": workflowRunId,
    }).Info("Calling for a previous workflow RunId...")

    var run *github.WorkflowRun
    var res *github.Response
    var err error

    // Retry transient failures (transport errors, 5xx) — same policy as
    // ReturnWorkflowRuns, and guards the nil-response dereference below.
    for attempt := 1; attempt <= maxAttempts; attempt++ {

        run, res, err = client.Actions.GetWorkflowRunByID(ctx, owner, repo, int64(workflowRunId))

        if !retryableAPIFailure(res, err) {
            break
        }

        log.WithFields(log.Fields{
            "attempt":       attempt,
            "maxAttempts":   maxAttempts,
            "repo":          repo,
            "owner":         owner,
            "workflowRunId": workflowRunId,
        }).Warn("Transient GitHub API failure fetching workflow run ...")

        if attempt < maxAttempts {
            sleepBetweenRetries(attempt)
        }
    }

    // transport error on every attempt — no HTTP response to inspect
    if res == nil {
        return "", &github.Timestamp{time.Time{}}, err
    }

    if res.StatusCode == 404 {

        log.WithFields(log.Fields{
            "repo":         repo,
            "owner":        owner,
            "workflowRunId": workflowRunId,
        }).Warn("Workflow not found ...")

        return "", &github.Timestamp{time.Time{}}, fmt.Errorf("Workflow run not found")

    }

    if res.StatusCode == 410 {

        log.WithFields(log.Fields{
            "repo":         repo,
            "owner":        owner,
            "workflowRunId": workflowRunId,
        }).Warn("received 410 code: API Method Gone...")

        return "", &github.Timestamp{time.Time{}}, fmt.Errorf("API Method Gone")
    }

    if res.StatusCode != 200 {

        log.WithFields(log.Fields{
            "Response Status":      res.StatusCode,
            "repo":         repo,
            "owner":        owner,
            "workflowRunId": workflowRunId,
        }).Warn("Request did not succeed: Response status received was not 200 ...")

        return "", &github.Timestamp{time.Time{}}, fmt.Errorf("Response status received was not 200")
    }

    if err != nil {

        return "", &github.Timestamp{time.Time{}}, err

    }

    log.WithFields(log.Fields{
        "repo":         repo,
        "owner":        owner,
        "workflowRunId": workflowRunId,
    }).Info("Workflow run was returned ...")

    return *run.Status, run.UpdatedAt, nil

}
