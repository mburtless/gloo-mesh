package main

import (
	"fmt"
	"os"

	// required import to enable kube client-go auth plugins
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	if err := commands.RootCommand().Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
