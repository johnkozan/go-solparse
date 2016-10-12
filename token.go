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
	LBRACE
	RBRACE
	SEMICOLON

	// Keywords
	PRAGMA
	IMPORT
	CONTRACT
	LIBRARY

	FUNCTION
	STRUCT
	ENUM
	MAPPING
	ELEM
	MODIFIER
	EVENT
	USING
)
