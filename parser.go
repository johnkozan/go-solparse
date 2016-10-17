package solparse

import (
	"errors"
	"fmt"
	"io"
)

type Node interface{}

type ContractDefinition struct {
	BaseContracts  []ContractDefinition
	SourceLocation string
	Name           string
	SubNodes       []Node
	IsLibrary      bool
}

// Declaration represents a solidity variable decleration
type VariableDeclaration struct {
	Type            TypeName
	Identifier      string
	Value           Expression
	IsStateVariable bool
	IsIndexed       bool
	IsDeclaredConst bool
	Location        string
}

// Assignment and Conditional should implement Expression
type Expression interface{}

type Statement struct {
	Token      Token
	Expression Expression
}

type IdentifierExpression struct {
	Literal string
}

type AssignmentExpression struct {
	Expression         Token
	AssignmentOperator Token
	RightHandSide      string
}

type ConditionalExpression struct {
	TrueExpression  string
	FalseExpression string
}

type UnaryOperation struct {
	Token         Token
	SubExpression Expression
	Something     bool
}

type BinaryOperation struct {
	Expression    Expression
	Operation     Token
	RightHandSide Expression
}

type TypeName interface{}

type ElementaryTypeName struct {
	Token
	firstSize  int
	secondSize int
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
	s *Scanner
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// Parse the buffer
func (p *Parser) Parse() (cd *ContractDefinition, err error) {
	// Must be import, pragma, contract, library
	tok := p.currentToken()
	for tok != EOS {
		switch tok {
		case Pragma:
			return cd, errors.New("pragma not yet implemented")
		case Import:
			return cd, errors.New("import not yet implemented")
		case Contract, Library:
			cd, err = p.parseContractDefination(false)
			if err != nil {
				return cd, err
			}
		default:
			err = errors.New("Expected import directive or contract definition")
		}
		tok = p.next()
	}

	return
}

// Parses contract or library definition
func (p *Parser) parseContractDefination(isLib bool) (cd *ContractDefinition, err error) {
	cd = &ContractDefinition{}
	if isLib {
		err = p.expectToken(Library)
	} else {
		err = p.expectToken(Contract)
	}
	if err != nil {
		return
	}
	cd.IsLibrary = isLib
	cd.Name, err = p.expectIdentifierToken()
	if err != nil {
		return cd, err
	}
	if p.currentToken() == Is {
		cd.BaseContracts, err = p.parseInheritanceSpecifier()
		if err != nil {
			return cd, err
		}
	}

	err = p.expectToken(LBrace)
	if err != nil {
		return
	}

outer:
	for {
		tok := p.currentToken()
		switch {
		case tok == RBrace:
			break outer
		case tok == Function:
			fd, err := p.parseFunctionDefinition()
			if err != nil {
				return cd, err
			}
			cd.SubNodes = append(cd.SubNodes, fd)
		case tok == Struct:
			return cd, errors.New("struct not yet implemented")
		case tok == Enum:
			return cd, errors.New("enum not yet implemented")
		case tok == Identifier || tok == Mapping || isElementaryTypeName(tok):
			vd, err := p.parseVariableDeclaration()
			if err != nil {
				return cd, err
			}
			cd.SubNodes = append(cd.SubNodes, vd)
			err = p.expectToken(Semicolon)
			if err != nil {
				return cd, err
			}
		case tok == Modifier:
			return cd, errors.New("modifier not yet implemented")
		case tok == Event:
			return cd, errors.New("event not yet implemented")
		case tok == Using:
			return cd, errors.New("using not yet implemented")
		default:
			return cd, errors.New("Function, variable, struct or modifier declaration expected")
		}
	}
	err = p.expectToken(RBrace)
	return
}

func (p *Parser) parseFunctionDefinition() (f FunctionDefinition, err error) {
	err = p.expectToken(Function)
	if err != nil {
		return f, err
	}
	tok := p.currentToken()
	if tok == LBrace {
		// anonymous function
	} else {
		f.Name, err = p.expectIdentifierToken()
		if err != nil {
			return f, err
		}
	}

	f.Paramaters, err = p.parseParameterList() // pass in options

	// Parse function modifiers like constant
	for {
		tok := p.currentToken()
		if tok == Const {
			f.IsDeclaredConst = true
			p.next()
		} else if tok == Payable {
			f.IsPayable = true
			p.next()
		} else if isVisibilitySpecifier(tok) {
			f.Visibility, err = p.parseVisibilitySpecifier()
			if err != nil {
				return f, err
			}
		} else if tok == Identifier {
			_, err := p.parseModifierInvocation()
			if err != nil {
				return f, err
			}
			// Add modifer to function def
		} else {
			break
		}
	}

	tok = p.currentToken()
	if tok == Returns {
		p.next()
		f.ReturnParameters, err = p.parseParameterList() // allowEmptyParamaterList = false
		if err != nil {
			return f, err
		}
	}

	tok = p.currentToken()
	if tok != Semicolon {
		f.Block, err = p.parseBlock()
		if err != nil {
			return f, err
		}
	} else {
		p.next() // consume ;
	}

	// If f.Name == _contractName { f.Constructor = true }
	return
}

func (p *Parser) parseBlock() (b Block, err error) {
	err = p.expectToken(LBrace)
	if err != nil {
		return
	}

	for p.currentToken() != RBrace {
		stmt, err := p.parseStatement()
		if err != nil {
			return b, err
		}
		b.Statements = append(b.Statements, stmt)
	}

	err = p.expectToken(RBrace)
	return b, err
}

func (p *Parser) parseStatement() (e Expression, err error) {
	s := Statement{}
	// check for comment
	tok := p.currentToken()
	switch tok {
	case If:
		return p.parseIfStatement()
	case While:
		return p.parseWhileStatement()
	case For:
		return p.parseForStatement()
	case LBrace:
		return p.parseBlock()
	case Continue, Break, Throw:
		s.Token = tok
		_ = p.next()
	case Return:
		// Build expression
		e := Statement{}
		if p.next() != Semicolon {
			e, err := p.parseExpression()
			if err != nil {
				return e, err
			}
		} else {
			s.Expression = e
		}
	case Assemby:
		return p.parseInlineAssembly()
	case Identifier:
		if p.currentLiteral() == "_" { // and inside modifer {
			return e, errors.New("'_' not yet implemented")
		}
		// if inside of a function modifier and current Literal = '_'
		//  -> statement = PlaceHolderStatement
		fallthrough
	default:
		s.Expression, err = p.parseSimpleStatement()
		if err != nil {
			return s, err
		}
	}

	err = p.expectToken(Semicolon)
	return s, err
}

func (p *Parser) parseInheritanceSpecifier() (cds []ContractDefinition, err error) {
	return cds, errors.New("Inheritance specifiers not yet implemented")
}

func (p *Parser) parseParameterList() (pl ParameterList, err error) {
	err = p.expectToken(LParen)
	if err != nil {
		return
	}

	if p.currentToken() != RParen { // || !_allowEmpty
		vd, err := p.parseVariableDeclaration()
		if err != nil {
			return pl, err
		}
		pl.Paramaters = append(pl.Paramaters, vd)

		tok := p.currentToken()
		for tok != RParen {
			err = p.expectToken(Comma)
			if err != nil {
				return pl, err
			}
			vd, err := p.parseVariableDeclaration()
			if err != nil {
				return pl, err
			}
			pl.Paramaters = append(pl.Paramaters, vd)
			tok = p.currentToken()
		}
	}
	p.next()
	return
}

func (p *Parser) parseVariableDeclaration() (v VariableDeclaration, err error) {
	// if lookAheadArrayType
	//
	// else {
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

	v.Identifier, err = p.expectIdentifierToken()

	// if allowInitialValue -> Check for ASSIGN, parseExpression
	if p.currentToken() == Assign {
		p.next()
		v.Value, err = p.parseExpression()
		if err != nil {
			return v, err
		}

	}
	return
}

func (p *Parser) parseTypeName() (t TypeName, err error) {
	tok := p.currentToken()
	if isElementaryTypeName(tok) {
		firstSize, secondSize := p.currentTokenInfo()
		t := ElementaryTypeName{tok, firstSize, secondSize}
		p.next()
		return t, nil
	}

	switch tok {
	case Var:
		//   return error if var not allowed (by option)
	case Mapping:
		t, err = p.parseMapping()
	case Identifier:
		t, err = p.parseUserDefinedTypeName()
	default:
		return t, errors.New("Expected type name")
	}
	if err != nil {
		return t, err
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

func (p *Parser) parseMapping() (t TypeName, err error) {
	return t, errors.New("mapping not yet implemented")
}

func (p *Parser) parseUserDefinedTypeName() (t TypeName, err error) {
	return t, errors.New("user defined types not yet implemented")
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

func (p *Parser) parseExpression() (e Expression, err error) {
	e, err = p.parseBinaryExpression(4) // + lookaheadDataStructure
	if err != nil {
		return e, err
	}

	if isAssignmentOp(p.currentToken()) {
		return e, errors.New("assignment expression not yet implemented")
		//assignmentOp := p.expectAssignmentOperator()
	} else if p.currentToken() == Conditional {
		return e, errors.New("conditional not yet implemented")
		//p.next()
	}

	return
}

func (p *Parser) parseInlineAssembly() (s Statement, err error) {
	return s, errors.New("inline assembly not yet implemented")
}

func (p *Parser) parseSimpleStatement() (e Expression, err error) {
	switch p.peekStatementType() {
	case VariableDeclarationStatement:
		return p.parseVariableDeclarationStatement()
	case ExpressionStatement:
		return p.parseExpressionStatement()
	}

	// TODO: Deal with [] and . variable access stuff

	return
}

func (p *Parser) parseVariableDeclarationStatement() (s Statement, err error) {
	return s, errors.New("implement variable decleartion stmt")
}

func (p *Parser) parseExpressionStatement() (s Expression, err error) {
	return p.parseExpression()
}

func (p *Parser) parseBinaryExpression(minPrecedence int) (e Expression, err error) {
	e, err = p.parseUnaryExpression()
	if err != nil {
		return e, err
	}

	// TODO: do we need tokenPrecence() or is this info stored with Token in Scanner
	precedence := tokenPrecedence(p.currentToken())
	for ; precedence >= minPrecedence; precedence-- {
		for tokenPrecedence(p.currentToken()) == precedence {
			op := p.currentToken()
			p.next()
			right, err := p.parseBinaryExpression(precedence + 1)
			if err != nil {
				return e, err
			}
			e = BinaryOperation{e, op, right}
		}
	}
	return
}

func (p *Parser) parseUnaryExpression() (e Expression, err error) {
	u := UnaryOperation{}
	u.Token = p.currentToken()

	if isUnaryOp(u.Token) || isCountOp(u.Token) {
		// prefix expression
		p.next()
		u.SubExpression, err = p.parseUnaryExpression()
		return u, err
	} else {
		u.SubExpression, err = p.parseLeftHandSideExpression()
		if err != nil {
			return u, err
		}
		tok := p.currentToken()
		if !isCountOp(tok) {
			return u.SubExpression, nil
		}
		p.next()
	}
	return
}

func (p *Parser) parseLeftHandSideExpression() (e Expression, err error) {
	// if we were passed the lookAheadDataType, e = lookAheadDataType
	// } else {
	tok := p.currentToken()
	if tok == New {
		return e, errors.New("new not yet implemented")
		// contract name = p.parseTypeName(false)
		// e = contract
	} else {
		e, err = p.parsePrimaryExpression()
		if err != nil {
			return e, err
		}
	}

out:
	for {
		switch p.currentToken() {
		case LBrack:
			p.next()
			if p.currentToken() != RBrack {
				index, err := p.parseExpression()
				if err != nil {
					return e, err
				}
				err = p.expectToken(RBrack)
				if err != nil {
					return e, err
				}
				return index, errors.New("index access not yet implemented")
			}
		case Period:
			p.next()
			return e, errors.New("member access not yet implemented")
		case LParen:
			p.next()
			return e, errors.New("function calls not yet implemented")
		default:
			break out
		}
	}
	return
}

func (p *Parser) parsePrimaryExpression() (e Expression, err error) {
	tok := p.currentToken()
	switch tok {
	case TrueLiteral, FalseLiteral:
		//l := Literal{}
		// Expression.Literal = p.getLiteralAndAdvance() // or currenetLIteral ??
		// expression.Token = tok
	case Number:
		return e, errors.New("number expression not yet implemented")
		//fallthrough
	case StringLiteral:
		// expression.Literal = p.getLiteralAndAdvance()
		// expression.Token = tok
	case Identifier:
		l := p.getLiteralAndAdvance()
		return IdentifierExpression{l}, err
	case LParen, LBrack:
		return e, errors.New("tuples / paranthesized expression not yet implemented")
	default:
		if isElementaryTypeName(tok) {
			firstSize, secondSize := p.currentTokenInfo()
			return ElementaryTypeName{tok, firstSize, secondSize}, nil
		} else {
			return e, errors.New("Expected primary expression")
		}
	}
	return
}

func (p *Parser) getLiteralAndAdvance() (lit string) {
	lit = p.currentLiteral()
	p.next()
	return
}

//returns next token from the underlying scanner, but only if it equals the given token
//returns error if the tokens do not equal
func (p *Parser) expectToken(expTok Token) error {
	tok := p.s.currentToken()
	if tok != expTok {
		return fmt.Errorf("Parse error: expected '%s' got '%s'", expTok, tok)
	}
	p.next()
	return nil
}

func (p *Parser) expectIdentifierToken() (lit string, err error) {
	tok := p.s.currentToken()
	lit = p.s.currentLiteral()
	if tok != Identifier {
		return lit, fmt.Errorf("Expected identifier, got '%s", lit)
	}
	p.next()
	return
}

func (p *Parser) currentToken() Token {
	return p.s.currentToken()
}

func (p *Parser) currentLiteral() string {
	return p.s.currentLiteral()
}

func (p *Parser) currentTokenInfo() (int, int) {
	_ = p.currentLiteral()
	//baseType = // position of M
	//tok := keywordByName(baseType)
	//switch tok {
	//case Bytes:
	//return m, 0
	//case Uint, Int:
	//return m, 0
	//case UFixed, Fixed:
	//return m, n
	//}

	return 0, 0
}

// Advance the scanner
func (p *Parser) next() Token {
	return p.s.next()
}

type StatementType int

const (
	ExpressionStatement StatementType = iota
	VariableDeclarationStatement
	IndexAccessStructure
)

// Distinguish between variable declaration (and potentially assignment) and expression statement
func (p *Parser) peekStatementType() StatementType {
	tok := p.currentToken()
	mightBeTypeName := (isElementaryTypeName(tok) || tok == Identifier)

	if (tok == Mapping) || (tok == Var) {
		return VariableDeclarationStatement
	}
	if mightBeTypeName {
		nextTok := p.s.peekNextToken()
		if (nextTok == Identifier) || isLocationSpecifier(nextTok) {
			return VariableDeclarationStatement
		}
		if (nextTok == LBrack) || (nextTok == Period) {
			return IndexAccessStructure
		}
	}
	return ExpressionStatement
}
