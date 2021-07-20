package checks_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation/checks"
)

var _ = Describe("Checks", func() {

	var (
		ctx = context.TODO()
	)

	It("should verify relay server addresses of different forms", func() {
		ipv4 := "192.0.2.146"
		ipv6 := "2001:db8:1f70::999:de8:7648:6e8"
		ipv4WithPort := "192.0.2.146:9090"
		ipv6WithPort := "[2001:db8:1f70::999:de8:7648:6e8]:9090"
		dnsname := "relay-server"
		dnsNameWithPort := "relay-server:9090"
		invalidDnsName := "asd#$%"
		invalidIpv4WithScheme := "http://192.0.2.146"
		invalidIpv6 := "[2001:db8:1f70::999:de8:7648:6e8]"

		for _, validAddress := range []string{
			ipv4WithPort,
			ipv4,
			ipv6,
			dnsname,
			ipv6WithPort,
			dnsNameWithPort,
		} {
			checkCtx, err := buildCheckContext(validAddress)
			Expect(err).To(BeNil())
			Expect(checkCtx.RunChecks(ctx, checks.Server, checks.PreInstall)).To(BeFalse())
		}

		for _, invalidAddress := range []string{
			invalidDnsName,
			invalidIpv4WithScheme,
			invalidIpv6,
		} {
			checkCtx, err := buildCheckContext(invalidAddress)
			Expect(err).To(BeNil())
			runChecks := checkCtx.RunChecks(ctx, checks.Server, checks.PreInstall)
			Expect(runChecks).To(BeTrue())
		}
	})
})

func buildCheckContext(relayServerAddress string) (checks.CheckContext, error) {
	return validation.NewTestCheckContext(
		nil,
		"",
		0,
		0,
		&checks.ServerParams{
			RelayServerAddress: relayServerAddress,
		},
		false,
	)
}
