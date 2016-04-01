package parse

import (
	"fmt"
	"strconv"

	"github.com/agatan/kaleigo/ast"
)

// Parser holds parsing info.
type Parser struct {
	lex       *lexer
	lookahead [3]token
	peekCount int
}

func (p *Parser) next() token {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.lookahead[0] = p.lex.nextToken()
	}
	return p.lookahead[p.peekCount]
}

func (p *Parser) backup() {
	p.peekCount--
}

func (p *Parser) backup2(t1 token) {
	p.lookahead[1] = t1
	p.peekCount = 2
}

func (p *Parser) backup3(t2, t1 token) {
	p.lookahead[1] = t1
	p.lookahead[2] = t2
	p.peekCount = 3
}

func (p *Parser) peek() token {
	if p.peekCount > 0 {
		return p.lookahead[p.peekCount-1]
	}
	p.peekCount = 1
	p.lookahead[0] = p.lex.nextToken()
	return p.lookahead[0]
}

func (p *Parser) errorf(format string, args ...interface{}) {
	panic(fmt.Errorf(format, args...))
}

func (p *Parser) error(err error) {
	panic(err)
}

func (p *Parser) parseNumber() ast.Expr {
	val, err := strconv.ParseFloat(p.peek().value, 64)
	if err != nil {
		p.error(err)
	}
	p.next()
	return &ast.NumberExpr{Val: val}
}
