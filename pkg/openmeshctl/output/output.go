package output

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Context is the context required to perform output operations.
type Context interface {
	// Out returns the writer to write normal output to.
	Out() io.Writer

	// ErrOut returns the writer to write error output to.
	ErrOut() io.Writer

	// Verbose returns whether or not print debug messages.
	Verbose() bool
}

// Format is a type of output format.
type Format string

const (
	Default Format = ""
	Wide    Format = "wide"
	JSON    Format = "json"
	YAML    Format = "yaml"
)

// AddFormatFlag adds the format flag to the given receiver.
func AddFormatFlag(flags *pflag.FlagSet, format *string) {
	opts := strings.Join([]string{string(Default), string(Wide), string(JSON), string(YAML)}, "|")
	flags.StringVarP(format, "output", "o", string(Default), "Resource output format. One of: "+opts)
}

// RefToString converts a reference to a string representation.
func RefToString(ref ezkube.ResourceId) string {
	if ref.GetName() == "" {
		return ""
	}
	var sb strings.Builder
	if clusterRef, ok := ref.(ezkube.ClusterResourceId); ok && clusterRef.GetClusterName() != "" {
		sb.WriteString(clusterRef.GetClusterName() + ".")
	}
	sb.WriteString(ref.GetNamespace() + "." + ref.GetName())
	return sb.String()
}

// FormatAge returns the age of an object formatted as days.
func FormatAge(created metav1.Time) string {
	diff := time.Now().Sub(created.Time)
	switch {
	case diff.Seconds() < 120:
		return fmt.Sprintf("%ds", int(diff.Seconds()))
	case diff.Minutes() < 10:
		return fmt.Sprintf("%dm%ds", int(diff.Minutes()), int(diff.Seconds()))
	case diff.Minutes() < 60:
		return fmt.Sprintf("%dm", int(diff.Minutes()))
	case diff.Hours() < 10:
		return fmt.Sprintf("%dh%dm", int(diff.Hours()), int(diff.Minutes()))
	case diff.Hours() < 24:
		return fmt.Sprintf("%dh", int(diff.Hours()))
	case diff.Hours() < 240: // 10 days
		return fmt.Sprintf("%dd%dh", int(diff.Hours())%24, int(diff.Hours()))
	default:
		return fmt.Sprintf("%dd", int(diff.Hours())%24)
	}
}

// DebugPrint prints out a debug message if verbose is enabled.
func DebugPrint(ctx Context, format string, a ...interface{}) {
	if ctx.Verbose() {
		fmt.Fprintf(ctx.ErrOut(), format, a...)
	}
}

// MakeSpinner returns a spinner to use with the given message.
// It returns the spinner as well as a function to cancel the spinner and have
// it add an error label. It can be safely deferred since if its not active, the
// cancel will be a no-op.
func MakeSpinner(ctx Context, format string, a ...interface{}) (*spinner.Spinner, func()) {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Writer = ctx.ErrOut() // don't pollute std out with the spinner
	s.Suffix = fmt.Sprintf(format, a...)
	s.FinalMSG = "DONE\n"
	return s, func() {
		if s.Active() {
			s.FinalMSG = "ERROR\n"
			s.Stop()
		}
	}
}
