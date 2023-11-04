package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"text/scanner"
)

type Parser struct {
	scanner scanner.Scanner
	tok     token.Token
	err     error
}

func ParseExpr(src string, name string) (expr ast.Expr, err error) {
	p := &Parser{
		scanner: NewScanner(src, name),
	}
	p.next()

	expr = p.parseExpr()
	err = p.err

	p.expect(token.EOF)

	return
}

func ParseValue(src string, filename string) (lit ast.Expr) {
	p := &Parser{scanner: NewScanner(src, filename)}
	p.next()

	expr := p.parseUnaryExpr()
	p.expect(token.EOF)
	if p.err != nil {
		return createStringLit(src)
	}

	switch op := expr.(type) {
	case *ast.UnaryExpr:
		if lit, ok := op.X.(*ast.BasicLit); ok {
			if lit.Kind == token.INT || lit.Kind == token.FLOAT {
				return op
			}
		}
	case *ast.BasicLit:
		return op
	}

	return createStringLit(src)
}

func createStringLit(value string) *ast.BasicLit {
	return &ast.BasicLit{
		Value: value,
		Kind:  token.STRING,
	}
}

func (p *Parser) parseExpr() ast.Expr {
	expr := p.parseBinaryExpr(nil, token.LowestPrec+1)

	return expr
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

	return p.parsePrimaryExpr()
}

func (p *Parser) parsePrimaryExpr() ast.Expr {
	x := p.parseOperand()

	switch p.tok {
	case token.LPAREN:
		x = p.parseCall(x)
	}

	return x

}

func (p *Parser) parseCall(fun ast.Expr) ast.Expr {
	p.expect(token.LPAREN)

	var list []ast.Expr
	for p.tok != token.LPAREN && p.tok != token.EOF {
		list = append(list, p.parseExpr())

		if !(p.tok == token.COMMA) {
			break
		}
		p.next()
	}

	p.expect(token.RPAREN)

	return &ast.CallExpr{
		Fun:  fun,
		Args: list,
	}
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
	',':           token.COMMA,
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
