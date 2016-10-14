package solparse

import (
	"bufio"
	"bytes"
	"io"
)

// Solidity scanner

type Scanner struct {
	r   *bufio.Reader
	buf bytes.Buffer
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

// Returns the last scanned literal.  Does not advance the scanner.
func (s *Scanner) CurrentLiteral() string {
	return s.buf.String()
}

// Scan returns the next token and literal value.  Ignores whitespace
func (s *Scanner) Scan() (tok Token, lit string) {
	// Read the next rune
	ch := s.read()

	// If we see whitespace then consume all contiguous whitespace.
	// If we see a letter then consume as an ident or reserved word
	if isWhitespace(ch) {
		s.unread()
		s.scanWhitespace()
		return s.Scan()
	} else if isLetter(ch) {
		s.unread()
		return s.scanIdent()
	} else if isDigit(ch) {
		s.unread()
		return s.scanNumber()
	}

	// Otherwise read the individual character
	switch ch {
	case eof:
		return EOS, ""
	case '"', '\'':
		s.unread()
		return s.scanString()
	default:
		tok, lit := stringToToken(string(ch))
		if tok != Identifier {
			return tok, lit
		}
	}
	return Illegal, string(ch)
}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	s.buf = bytes.Buffer{}
	s.buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			s.buf.WriteRune(ch)
		}
	}

	return Whitespace, s.buf.String()
}

// scanIdent consumes the current rune and all contiguous ident runes.
func (s *Scanner) scanIdent() (tok Token, lit string) {
	// Create a buffer and read the current character into it.
	s.buf = bytes.Buffer{}
	s.buf.WriteRune(s.read())

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) && ch != '_' {
			s.unread()
			break
		} else {
			_, _ = s.buf.WriteRune(ch)
		}
	}

	return stringToToken(s.buf.String())
}

type NumberKind int

const (
	DecimalKind NumberKind = iota
	HexKind
	BinaryKind
)

func (s *Scanner) scanNumber() (tok Token, lit string) {
	kind := DecimalKind

	// Create a buffer and read the current character into it.
	s.buf = bytes.Buffer{}
	//buf.WriteRune(charSeen)

	//if charSeen == '.' {
	// do stuff
	//} // else {
	//if charSeen == '0' {
	//r := s.scan()
	//buf.WriteRune(r)
	//if r == 'x' || r == 'X' {
	//hex
	//kind = HEX
	//}
	//}

	if kind == DecimalKind {
		// scan all digits
		for {
			if ch := s.read(); ch == eof {
				break
			} else if !isDigit(ch) && ch != '.' {
				s.unread()
				break
			} else {
				_, _ = s.buf.WriteRune(ch)
			}
		}
	}
	// scan exponent, if any
	// do some check for not identifier or decimal

	return Number, s.buf.String()
}

func (s *Scanner) scanString() (tok Token, lit string) {
	s.buf = bytes.Buffer{}
	quote := s.read()

	for {
		if ch := s.read(); ch == eof || ch == quote || isLineTerminator(ch) {
			break
		} else {
			_, _ = s.buf.WriteRune(ch)
		}
		// Handle "\\"
	}
	return StringLiteral, s.buf.String()
}

func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() { _ = s.r.UnreadRune() }

// eof represents a marker rune for the end of the reader.
var eof = rune(0)

func isWhitespace(ch rune) bool     { return ch == ' ' || ch == '\t' || ch == '\n' }
func isLetter(ch rune) bool         { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }
func isDigit(ch rune) bool          { return (ch >= '0' && ch <= '9') }
func isLineTerminator(ch rune) bool { return ch == '\n' }
