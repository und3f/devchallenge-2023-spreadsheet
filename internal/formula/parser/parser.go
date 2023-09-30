package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"text/scanner"
)

func ParseExpr(src string, name string) (expr ast.Expr, err error) {
	p := &Parser{
		scanner: NewScanner(src, name),
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
		x := p.parseOperand()
		return &ast.UnaryExpr{X: x, Op: op}
	}

	return p.parseOperand()
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

	p.error(fmt.Errorf("Unexpected operand %s", p.tok.String()))
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
		p.error(fmt.Errorf("Unexpected character occured: %q", r))
	}
}
