package util

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/pmezard/go-difflib/difflib"
)

var spewConfig = spew.ConfigState{
	Indent:                  " ",
	DisablePointerAddresses: true,
	DisableCapacities:       true,
	SortKeys:                true,
	MaxDepth:                10,
}

var _ gomock.Matcher = &diffMatcher{}

type diffMatcher struct {
	expected string
	diff     string
}

// DiffEq returns a gomock matcher that prints a line by line struct diff in order to make tests that use the library
// not the most absolutely miserable experience of all time.
func DiffEq(expected interface{}) gomock.Matcher {
	return &diffMatcher{expected: spewConfig.Sdump(expected)}
}

func (m *diffMatcher) Matches(x interface{}) bool {
	actual := spewConfig.Sdump(x)
	diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(m.expected),
		B:        difflib.SplitLines(actual),
		FromFile: "Expected",
		ToFile:   "Actual",
		Context:  1,
	})
	if err != nil {
		panic(err)
	}
	m.diff = "\n" + diff

	return len(diff) == 0
}

func (m *diffMatcher) String() string {
	return m.diff
}

// MatchGoldenFile returns a gomega matcher that reads and matches a golden file.
// It uses gomega matchers for JSON and YAML and a custom diff matcher for everything else.
func MatchGoldenFile(fs fs.ReadFileFS, fname string) types.GomegaMatcher {
	b, err := fs.ReadFile(fname)
	if err != nil {
		panic(err)
	}

	switch filepath.Ext(fname) {
	case ".json":
		return gomega.MatchJSON(b)
	case ".yaml":
		return gomega.MatchYAML(b)
	default:
		return &textMatcher{fname: fname, expected: string(b)}
	}
}

type textMatcher struct {
	fname    string
	expected string
	actual   string
}

func (m *textMatcher) Match(actual interface{}) (bool, error) {
	m.actual = m.toString(actual)
	return m.actual == m.expected, nil
}

func (m *textMatcher) toString(x interface{}) string {
	switch xx := x.(type) {
	case string:
		return xx
	case []byte:
		return string(xx)
	case fmt.Stringer:
		return xx.String()
	default:
		return fmt.Sprint(xx)
	}
}

func (m *textMatcher) FailureMessage(actual interface{}) string {
	diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(m.expected),
		B:        difflib.SplitLines(m.actual),
		FromFile: "Expected",
		ToFile:   "Actual",
		Context:  1,
	})
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("Expected to match golden file.\n\nDiff:\n%s", diff)
}

func (m *textMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected not to match golden file.")
}
