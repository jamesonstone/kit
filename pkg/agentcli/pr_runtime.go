package agentcli

import (
	"context"

	"github.com/jamesonstone/kit/internal/prfix"
)

type prLaneResolver interface {
	Resolve(context.Context, string, prfix.Target, prfix.PullRequest) (prfix.Lane, error)
}

type prStateStore interface {
	Acquire(prfix.Target, string) (func(), error)
	Save(prfix.Target, string, prfix.AwaitResult, []prfix.Feedback) error
}

type prFixRuntime struct {
	github  prfix.GitHub
	lane    prLaneResolver
	monitor prfix.Monitor
	state   prStateStore
}

var newPRFixRuntime = func(output prfix.Output) (prFixRuntime, error) {
	runner := prfix.ExecRunner{}
	state, err := prfix.NewStateStore()
	if err != nil {
		return prFixRuntime{}, err
	}
	return prFixRuntime{
		github: prfix.NewGitHubClient(runner), lane: prfix.NewLaneResolver(runner, output),
		monitor: prfix.NewMonitor(), state: state,
	}, nil
}
