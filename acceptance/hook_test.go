package acceptance

import (
	"context"

	"github.com/cucumber/godog"
)

func InjectHooks(ctx *godog.ScenarioContext) {
	ctx.Before(injectRun)
}

func injectRun(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	return context.WithValue(ctx, "run", sc.Id), nil
}
