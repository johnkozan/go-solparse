package solparse

import (
	"fmt"
	"strings"
)

type Token int

const (
	// End of source indicator
	EOS Token = iota

	// Punctuators
	LParen
	RParen
	LBrack
	RBrack
	LBrace
	RBrace
	Colon
	Semicolon
	Period
	Conditional
	Arrow

	// Assignment
	Assign
	AssignBitOr
	AssignBitXor
	AssignBitAnd
	AssignShl
	AssignSar
	AsignShr
	AssignAdd
	AssignSub
	AssignMul
	AssignDiv
	AssignMod

	// Binary operators, sorted by precedence
	Comma
	Or
	And
	BitOr
	BitXor
	BitAnd
	SHL
	SAR
	SHR
	Add
	Sub
	Mul
	Div
	Mod
	Exp

	// Compare operators, sorted by precedence
	Equal
	NotEqual
	LessThan
	GreaterThan
	LessThanOrEqual
	GreaterThanOrEqual

	// Unary operators
	Not
	BitNot
	Inc
	Dec
	Delete

	// Keywords
	Anonymous
	As
	Assemby
	Break
	Const
	Continue
	Contract
	Default
	Do
	Else
	Enum
	Event
	External
	For
	Function
	Hex
	If
	Indexed
	Internal
	Import
	Is
	Library
	Mapping
	Memory
	Modifier
	New
	Payable
	Public
	Pragma
	Private
	Return
	Returns
	Storage
	Struct
	Throw
	Using
	Var
	While

	// Ether subdenominations
	SubWei
	SubSzabo
	SubFinney
	SubEther
	SubSecond
	SubMinute
	SubHour
	SubDay
	SubWeek
	SubYear

	// Keywords
	Int
	Uint
	Bytes
	Byte
	String
	Address
	Bool
	Fixed
	UFixed
	IntM
	UIntM
	BytesM
	FixedMxN
	UFixedMxN
	TypesEnd // used as type enum end marker

	// Literals
	NullLiteral
	TrueLiteral
	FalseLiteral
	Number
	StringLiteral
	CommentLiteral

	// Identifiers
	Identifier

	// Keywords reserved for future use
	Abstract
	After
	Case
	Catch
	Final
	In
	Inline
	Interface
	Let
	Match
	Of
	Pure
	Relocatable
	Static
	Switch
	Try
	Type
	TypeOf
	View

	// Illegal token
	Illegal

	// Scanner internal use only
	Whitespace
)

var tokenLiterals = []struct {
	Name       string
	Precedence int
}{
	{"EOS", 0},

	{"(", 0},
	{")", 0},
	{"[", 0},
	{"]", 0},
	{"{", 0},
	{"}", 0},
	{":", 0},
	{";", 0},
	{".", 0},
	{"?", 3},
	{"=>", 0},

	{"=", 2},
	{"|=", 2},
	{"^=", 2},
	{"&=", 2},
	{"<<=", 2},
	{">>=", 2},
	{">>>=", 2},
	{"+=", 2},
	{"-=", 2},
	{"*=", 2},
	{"/=", 2},
	{"%=", 2},

	{",", 1},
	{"||", 4},
	{"&&", 5},
	{"|", 8},
	{"^", 9},
	{"&", 10},
	{"<<", 11},
	{">>", 11},
	{">>>", 11},
	{"+", 12},
	{"-", 12},
	{"*", 13},
	{"/", 13},
	{"%", 13},
	{"**", 14},

	{"==", 6},
	{"!=", 6},
	{"<", 7},
	{">", 7},
	{"<=", 7},
	{">=", 7},

	{"!", 0},
	{"~", 0},
	{"++", 0},
	{"--", 0},
	{"delete", 0},

	{"anonymous", 0},
	{"as", 0},
	{"assembly", 0},
	{"break", 0},
	{"constant", 0},
	{"continue", 0},
	{"contract", 0},
	{"default", 0},
	{"do", 0},
	{"else", 0},
	{"enum", 0},
	{"event", 0},
	{"external", 0},
	{"for", 0},
	{"function", 0},
	{"hex", 0},
	{"if", 0},
	{"indexed", 0},
	{"internal", 0},
	{"import", 0},
	{"is", 0},
	{"library", 0},
	{"mapping", 0},
	{"memory", 0},
	{"modifier", 0},
	{"new", 0},
	{"payable", 0},
	{"public", 0},
	{"pragma", 0},
	{"private", 0},
	{"return", 0},
	{"returns", 0},
	{"storage", 0},
	{"struct", 0},
	{"throw", 0},
	{"using", 0},
	{"var", 0},
	{"while", 0},

	{"wei", 0},
	{"szabo", 0},
	{"finney", 0},
	{"ether", 0},
	{"seconds", 0},
	{"minutes", 0},
	{"hours", 0},
	{"days", 0},
	{"weeks", 0},
	{"years", 0},

	{"int", 0},
	{"uint", 0},
	{"bytes", 0},
	{"byte", 0},
	{"string", 0},
	{"address", 0},
	{"bool", 0},
	{"fixed", 0},
	{"ufixed", 0},
	{"intM", 0},
	{"uintM", 0},
	{"bytesM", 0},
	{"fixedMxN", 0},
	{"ufixedMxN", 0},
	{"", 0},

	{"null", 0},
	{"true", 0},
	{"false", 0},
	{"", 0},
	{"", 0},
	{"", 0},

	{"", 0},

	{"abstract", 0},
	{"after", 0},
	{"case", 0},
	{"catch", 0},
	{"final", 0},
	{"in", 0},
	{"inline", 0},
	{"interface", 0},
	{"let", 0},
	{"match", 0},
	{"of", 0},
	{"pure", 0},
	{"relocatable", 0},
	{"static", 0},
	{"switch", 0},
	{"try", 0},
	{"type", 0},
	{"typeof", 0},
	{"view", 0},

	{"", 0},
}

func (t Token) String() string { return tokenLiterals[t].Name }

func stringToToken(s string) (tok Token, lit string) {
	//If the string matches a keyword then return that keyword.
	for k, v := range tokenLiterals {
		if v.Name == strings.ToLower(s) {
			return Token(k), s
		}
	}
	// Otherwise return as a regular identifier.
	return Identifier, s
}

func isVisibilitySpecifier(tok Token) bool {
	return tok == External || tok == Public || tok == Internal || tok == Private
}

func isLocationSpecifier(tok Token) bool {
	return tok == Memory || tok == Storage
}

func isUnaryOp(tok Token) bool {
	return (Not <= tok && tok <= Delete) || tok == Add || tok == Sub
}

func isCountOp(tok Token) bool {
	return tok == Inc || tok == Dec
}

func isElementaryTypeName(lit string) bool {
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
