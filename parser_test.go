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
			name: "smoke test",
			source: `contract test {
			uint256 stateVariable1;
		}`,
			valid: true,
		},
		{
			name: "missing variable name in declaration",
			source: `contract test {
			uint256 ;
		}`,
			valid: false,
		},
		{
			name: "empty function",
			source: `contract test {
			uint256 stateVar;
			function functionName(bytes20 arg1, address addr) constant
			  returns (int id)
			  { }
		}`,
			valid: true,
		},
	}

	for _, tt := range tests {
		_, err := NewParser(strings.NewReader(tt.source)).Parse()
		if tt.valid && err != nil {
			t.Errorf("%s should be valid got: %s\n\n", tt.name, errstring(err))
		}
		if !tt.valid && err == nil {
			t.Errorf("%s should not be valid.  Parsed as valid", tt.name)
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
