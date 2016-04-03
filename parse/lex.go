package parse

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type token struct {
	kind  tokenType
	value string
}

type tokenType int

const (
	tokError tokenType = iota

	tokEOF

	tokDef
	tokExtern
	tokIf
	tokThen
	tokElse

	tokIdentifier
	tokNumber

	tokSemi
	tokComma
	tokLparen
	tokRparen

	tokOther

	// operators
	tokUserUnaryOp
	tokUserBinaryOp
	tokEqual
	tokPlus
	tokMinus
	tokStar
	tokSlash
	tokLessThan
)

var keywords = map[string]tokenType{
	"def":    tokDef,
	"extern": tokExtern,
	"if":     tokIf,
	"then":   tokThen,
	"else":   tokElse,
}

var op = map[rune]tokenType{
	'=': tokEqual,
	'+': tokPlus,
	'-': tokMinus,
	'*': tokStar,
	'/': tokSlash,
	'<': tokLessThan,
}

type userOpType int

const (
	uopNOP userOpType = iota
	uopUnaryOp
	uopBinaryOp
)

type stateFn func(*lexer) stateFn

// lexer has a scanner state.
type lexer struct {
	input         string
	name          string
	pos           int
	start         int
	width         int
	tokens        chan token
	state         stateFn
	userOperators map[rune]userOpType
}

// Lex creates a new lexer.
func lex(name, input string) *lexer {
	l := &lexer{
		name:          name,
		input:         input,
		tokens:        make(chan token, 10),
		userOperators: map[rune]userOpType{},
	}
	go l.run()
	return l
}

func (l *lexer) run() {
	for l.state = lexToplevel; l.state != nil; {
		l.state = l.state(l)
	}
	close(l.tokens)
}

// Nexttoken returns the next token.
func (l *lexer) nextToken() token {
	return <-l.tokens
}

func (l *lexer) word() string {
	return l.input[l.start:l.pos]
}

const eof = -1

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += w
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) emit(t tokenType) {
	l.tokens <- token{kind: t, value: l.word()}
	l.start = l.pos
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token{
		kind:  tokError,
		value: fmt.Sprintf(format, args...),
	}
	return nil
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) accept(set string) bool {
	if strings.ContainsRune(set, l.next()) {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(set string) {
	for strings.ContainsRune(set, l.next()) {
	}
	l.backup()
}

func lexToplevel(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof:
			l.emit(tokEOF)
			return nil
		case isSpace(r) || isEOL(r):
			l.backup()
			skipWhite(l)
		case r == ';':
			l.emit(tokSemi)
		case r == ',':
			l.emit(tokComma)
		case r == '(':
			l.emit(tokLparen)
		case r == ')':
			l.emit(tokRparen)
		case isNumeric(r):
			l.backup()
			return lexNumber
		case isAlpha(r):
			l.backup()
			return lexIdentifier
		case op[r] > tokUserBinaryOp:
			l.emit(op[r])
		case l.userOperators[r] == uopBinaryOp:
			l.emit(tokUserBinaryOp)
		case l.userOperators[r] == uopUnaryOp:
			l.emit(tokUserUnaryOp)
		default:
			return l.errorf("unrecognized character: %#U", r)
		}
	}
}

func lexNumber(l *lexer) stateFn {
	digits := "0123456789"
	if l.accept("0") && l.accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789")
	}
	if isAlphaNumeric(l.peek()) {
		return l.errorf("bad number syntax: %q", l.word()+string(l.peek()))
	}
	l.emit(tokNumber)
	return lexToplevel
}

func lexIdentifier(l *lexer) stateFn {
	for isAlphaNumeric(l.next()) {
	}
	l.backup()
	word := l.word()
	if tok, ok := keywords[word]; ok {
		l.emit(tok)
	} else {
		l.emit(tokIdentifier)
	}
	return lexToplevel
}

// lexing helper functions {{{

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isEOL(r rune) bool {
	return r == '\n' || r == '\r'
}

func skipWhite(l *lexer) {
	r := l.next()
	for isSpace(r) || isEOL(r) {
		r = l.next()
	}
	l.backup()
	l.ignore()
}

func isNumeric(r rune) bool {
	return '0' <= r && r <= '9'
}

func isAlpha(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isAlphaNumeric(r rune) bool {
	return isNumeric(r) || isAlpha(r)
}

// }}}
