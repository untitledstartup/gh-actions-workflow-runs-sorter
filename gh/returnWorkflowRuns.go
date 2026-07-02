package gh

import (
    "context"
    "fmt"

    "github.com/google/go-github/v47/github"

    log "github.com/sirupsen/logrus"
)

func ReturnWorkflowRuns(branchName string, ctx context.Context, client *github.Client, owner string, repo string, workflowFile string, workflowRunsToReturn int) ([]*github.WorkflowRun, error) {

    log.WithFields(log.Fields{
        "repo":         repo,
        "owner":        owner,
        "workflowFile": workflowFile,
        "workflowRunsToReturn": workflowRunsToReturn,
    }).Info("Calling for last few runs from workflow...")

    opts := &github.ListWorkflowRunsOptions{
        Branch: branchName,
        ListOptions: github.ListOptions{
            Page: 1,
            PerPage: workflowRunsToReturn,
        },
    }

    var runs *github.WorkflowRuns
    var res *github.Response
    var err error

    // Retry transient failures (transport errors, 5xx) — a single GitHub API
    // blip here otherwise aborts the whole shouldExecute decision.
    for attempt := 1; attempt <= maxAttempts; attempt++ {

        runs, res, err = client.Actions.ListWorkflowRunsByFileName(ctx, owner, repo, workflowFile, opts)

        if !retryableAPIFailure(res, err) {
            break
        }

        log.WithFields(log.Fields{
            "attempt":      attempt,
            "maxAttempts":  maxAttempts,
            "repo":         repo,
            "owner":        owner,
            "workflowFile": workflowFile,
        }).Warn("Transient GitHub API failure listing workflow runs ...")

        if attempt < maxAttempts {
            sleepBetweenRetries(attempt)
        }
    }

    // transport error on every attempt — no HTTP response to inspect
    if res == nil {
        return nil, err
    }

    if res.StatusCode == 404 {

        log.WithFields(log.Fields{
            "Response Status":      res.StatusCode,
            "repo":                 repo,
            "owner":                owner,
            "workflowFile":         workflowFile,
            "workflowRunsToReturn": workflowRunsToReturn,
        }).Warn("Workflow not found ...")

        return nil, fmt.Errorf("Workflow not found")

    }

    if res.StatusCode == 410 {

        log.WithFields(log.Fields{
            "Response Status":      res.StatusCode,
            "repo":                 repo,
            "owner":                owner,
            "workflowFile":         workflowFile,
            "workflowRunsToReturn": workflowRunsToReturn,
        }).Warn("received 410 code: API Method Gone...")

        return nil, fmt.Errorf("API Method Gone")
    }

    if res.StatusCode != 200 {

        log.WithFields(log.Fields{
            "Response Status":      res.StatusCode,
            "repo":                 repo,
            "owner":                owner,
            "workflowFile":         workflowFile,
            "workflowRunsToReturn": workflowRunsToReturn,
        }).Warn("Request did not succeed: Response status received was not 200 ...")

        return nil, fmt.Errorf("Response status received was not 200")
    }

    if err != nil {

        return nil, err
    }

    log.WithFields(log.Fields{
        "repo":         repo,
        "owner":        owner,
        "workflowFile": workflowFile,
        "workflowRunsToReturn": workflowRunsToReturn,
    }).Info("Runs were returned ...")

    return runs.WorkflowRuns, nil

}
