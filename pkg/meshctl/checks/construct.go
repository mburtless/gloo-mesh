package checks

import (
	"context"
	"fmt"
	"strings"
)

var (
	successSymbol = "🟢"
	warningSymbol = "🟡"
	failureSymbol = "🔴"
)

type mapKey struct {
	Stage     Stage
	Component Component
}

func toKey(Component Component, Stage Stage) mapKey {
	return mapKey{Stage: Stage, Component: Component}
}

func constructChecks() map[mapKey][]Category {
	managementPlane := Category{
		Name: "Gloo Mesh Installation",
		Checks: []Check{
			NewDeploymentsCheck(),
			NewEnterpriseRegistrationCheck(),
		},
	}

	configuration := Category{
		Name: "Management Configuration",
		Checks: []Check{
			NewNetworkingCrdCheck(),
		},
	}

	return map[mapKey][]Category{
		toKey(Server, PostInstall): {
			managementPlane,
			configuration,
		},
	}
}

func RunChecks(ctx context.Context, checkCtx CheckContext, c Component, st Stage) bool {
	var foundFailure bool

	for _, category := range constructChecks()[toKey(c, st)] {
		fmt.Println(category.Name)
		fmt.Printf(strings.Repeat("-", len(category.Name)+3) + "\n")
		for _, check := range category.Checks {
			result := check.Run(ctx, checkCtx)

			if result.IsFailure() {
				foundFailure = true
			}

			printResult(result, check.GetDescription())
		}
		fmt.Println()
	}

	return foundFailure
}

func printResult(result *Result, description string) {
	var msg strings.Builder

	if result.IsSuccess() {
		// success state
		msg.WriteString(fmt.Sprintf("%s %s\n\n", successSymbol, description))
	} else if result.IsFailure() {
		// error state
		msg.WriteString(fmt.Sprintf("%s %s\n", failureSymbol, description))

		for _, err := range result.Errors {
			msg.WriteString(fmt.Sprintf("  * %s\n", err.Error()))
		}

		writeHints(&msg, result.Hints)
	} else {
		// warning state
		msg.WriteString(fmt.Sprintf("%s %s\n", warningSymbol, description))
		writeHints(&msg, result.Hints)
	}

	fmt.Print(msg.String())
}

// mutate the msg builder with hints
func writeHints(msg *strings.Builder, hints []Hint) {
	if len(hints) > 0 {
		msg.WriteString(fmt.Sprintf("    Hints:\n"))
		for _, hint := range hints {
			msg.WriteString(fmt.Sprintf("    * %s", hint.Hint))
			if hint.DocsLink != nil {
				msg.WriteString(fmt.Sprintf(" (For more info, see: %s)\n", hint.DocsLink.String()))
			}
		}
	}
}
