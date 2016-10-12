package solparse

import (
	"reflect"
	"strings"
	"testing"
)

// Ensure the parser can parse strings into ASTs.
func TestParser_ParseContract(t *testing.T) {
	var tests = []struct {
		s    string
		stmt *ContractDefinition
		err  string
	}{
		// Empty contract
		{
			s: `contract TestContract {}`,
			stmt: &ContractDefinition{
				Name:      "TestContract",
				IsLibrary: false,
			},
		},

		// Errors
		{s: `foo`, err: `Expected import directive or contract definition`},
	}

	for i, tt := range tests {
		stmt, err := NewParser(strings.NewReader(tt.s)).Parse()
		if !reflect.DeepEqual(tt.err, errstring(err)) {
			t.Errorf("%d. %q: error mismatch:\n  exp=%s\n  got=%s\n\n", i, tt.s, tt.err, err)
		} else if tt.err == "" && !reflect.DeepEqual(tt.stmt, stmt) {
			t.Errorf("%d. %q\n\nstmt mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.s, tt.stmt, stmt)
		}
	}
}

// errstring returns the string representation of an error.
func errstring(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
