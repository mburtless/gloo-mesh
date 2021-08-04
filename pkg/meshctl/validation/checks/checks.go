package checks

import (
	"context"
	"fmt"
	"strings"

	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation/consts"
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

// checks for all components and stages are defined here
func allChecks() map[mapKey][]Category {
	deploymentsCheck := NewDeploymentsCheck()
	serverCrdUpgradeCheck := NewCrdUpgradeCheck(consts.MgmtDeployName)
	agentCrdUpgradeCheck := NewCrdUpgradeCheck(consts.AgentDeployName)

	// check the state of the management cluster
	managementInstallation := Category{
		Name: "Gloo Mesh Management Cluster Installation",
		Checks: []Check{
			deploymentsCheck,
			NewEnterpriseRegistrationCheck(),
		},
	}

	// check the state of the remote cluster
	remoteInstallation := Category{
		Name: "Gloo Mesh Registered Cluster Installation",
		Checks: []Check{
			deploymentsCheck,
		},
	}

	agentCrds := Category{
		Name: "Gloo Mesh CRDs",
		Checks: []Check{
			agentCrdUpgradeCheck,
		},
	}

	serverCrds := Category{
		Name: "Gloo Mesh CRDs",
		Checks: []Check{
			serverCrdUpgradeCheck,
		},
	}

	agentConfig := Category{
		Name: "Agent Configuration",
		Checks: []Check{
			agentCrdUpgradeCheck,
		},
	}

	managementConfig := Category{
		Name: "Management Configuration",
		Checks: []Check{
			serverCrdUpgradeCheck,
			NewNetworkingCrdCheck(),
		},
	}

	agentParams := Category{
		Name: "Agent Parameters",
		Checks: []Check{
			NewAgentParametersCheck(),
		},
	}

	agentPreInstallConnectivity := Category{
		Name: "Relay Connectivity",
		Checks: []Check{
			NewRelayConnectivityCheck(true),
		},
	}

	agentPostInstallConnectivity := Category{
		Name: "Relay Connectivity",
		Checks: []Check{
			NewRelayConnectivityCheck(false),
		},
	}

	allchecks := map[mapKey][]Category{
		toKey(Server, PreInstall): {
			serverCrds,
		},
		toKey(Server, PostInstall): {
			managementInstallation,
			managementConfig,
		},
		toKey(Agent, PreInstall): {
			agentCrds,
			agentParams,
			agentPreInstallConnectivity,
		},
		toKey(Agent, PostInstall): {
			remoteInstallation,
			agentConfig,
			agentPostInstallConnectivity,
		},
	}

	return allchecks
}

// invoked by either meshctl or Helm hooks
// execute the checks for the given component and stage, and return true if a failure was found
func RunChecks(ctx context.Context, checkCtx CheckContext, c Component, st Stage) bool {
	if checkCtx.Context().SkipChecks {
		return false
	}
	var foundFailure bool

	for _, category := range allChecks()[toKey(c, st)] {
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
		msg.WriteString(fmt.Sprintf("\n%s %s\n", successSymbol, description))
	} else if result.IsFailure() {
		// error state
		msg.WriteString(fmt.Sprintf("\n%s %s\n", failureSymbol, description))

		for _, err := range result.Errors {
			msg.WriteString(fmt.Sprintf("  * %s\n", err.Error()))
		}

		writeHints(&msg, result.Hints)
	} else {
		// warning state
		msg.WriteString(fmt.Sprintf("\n%s %s\n", warningSymbol, description))
		writeHints(&msg, result.Hints)
	}

	fmt.Print(msg.String())
}

// mutate the msg builder with hints
func writeHints(msg *strings.Builder, hints []Hint) {
	if len(hints) > 0 {
		msg.WriteString(fmt.Sprintf("    Hints:\n"))
		for _, hint := range hints {
			msg.WriteString(fmt.Sprintf("    * %s\n", hint.Hint))
			if hint.DocsLink != nil {
				msg.WriteString(fmt.Sprintf("      For more info, see: %s\n", hint.DocsLink.String()))
			}
		}
	}
}
