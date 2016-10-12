package solparse

import (
	"errors"
	"fmt"
	"io"
)

type Node interface{}

type ContractDefinition struct {
	SourceLocation string
	Name           string
	SubNodes       []Node
	IsLibrary      bool
}

// Declaration represents a solidity variable decleration
type DeclarationStatement struct {
	Type      string
	Name      string
	Modifiers []string // public, private, const, etc
}

// Parser represents a parser.
type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// Parse the buffer
func (p *Parser) Parse() (cd *ContractDefinition, err error) {
	// Must be import, pragma, contract, library
	tok, _ := p.scanIgnoreWhitespace()
	switch tok {
	case PRAGMA:
		return cd, errors.New("pragma not yet implemented")
	case IMPORT:
		return cd, errors.New("import not yet implemented")
	case CONTRACT, LIBRARY:
		cd, err = p.parseContractDefination(false)
	default:
		err = errors.New("Expected import directive or contract definition")
	}

	return
}

// Parses contract or library definition
func (p *Parser) parseContractDefination(isLib bool) (cd *ContractDefinition, err error) {
	cd = &ContractDefinition{}

	_, cd.Name, err = p.expectIdentifier()
	if err != nil {
		return
	}

	//
	// if next token -> `is`, loop through and parseInheritance()
	//
	//
	// expect LEFT_BRACE
	//
	// Loop while true
	//
	//   if RIGHT BRACE, Break!
	//
	//   if function, parseFunctionDifintion
	//
	//   if struct, parseStructDefinition
	//
	//   if enum, parseEnumDefinition
	//
	//   if Identifier, Mapping, or Elem. type
	//     -> parseVariableDeclaration;
	//        expect SEMICOLON
	//
	//
	return
}

// returns next token from the underlying scanner, but only if it equals the given token
// returns error if the tokens do not equal
func (p *Parser) expectToken(expTok Token) (tok Token, lit string, err error) {
	tok, lit = p.scanIgnoreWhitespace()
	if tok != expTok {
		return tok, lit, fmt.Errorf("Parse error. Expected %d, got %s", expTok, lit)
	}
	return
}

// returns next identifier token. Returns error if next token is not an identifier
func (p *Parser) expectIdentifier() (tok Token, lit string, err error) {
	tok, lit = p.scanIgnoreWhitespace()
	if tok != IDENT {
		return tok, lit, fmt.Errorf("Parse error. Expected identifier, got: %q", lit)
	}
	return
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }
