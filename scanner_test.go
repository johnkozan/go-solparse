package solparse

import (
	"strings"
	"testing"
)

// implementation of tests from https://github.com/ethereum/solidity/blob/develop/test/libsolidity/SolidityScanner.cpp

// Ensure the scanner can scan tokens correctly.
func TestScanner_Scan(t *testing.T) {
	type tc struct {
		f string      // func to call on parser
		e interface{} // expected result, Token or string
	}
	var tests = []struct {
		n   string
		s   string
		exp []tc
	}{
		{"test empty", "", []tc{
			tc{"currentToken", EOS},
		}},

		{"smoke test", "function break;765  \t  \"string1\",'string2'\nidentifier1", []tc{
			tc{"currentToken", Function},
			tc{"next", Break},
			tc{"next", Semicolon},
			tc{"next", Number},
			tc{"currentLiteral", "765"},
			tc{"next", StringLiteral},
			tc{"currentLiteral", "string1"},
			tc{"next", Comma},
			tc{"next", StringLiteral},
			tc{"currentLiteral", "string2"},
			tc{"next", Identifier},
			tc{"currentLiteral", "identifier1"},
			tc{"currentToken", EOS},
		}},
	}

	for _, tt := range tests {
		s := NewScanner(strings.NewReader(tt.s))

		for _, c := range tt.exp {
			switch c.f {
			case "currentToken", "next":
				tok, _ := s.Scan()
				if tok != c.e.(Token) {
					t.Errorf("%s - Expected to scan token %s got %s", tt.n, c.e, tok)
				}
			case "currentLiteral":
				lit := s.CurrentLiteral()
				if lit != c.e.(string) {
					t.Errorf("%s - Expected to scan literal %s got %s", tt.n, c.e, lit)
				}
			default:
				t.Error("invalid test func ", c.f)
			}
		}
	}
}
