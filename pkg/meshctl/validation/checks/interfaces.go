package checks

import (
	"context"
	"net/url"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Stage string

const (
	PreInstall  Stage = "pre-install"
	PostInstall       = "post-install"
	PreUpgrade        = "pre-upgrade"
	PostUpgrade       = "post-upgrade"
	Test              = "test"
)

type Component string

const (
	Server Component = "server"
	Agent            = "agent"
)

type Environment struct {
	// Gloo Mesh management plane metrics port
	AdminPort uint32
	// the Gloo Mesh installation namespace
	Namespace string
	// true if this validation check is being run from an in-cluster workload
	InCluster bool
}

type OperateOnAdminPort = func(ctx context.Context, adminUrl *url.URL) (error, string)

type CommonContext struct {
	// if true, skip checks
	SkipChecks bool

	Env Environment
	Cli client.Client
	// server install or upgrade parameters
	ServerParams *ServerParams
}

func (c *CommonContext) Environment() Environment {
	return c.Env
}
func (c *CommonContext) Client() client.Client {
	return c.Cli
}

type CheckContext interface {
	Environment() Environment
	Client() client.Client
	AccessAdminPort(ctx context.Context, deployment string, op OperateOnAdminPort) (error, string)
	Context() CommonContext

	// execute the checks for the given component and stage, and return true if all checks passed
	RunChecks(ctx context.Context, c Component, st Stage) bool
}

type Check interface {
	// description of what is being checked
	GetDescription() string

	// Execute the check, pass in the namespace that Gloo Mesh is installed in
	Run(ctx context.Context, checkCtx CheckContext) *Result
}

type Category struct {
	Name   string
	Checks []Check
}

type Result struct {
	// user-facing error message describing failed check
	Errors []error

	// an optional suggestion for a next action for the user to take for resolving a failed check or warning
	Hints []Hint
}

func (f *Result) AddError(err ...error) *Result {
	f.Errors = append(f.Errors, err...)
	return f
}

// result is a failure if there are any errors
func (f *Result) IsFailure() bool {
	if f == nil {
		return false
	}
	return len(f.Errors) > 0
}

// result is a warning if there are no errors but there are hints
func (f *Result) IsWarning() bool {
	if f == nil {
		return false
	}
	return len(f.Errors) == 0 && len(f.Hints) > 0
}

// result is a success if nil or no errors and no hints
func (f *Result) IsSuccess() bool {
	if f == nil {
		return true
	}
	return len(f.Errors) == 0 && len(f.Hints) == 0
}

// add a hint with an optional docs link
func (f *Result) AddHint(hint string, docsLink string) *Result {
	if hint != "" {
		var u *url.URL
		if docsLink != "" {
			var err error
			u, err = url.Parse(docsLink)
			if err != nil {
				// this should never happen
				// but we also don't care that much if it does
				// so we just ignore the error.
			}
		}
		f.Hints = append(f.Hints, Hint{Hint: hint, DocsLink: u})
	}
	return f
}

type Hint struct {
	// an optional suggestion for a next action for the user to take for resolving a failed check
	Hint string
	// optionally provide a link to a docs page that a user should consult to resolve the error
	DocsLink *url.URL
}
