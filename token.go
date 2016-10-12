package solparse

type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	WS

	// Literals
	IDENT

	// Misc characters
	COMMA

	// Keywords
	PRAGMA
	IMPORT
	CONTRACT
	LIBRARY
)
