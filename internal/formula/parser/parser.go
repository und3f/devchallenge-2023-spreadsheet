package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
	"text/scanner"
	"unicode"
)

func ParseExpr(src string, name string) (expr ast.Expr, err error) {
	var s scanner.Scanner

	s.Init(strings.NewReader(src))
	s.Filename = "percent"
	s.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats

	s.IsIdentRune = func(ch rune, i int) bool {
		return unicode.IsLetter(ch) || i > 0 &&
			(unicode.IsDigit(ch) || ch == '%' || ch == '_')
	}

	p := &Parser{
		scanner: s,
	}
	p.next()

	expr = p.parseExpr()
	err = p.err

	return
}

type Parser struct {
	scanner scanner.Scanner
	tok     token.Token
	err     error
}

func (p *Parser) parseExpr() ast.Expr {
	return p.parseBinaryExpr(nil, token.LowestPrec+1)
}

func (p *Parser) parseBinaryExpr(x ast.Expr, prec1 int) ast.Expr {
	if x == nil {
		x = p.parseUnaryExpr()
	}

	var n int
	for n = 1; ; n++ {
		op := p.tok
		oprec := op.Precedence()

		if oprec < prec1 {
			return x
		}

		p.next()

		y := p.parseBinaryExpr(nil, oprec+1)
		x = &ast.BinaryExpr{X: x, Y: y, Op: op}
	}
}

func (p *Parser) parseUnaryExpr() ast.Expr {
	switch p.tok {
	case token.ADD, token.SUB:
		op := p.tok
		p.next()
		x := p.parseUnaryExpr()
		return &ast.UnaryExpr{X: x, Op: op}
	}

	return p.parsePrimaryExpr(nil)
}

func (p *Parser) parsePrimaryExpr(x ast.Expr) ast.Expr {
	if x == nil {
		x = p.parseOperand()
	}
	/*
		switch (p.tok) {
			case token.IDENT
		}
	*/
	return x
}

func (p *Parser) parseOperand() ast.Expr {
	switch p.tok {
	case token.IDENT:
		expr := &ast.Ident{Name: p.scanner.TokenText()}
		p.next()
		return expr

	case token.INT, token.FLOAT:
		x := &ast.BasicLit{Kind: p.tok, Value: p.scanner.TokenText()}
		p.next()
		return x

	case token.LPAREN:
		p.next()
		x := p.parseExpr()
		p.expect(token.RPAREN)
		return &ast.ParenExpr{X: x}
	}

	return &ast.BadExpr{}

}

var tokenConverter map[rune]token.Token = map[rune]token.Token{
	scanner.EOF:   token.EOF,
	scanner.Ident: token.IDENT,
	scanner.Int:   token.INT,
	scanner.Float: token.FLOAT,
	'+':           token.ADD,
	'-':           token.SUB,
	'*':           token.MUL,
	'/':           token.QUO,
	'(':           token.LPAREN,
	')':           token.RPAREN,
}

func (p *Parser) error(err error) {
	if p.err != nil {
		return
	}

	p.err = err
}

func (p *Parser) expect(tok token.Token) {
	if p.tok != tok {
		p.error(fmt.Errorf("expected %s, found %s", tok.String(), p.tok.String()))
	}
	p.next()
}

func (p *Parser) next() {
	r := p.scanner.Scan()

	if tok, exists := tokenConverter[r]; exists {
		p.tok = tok
	} else {
		p.error(fmt.Errorf("Unexpected rune: %q", r))
	}
}
