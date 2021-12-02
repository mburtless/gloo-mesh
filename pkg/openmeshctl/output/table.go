package output

// Table is used to represent an object in a tabular manner.
type Table struct {
	Headers []string
	NextRow func() []string

	nextRowData []string
}

// HasNextRow advances the table to the next row.
func (t *Table) HasNextRow() bool {
	row := t.NextRow()
	if row == nil {
		return false
	}

	t.nextRowData = row
	return true
}

// GetNextRow returns the current data row.
func (t *Table) GetNextRow() []string {
	return t.nextRowData
}
