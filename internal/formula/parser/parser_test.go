package parser

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCall(t *testing.T) {
	tree, formulaError := ParseExpr("min(1, 2, 3) + 1", "test")

	assert.NoError(t, formulaError)
	assert.Equal(t, &ast.BinaryExpr{
		Op: token.ADD,
		X: &ast.CallExpr{
			Fun: &ast.Ident{
				Name: "min",
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.INT, Value: "1"},
				&ast.BasicLit{Kind: token.INT, Value: "2"},
				&ast.BasicLit{Kind: token.INT, Value: "3"},
			},
		},
		Y: &ast.BasicLit{Kind: token.INT, Value: "1"},
	}, tree)
}

func TestParseCall2(t *testing.T) {
	tree, formulaError := ParseExpr("MIN(var1, var2)", "test")

	assert.NoError(t, formulaError)
	assert.Equal(t, &ast.CallExpr{
		Fun: &ast.Ident{
			Name: "MIN",
		},
		Args: []ast.Expr{
			&ast.Ident{Name: "var1"},
			&ast.Ident{Name: "var2"},
		},
	}, tree)
}
