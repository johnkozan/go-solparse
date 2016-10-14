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
type VariableDeclaration struct {
	Type            string
	Identifier      string
	Value           string
	IsStateVariable bool
	IsIndexed       bool
	IsDeclaredConst bool
	Location        string
}

type Statement struct {
	Token      Token
	Expression Node
	Nodes      []Node
}

type Block struct {
	Statements []Node
}

type FunctionDefinition struct {
	Name             string
	Visibility       string
	IsConstructor    bool
	DocString        string
	Paramaters       ParameterList
	IsDeclaredConst  bool
	Modifiers        []string
	ReturnParameters ParameterList
	IsPayable        bool
	Block            Node
}

type ParameterList struct {
	Paramaters []VariableDeclaration
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
	tok, _ := p.scan()
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

	_, cd.Name, err = p.expectToken(IDENT)
	if err != nil {
		return
	}

	// if next token -> `is`, loop through and parseInheritance()

	_, _, err = p.expectToken(LBRACE)
	if err != nil {
		return
	}

outer:
	for {
		ct, _ := p.scan()
		switch ct {
		case RBRACE:
			break outer
		case FUNCTION:
			fd, err := p.parseFunctionDefinition()
			if err != nil {
				return cd, err
			}
			cd.SubNodes = append(cd.SubNodes, fd)
		case STRUCT:
			return cd, errors.New("struct not yet implemented")
		case ENUM:
			return cd, errors.New("enum not yet implemented")
		case IDENT, MAPPING, ELEM:
			p.unscan()
			vd, err := p.parseVariableDeclaration()
			if err != nil {
				return cd, err
			}
			cd.SubNodes = append(cd.SubNodes, vd)
			_, _, err = p.expectToken(SEMICOLON)
			if err != nil {
				return cd, err
			}
		case MODIFIER:
			return cd, errors.New("modifier not yet implemented")
		case EVENT:
			return cd, errors.New("event not yet implemented")
		case USING:
			return cd, errors.New("using not yet implemented")
		default:
			return cd, errors.New("Function, variable, struct or modifier declaration expected")
		}
	}
	return
}

func (p *Parser) parseFunctionDefinition() (f FunctionDefinition, err error) {
	tok, lit := p.scan()
	if tok == LBRACE {
		// anonymous function
	} else {
		f.Name = lit
	}

	f.Paramaters, err = p.parseParameterList() // pass in options

	// Parse function modifiers like constant
	for {
		tok, _ := p.scan()
		if tok == CONST {
			f.IsDeclaredConst = true
		} else if tok == PAYABLE {
			f.IsPayable = true
		} else if isVisibilitySpecifier(tok) {
			f.Visibility, err = p.parseVisibilitySpecifier()
			if err != nil {
				return f, err
			}
		} else if tok == IDENT {
			_, err := p.parseModifierInvocation()
			if err != nil {
				return f, err
			}
			// Add modifer to function def
		} else {
			break
		}
	}

	p.unscan()
	tok, _ = p.scan()
	if tok == RETURNS {
		f.ReturnParameters, err = p.parseParameterList() // allowEmptyParamaterList = false
		if err != nil {
			return f, err
		}
	}

	tok, lit = p.scan()
	if tok != SEMICOLON {
		p.unscan()
		f.Block, err = p.parseBlock()
		if err != nil {
			return f, err
		}
	}

	// If f.Name == _contractName { f.Constructor = true }
	return
}

func (p *Parser) parseBlock() (b Block, err error) {
	_, _, err = p.expectToken(LBRACE)
	if err != nil {
		return
	}
	for {
		tok, _ := p.scan()
		if tok == RBRACE {
			break
		}
		stmt, err := p.parseStatement()
		if err != nil {
			return b, err
		}
		b.Statements = append(b.Statements, stmt)
	}
	return b, err
}

func (p *Parser) parseStatement() (n Node, err error) {
	var s Statement
	// check for comment
	tok, _ := p.scan()
	switch tok {
	case IF:
		return p.parseIfStatement()
	case WHILE:
		return p.parseWhileStatement()
	case FOR:
		return p.parseForStatement()
	case LBRACE:
		return p.parseBlock()
	case CONTINUE, BREAK, THROW:
		s.Token = tok
	case RETURN:
		// Build expression
		// if token != ; ->
		_, err := p.parseExpression()
		if err != nil {
			return n, err
		}
		// put expression in statement
	case ASSEMBLY:
		return p.parseInlineAssembly()
	case IDENT:
		return n, errors.New("'_' not yet implemented")
		// if inside of a function modifier and current Literal = '_'
		//  -> statement = PlaceHolderStatement
	default:
		s, err = p.parseSimpleStatement()
		if err != nil {
			return s, err
		}
	}

	_, _, err = p.expectToken(SEMICOLON)
	return s, err
}

func (p *Parser) parseParameterList() (pl ParameterList, err error) {
	_, _, err = p.expectToken(LPAREN)
	if err != nil {
		return
	}

	tok, _ := p.scan()
	if tok != RPAREN {
		p.unscan()
		vd, err := p.parseVariableDeclaration()
		if err != nil {
			return pl, err
		}
		pl.Paramaters = append(pl.Paramaters, vd)

		for {
			tok, _ := p.scan()
			if tok == RPAREN {
				break
			} else if tok == COMMA {
				vd, err := p.parseVariableDeclaration()
				if err != nil {
					return pl, err
				}
				pl.Paramaters = append(pl.Paramaters, vd)
			}
		}
	}
	return
}

func (p *Parser) parseVariableDeclaration() (v VariableDeclaration, err error) {
	v.Type, err = p.parseTypeName() // option: allowVar
	if err != nil {
		return v, err
	}

	// next check const, indexed, storage, memory, etc
	//for {
	//tok, lit := p.currentToken

	// if isVariableVisibilitySpecifier(tok) -> set v.Visibility
	// else
	//   if INDEXED -> set Indexed
	//   else if CONST -> set Const
	//   else if isLocationSpecifier -> set location
	//   else break
	//break
	//}

	// if allowEmptyName && currentToken != IDENT { identifier = "" }

	_, v.Identifier, err = p.expectToken(IDENT)

	// if allowInitialValue -> Check for ASSIGN, parseExpression
	return
}

func (p *Parser) parseTypeName() (n string, err error) {
	tok, lit := p.scan()

	if isElementaryTypeName(lit) {
		return lit, nil
	}

	switch tok {
	case VAR:
		//   return error if var not allowed (by option)
	case MAPPING:
		n, err = p.parseMapping()
	case IDENT:
		n, err = p.parseUserDefinedTypeName()
	default:
		return n, errors.New("Expected type name")
	}
	if err != nil {
		return n, err
	}

	// Handle [...] postfix for arrays
	return
}

func (p *Parser) parseVisibilitySpecifier() (v string, err error) {
	return v, errors.New("visibilty specifiers not yet implemented")
}

func (p *Parser) parseModifierInvocation() (m string, err error) {
	return m, errors.New("function modifiers not yet implemented")
}

func (p *Parser) parseMapping() (m string, err error) {
	return m, errors.New("mapping not yet implemented")
}

func (p *Parser) parseUserDefinedTypeName() (u string, err error) {
	return u, errors.New("user defined types not yet implemented")
}

func (p *Parser) parseIfStatement() (s Statement, err error) {
	return s, errors.New("if statement not yet implemented")
}

func (p *Parser) parseWhileStatement() (s Statement, err error) {
	return s, errors.New("while statement not yet implemented")
}

func (p *Parser) parseForStatement() (s Statement, err error) {
	return s, errors.New("for statement not yet implemented")
}

func (p *Parser) parseExpression() (s Node, err error) {
	return s, errors.New("expression statement not yet implemented")
}

func (p *Parser) parseInlineAssembly() (s Statement, err error) {
	return s, errors.New("inline assembly not yet implemented")
}

func (p *Parser) parseSimpleStatement() (s Statement, err error) {
	return s, errors.New("simple statement not yet implemented")
}

// returns next token from the underlying scanner, but only if it equals the given token
// returns error if the tokens do not equal
func (p *Parser) expectToken(expTok Token) (tok Token, lit string, err error) {
	tok, lit = p.scan()
	if tok != expTok {
		return tok, lit, fmt.Errorf("Parse error: expected '%s' got '%s'", tokLit[expTok], lit)
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

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }
