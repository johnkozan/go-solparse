package solparse

import "fmt"

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
	CONST
	PAYABLE
	EXTERNAL
	PUBLIC
	INTERNAL
	PRIVATE
	RETURNS
	IF
	WHILE
	FOR
	CONTINUE
	RETURN
	THROW
	ASSEMBLY
)

var tokLit = []string{"ILLEGAL", "EOF", "<whitespace>", "Identifier", "String literal", "Number", ",", "{", "}", ";", "(", ")",
	"pragma", "import", "contract", "library", "function", "struct", "enum", "mapping", "Type", "modifier",
	"event", "using", "var", "break", "constant", "payable", "external", "public", "internal", "private", "returns",
	"if", "while", "for", "continue", "return", "throw", "assembly"}

func (t Token) String() string { return tokLit[t] }

var visibiltySpecifiers = []Token{EXTERNAL, PUBLIC, INTERNAL, PRIVATE}

func isVisibilitySpecifier(tok Token) bool {
	for _, v := range visibiltySpecifiers {
		if tok == v {
			return true
		}
	}
	return false
}

func isElementaryTypeName(lit string) bool {
	//log.Println("is vis: ", lit)
	for _, v := range elementaryTypes {
		if v == lit {
			return true
		}
	}
	return false
}

var elementaryTypes []string

func init() {
	// Build list of elementary types
	elementaryTypes = []string{"bool", "string", "int", "uint", "byte", "bytes", "address"}
	// int, uint by 8 to 256
	for i := 8; i <= 256; i += 8 {
		elementaryTypes = append(elementaryTypes, fmt.Sprintf("int%d", i))
		elementaryTypes = append(elementaryTypes, fmt.Sprintf("uint%d", i))
	}
	// bytes by 1 to 32
	for b := 1; b <= 32; b += 1 {
		elementaryTypes = append(elementaryTypes, fmt.Sprintf("bytes%d", b))
	}
}
