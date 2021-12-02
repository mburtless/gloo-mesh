package output

import (
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Summary holds human readable information about a resource.
type Summary struct {
	// Meta for the object the summary is for.
	Meta metav1.Object

	// Fields additional fields to print for the object beyond the meta
	Fields FieldSet
}

// FieldSet is an ordered set of fields
type FieldSet []Field

// AddField adds a field to the set.
func (fs *FieldSet) AddField(label string, value interface{}) {
	*fs = append(*fs, Field{Label: label, Value: value})
}

// Sort sorts the fields in the event that they are unordered to enforce order.
// For example, if a field set is built from a map, Sort can be used to add a guaranteed order.
func (fs *FieldSet) Sort() {
	sort.Slice(*fs, func(i, j int) bool {
		return strings.Compare((*fs)[i].Label, (*fs)[j].Label) < 0
	})
}

// Field is a field that can be added to the summary
type Field struct {
	Label string
	Value interface{}
}
