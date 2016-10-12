package solparse

import (
	"strings"
	"testing"
)

// Ensure the parser can parse strings into ASTs.
func TestParser(t *testing.T) {
	var tests = []struct {
		name   string
		source string
		valid  bool
		fn     func(ContractDefinition, *testing.T)
	}{
		{
			name: "smoke_test",
			source: `contract test {
			uint256 stateVariable1;
		}`,
			valid: true,
		},
	}

	for _, tt := range tests {
		_, err := NewParser(strings.NewReader(tt.source)).Parse()
		if tt.valid && err != nil {
			t.Errorf("%s should be valid got: %s\n\n", tt.name, errstring(err))
		}
		// if has fn, call it
	}
}

// errstring returns the string representation of an error.
func errstring(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
