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
			{"currentToken", EOS},
		}},

		{"smoke test", "function break;765  \t  \"string1\",'string2'\nidentifier1", []tc{
			{"currentToken", Function},
			{"next", Break},
			{"next", Semicolon},
			{"next", Number},
			{"currentLiteral", "765"},
			{"next", StringLiteral},
			{"currentLiteral", "string1"},
			{"next", Comma},
			{"next", StringLiteral},
			{"currentLiteral", "string2"},
			{"next", Identifier},
			{"currentLiteral", "identifier1"},
			{"next", EOS},
		}},

		{"string escapes", "  { \"a\\x61\"", []tc{
			{"currentToken", LBrace},
			{"next", StringLiteral},
			{"currentLiteral", "aa"},
		}},

		{"string escapes with zeros", "  { \"a\\x61\\x00abc\"", []tc{
			{"currentToken", LBrace},
			{"next", StringLiteral},
			{"currentLiteral", "aa\000abc"},
		}},

		{"string escape illegal", " bla \"\\x6rf\" (illegalescape)", []tc{
			{"currentToken", Identifier},
			{"next", Illegal},
			{"currentLiteral", ""},
			// TODO: does illegal handling match offical implementation?
			//{"next", Illegal},
			{"next", Identifier},
			{"next", Illegal},
			{"next", EOS},
		}},

		{"hex numbers", "var x = 0x765432536763762734623472346;", []tc{
			{"currentToken", Var},
			{"next", Identifier},
			{"next", Assign},
			{"next", Number},
			{"currentLiteral", "0x765432536763762734623472346"},
			{"next", Semicolon},
			{"next", EOS},
		}},
	}

	for _, tt := range tests {
		s := NewScanner(strings.NewReader(tt.s))

		for k, c := range tt.exp {
			switch c.f {
			case "currentToken":
				tok := s.currentToken()
				if tok != c.e.(Token) {
					t.Errorf("%s , case: %d -- Expected current token '%s' got '%s'", tt.n, k, c.e, tok)
				}
			case "next":
				tok := s.next()
				if tok != c.e.(Token) {
					t.Errorf("%s , case: %d -- Expected next token '%s' got '%s' - literal '%s'", tt.n, k, c.e, tok, s.currentLiteral())
				}
			case "currentLiteral":
				lit := s.currentLiteral()
				if lit != c.e.(string) {
					t.Errorf("%s , case %d  -- Expected current literal '%s' got '%s'", tt.n, k, c.e, lit)
				}
			default:
				t.Error("invalid test func ", c.f)
			}
		}
	}
}
