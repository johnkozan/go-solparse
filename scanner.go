package solparse

import (
	"bufio"
	"bytes"
	"io"
)

// Solidity scanner

type Scanner struct {
	r       *bufio.Reader
	curTok  *TokenDesc
	nextTok *TokenDesc
	char    rune // 1 character look-ahead
}

type TokenDesc struct {
	token    Token
	location SourceLocation
	lit      LiteralScope
	info     ExtendedTokenInfo
}

type LiteralScope struct {
	literal  string
	complete bool
	buf      bytes.Buffer
}

type SourceLocation struct {
	start      int
	end        int
	sourceName string
}

type ExtendedTokenInfo struct {
	firstSize  int
	secondSize int
}

func NewScanner(r io.Reader) *Scanner {
	s := &Scanner{r: bufio.NewReader(r)}
	s.reset()
	return s
}

func (s *Scanner) advance() rune {
	var err error
	s.char, _, err = s.r.ReadRune()
	if err != nil {
		s.char = eof
		return eof
	}
	return s.char
}

func (s *Scanner) next() Token {
	s.curTok = s.nextTok
	s.scanToken()
	return s.curTok.token
}

func (s *Scanner) selectToken(then Token) Token {
	s.advance()
	return then
}

func (s *Scanner) scanToken() {
	s.nextTok = &TokenDesc{}

	var m, n int
	var tok Token
	for {
		switch s.char {
		case '\n', ' ', '\t':
			tok = s.selectToken(Whitespace)
		case '"', '\'':
			tok = s.scanString()
		case '<':
			panic("handle <")
		case '>':
			panic("handle >")
		case '=':
			panic("handle =")
		case '!':
			panic("handle !")
		case '-':
			panic("handle -")
		case '*':
			panic("handle *")
		case '%':
			panic("handle mod")
		case '/':
			panic("handle /")
		case '&':
			panic("handle &")
		case '|':
			panic("handle |")
		case '^':
			panic("handle ^")
		case '.':
			panic("handle number")
		case ':':
			tok = s.selectToken(Colon)
		case ';':
			tok = s.selectToken(Semicolon)
		case ',':
			tok = s.selectToken(Comma)
		case '(':
			tok = s.selectToken(LParen)
		case ')':
			tok = s.selectToken(RParen)
		case '[':
			tok = s.selectToken(LBrack)
		case ']':
			tok = s.selectToken(RBrack)
		case '{':
			tok = s.selectToken(LBrace)
		case '}':
			tok = s.selectToken(RBrace)
		case '?':
			tok = s.selectToken(Conditional)
		case '~':
			tok = s.selectToken(BitNot)
		default:
			if isIdentifierStart(s.char) {
				tok, m, n = s.scanIdentifierOrKeyword()
				if tok == Hex {
					m, n = 0, 0
					if s.char == '"' || s.char == '\'' {
						tok = s.scanHexString()
					} else {
						tok = Illegal
					}
				}
			} else if isDecimalDigit(s.char) {
				tok = s.scanNumber(s.char)
			} else if s.char == eof {
				tok = EOS
			} else {
				tok = s.selectToken(Illegal)
			}
			// skipWhitespcae ?
		}
		if tok != Whitespace {
			break
		}
	}
	info := ExtendedTokenInfo{firstSize: m, secondSize: n}
	s.nextTok = &TokenDesc{token: tok, info: info, lit: s.nextTok.lit}
}

func (s *Scanner) scanIdentifierOrKeyword() (Token, int, int) {
	if !isIdentifierStart(s.char) {
		panic("is not identifier start")
	}
	s.addLiteralCharAndAdvance()
	for isIdentifierPart(s.char) {
		s.addLiteralCharAndAdvance()
	}
	return tokenFromIdentifierOrKeyword(s.nextTok.lit.String())
}

func (s *Scanner) addLiteralCharAndAdvance() {
	s.addLiteralChar(s.char)
	s.advance()
}

func (s *Scanner) scanDecimalDigits() {
	for isDecimalDigit(s.char) {
		s.addLiteralCharAndAdvance()
	}
}

func (s *Scanner) addLiteralChar(c rune) {
	s.nextTok.lit.buf.WriteRune(c)
}

func (s *Scanner) currentToken() Token {
	return s.curTok.token
}

func (s *Scanner) currentLiteral() string {
	return s.curTok.lit.String()
}

func (s *Scanner) peekNextToken() Token {
	return s.nextTok.token
}

func (l LiteralScope) String() string {
	return l.buf.String()
}

type NumberKind int

const (
	DecimalKind NumberKind = iota
	HexKind
	BinaryKind
)

func (s *Scanner) scanNumber(charSeen rune) (tok Token) {
	kind := DecimalKind
	s.nextTok.lit = LiteralScope{}

	if charSeen == '.' {
		s.addLiteralChar('.')
		s.scanDecimalDigits()
	} else {
		if charSeen == '0' {
			s.addLiteralCharAndAdvance()
			if s.char == 'x' || s.char == 'X' {
				kind = HexKind
				s.addLiteralCharAndAdvance()
				if !isHexDigit(s.char) {
					return Illegal
				}
				for isHexDigit(s.char) {
					s.addLiteralCharAndAdvance()
				}
			}
		}

		if kind == DecimalKind {
			s.scanDecimalDigits()
			if s.char == '.' {
				s.addLiteralCharAndAdvance()
				s.scanDecimalDigits()
			}
		}
	}

	// scan exponent, if any
	if s.char == 'e' || s.char == 'E' {
		if kind != DecimalKind {
			return Illegal
		}
		s.addLiteralCharAndAdvance()
		if s.char == '+' || s.char == '-' {
			s.addLiteralCharAndAdvance()
		}
		if !isDecimalDigit(s.char) {
			return Illegal
		}
		s.scanDecimalDigits()
	}

	// The source char immediately following a numberic literal must not be an identifier part or deciaml digit
	if isDecimalDigit(s.char) || isIdentifierStart(s.char) {
		return Illegal
	}

	return Number
}

func (s *Scanner) scanString() Token {
	quote := s.char
	s.nextTok.lit = LiteralScope{}
	s.advance() // consume quote
	for s.char != quote && s.char != eof && !isLineTerminator(s.char) {
		c := s.char
		s.advance()
		if c == '\\' {
			if !s.scanEscape() {
				return Illegal
			}
		} else {
			s.addLiteralChar(c)
		}
	}
	if s.char != quote {
		return Illegal
	}
	s.advance()
	return StringLiteral
}

func (s *Scanner) scanEscape() bool {
	c := s.char
	s.advance()
	if isLineTerminator(c) {
		return true
	}

	switch c {
	case '\'', '"', '\\':
	case 'b':
		c = '\b'
	case 'f':
		c = '\f'
	case 'n':
		c = '\n'
	case 'r':
		c = '\r'
	case 't':
		c = '\t'
	case 'v':
		c = '\v'
	case 'u':
		panic("unicode not yet implemted")
		//var codepoint rune
		//if !s.scanUnicode(&codepoint) {
		//return false
		//}
		//s.buf.WriteRune(codepoint)
	case 'x':
		var ok bool
		if c, ok = s.scanHexByte(); !ok {
			return false
		}
	}

	s.addLiteralChar(c)
	return true
}

func (s *Scanner) scanUnicode(cp *rune) bool {
	panic("unicode not yet implemented")
}

func (s *Scanner) scanHexString() Token {
	panic("hex string not yet implemented")
}

func (s *Scanner) scanHexByte() (rune, bool) {
	x := 0
	for i := 0; i < 2; i++ {
		d := hexValue(s.char)
		if d < 0 {
			s.rollback(i)
			return rune(0), false
		}
		x = x*16 + d
		s.advance()
	}
	return rune(x), true
}

func (s *Scanner) skipWhitespace() bool {
	ret := false
	for isWhitespace(s.char) {
		ret = true
		s.advance()
	}
	return ret
}

func (s *Scanner) reset() {
	// source.reset()
	var err error
	s.char, _, err = s.r.ReadRune()
	if err != nil {
		s.char = eof
	}
	s.skipWhitespace()
	s.scanToken()
	s.next()
}

func (s *Scanner) rollback(i int) {
	if i <= 0 {
		return
	}
	for x := 0; x < i+1; x++ {
		_ = s.r.UnreadRune()
	}
	var err error
	s.char, _, _ = s.r.ReadRune()
	if err != nil {
		s.char = eof
	}
}

// eof represents a marker rune for the end of the reader.
var eof = rune(-1)

func isIdentifierStart(ch rune) bool {
	return ch == '_' || ch == '$' || ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z')
}
func isHexDigit(ch rune) bool {
	return isDecimalDigit(ch) || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
}
func isIdentifierPart(ch rune) bool { return isIdentifierStart(ch) || isDecimalDigit(ch) }
func isWhitespace(ch rune) bool     { return ch == ' ' || ch == '\t' || ch == '\n' }
func isLetter(ch rune) bool         { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }
func isDecimalDigit(ch rune) bool   { return (ch >= '0' && ch <= '9') }
func isLineTerminator(ch rune) bool { return ch == '\n' }
func hexValue(ch rune) int {
	if ch >= '0' && ch <= '9' {
		return int(ch - '0')
	}
	return -1
}
