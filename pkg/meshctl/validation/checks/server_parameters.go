package checks

import (
	"context"

	"github.com/asaskevich/govalidator"
	"github.com/rotisserie/eris"
)

type ServerParams struct {
	RelayServerAddress string
}

type parameterCheck struct{}

func NewServerParametersCheck() *parameterCheck {
	return &parameterCheck{}
}

func (p *parameterCheck) GetDescription() string {
	return "Gloo Mesh server parameters are valid"
}

func (p *parameterCheck) Run(_ context.Context, checkCtx CheckContext) *Result {
	serverParams := checkCtx.Context().ServerParams

	result := &Result{}

	// update result with all server paramater checks
	isValidRelayServerAddress(serverParams.RelayServerAddress, result)

	return result
}

// return true is relay server address exists and is a valid address (either hostname or IP, with optional port)
// validate according to the gRPC spec: https://github.com/grpc/grpc/blob/master/doc/naming.md
func isValidRelayServerAddress(relayServerAddress string, result *Result) {
	if relayServerAddress == "" {
		result.Errors = append(result.Errors, eris.Errorf("relay server address cannot be empty"))
		return
	}

	// relay server address must be one of the following:
	//  * a dial string (contains both host and port)
	//  * IP address (either ipv4 or ipv6) without port
	//  * DNS name without port
	if !(govalidator.IsDialString(relayServerAddress) ||
		govalidator.IsIP(relayServerAddress) ||
		govalidator.IsDNSName(relayServerAddress)) {
		result.Errors = append(result.Errors, eris.Errorf("relay server address is not valid: %s", relayServerAddress))
	}
}
