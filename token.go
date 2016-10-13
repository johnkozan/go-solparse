package solparse

type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	WS

	// Literals
	IDENT
	STRINGLIT
	NUMBER

	// Misc characters
	COMMA
	LBRACE
	RBRACE
	SEMICOLON
	LPAREN
	RPAREN

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
	VAR
	BREAK
)

var tokLit = []string{"ILLEGAL", "EOF", "<whitespace>", "Identifier", "String literal", "Number", ",", "{", "}", ";", "(", ")",
	"pragma", "import", "contract", "library", "function", "struct", "enum", "mapping", "Type", "modifier",
	"event", "using", "var", "break"}

func (t Token) String() string { return tokLit[t] }
