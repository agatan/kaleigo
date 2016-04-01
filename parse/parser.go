package parse

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/agatan/kaleigo/ast"
)

// Parser holds parsing info.
type Parser struct {
	lex          *lexer
	lookahead    [3]token
	peekCount    int
	binaryOpPrec map[rune]int
}

// New creates a new parser.
func New(name, input string) *Parser {
	binop := map[rune]int{
		'<': 10,
		'+': 20,
		'-': 20,
		'*': 20,
	}
	return &Parser{
		lex:          lex(name, input),
		binaryOpPrec: binop,
	}
}

type ToplevelType int

const (
	ToplevelExtern ToplevelType = iota
	ToplevelDef
	ToplevelOther
)

func (p *Parser) next() token {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.lookahead[0] = p.lex.nextToken()
	}
	return p.lookahead[p.peekCount]
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

func (p *Parser) tokenPrecedence(token rune) int {
	if prec, ok := p.binaryOpPrec[token]; ok {
		return prec
	}
	return -1
}

func (p *Parser) ToplevelKind() ToplevelType {
	switch p.peek().kind {
	case tokDef:
		return ToplevelDef
	case tokExtern:
		return ToplevelExtern
	default:
		return ToplevelOther
	}
}

// ParseTopLevelExpr consumes a expression and wrap it with anonymous function.
func (p *Parser) ParseTopLevelExpr() *ast.Function {
	expr := p.parseExpression()
	proto := &ast.Prototype{}
	return &ast.Function{Prototype: proto, Body: expr}
}

// ParseDefinition consumes a function definition.
func (p *Parser) ParseDefinition() *ast.Function {
	p.next()
	proto := p.parsePrototype()
	body := p.parseExpression()
	return &ast.Function{Prototype: proto, Body: body}
}

// ParseExtern consumes a external function declaration.
func (p *Parser) ParseExtern() *ast.Prototype {
	p.next()
	return p.parsePrototype()
}

func (p *Parser) parsePrototype() *ast.Prototype {
	name := p.peek().value
	p.next()
	if p.peek().kind != tokLparen {
		p.errorf("unexpected token: %q", p.peek().value)
	}
	p.next()

	args := []string{}
	if p.peek().kind != tokRparen {
		for {
			if p.peek().kind != tokIdentifier {
				p.errorf("unexpected token: %q", p.peek().value)
			}
			args = append(args, p.peek().value)
			p.next()

			if p.peek().kind == tokRparen {
				break
			}
			if p.peek().kind != tokComma {
				p.errorf("expected ','")
			}
			p.next()
		}
	}
	p.next()
	return &ast.Prototype{Name: name, Args: args}
}

func (p *Parser) parseExpression() ast.Expr {
	lhs := p.parsePrimary()
	return p.parseBinOpRHS(0, lhs)
}

func (p *Parser) parseBinOpRHS(prec int, lhs ast.Expr) ast.Expr {
	for {
		op, _ := utf8.DecodeRuneInString(p.peek().value)
		cprec := p.tokenPrecedence(op)
		if cprec < prec {
			return lhs
		}
		// skip op
		p.next()
		rhs := p.parsePrimary()
		op2, _ := utf8.DecodeRuneInString(p.peek().value)
		nprec := p.tokenPrecedence(op2)
		if prec < nprec {
			rhs = p.parseBinOpRHS(prec+1, rhs)
		}
		lhs = &ast.BinaryExpr{Op: op, LHS: lhs, RHS: rhs}
	}
}

func (p *Parser) parsePrimary() ast.Expr {
	switch p.peek().kind {
	case tokIdentifier:
		return p.parseIdentifier()
	case tokNumber:
		return p.parseNumber()
	case tokLparen:
		return p.parseParenExpr()
	}
	p.errorf("unexpected token: %q", p.peek().value)
	return nil
}

func (p *Parser) parseNumber() ast.Expr {
	val, err := strconv.ParseFloat(p.peek().value, 64)
	if err != nil {
		p.error(err)
	}
	p.next()
	return &ast.NumberExpr{Val: val}
}

func (p *Parser) parseIdentifier() ast.Expr {
	name := p.peek().value
	p.next()
	if p.peek().kind != tokLparen {
		return &ast.VariableExpr{Name: name}
	}
	// skip '('
	p.next()
	args := []ast.Expr{}
	if p.peek().kind != tokRparen {
		for {
			args = append(args, p.parseExpression())
			if p.peek().kind == tokRparen {
				break
			}
			if p.peek().kind != tokComma {
				p.error(fmt.Errorf("expected ','"))
			}
			// skip ','
			p.next()
		}
	}
	// skip ')'
	p.next()
	return &ast.CallExpr{Callee: name, Args: args}
}

func (p *Parser) parseParenExpr() ast.Expr {
	// skip '('
	p.next()
	expr := p.parseExpression()
	if p.peek().kind != tokRparen {
		p.error(fmt.Errorf("expected ')'"))
	}
	p.next()
	return expr
}
