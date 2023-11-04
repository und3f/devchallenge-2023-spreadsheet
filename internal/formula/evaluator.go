package formula

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"math/big"

	"devchallenge.it/spreadsheet/internal/formula/parser"
)

func (s *Solver) evalNode(n ast.Node) (*ast.BasicLit, error) {
	switch nod := n.(type) {
	case *ast.Ident:
		return s.expandVariable(nod)
	case *ast.BasicLit:
		return nod, nil
	case *ast.ParenExpr:
		return s.evalNode(nod.X)
	case *ast.BinaryExpr:
		lit1, err := s.evalNode(nod.X)
		if err != nil {
			return nil, err
		}
		lit2, err := s.evalNode(nod.Y)
		if err != nil {
			return nil, err
		}
		return s.evalBinOperator(lit1, lit2, nod.Op)
	case *ast.UnaryExpr:
		lit2, err := s.evalNode(nod.X)
		if err != nil {
			return nil, err
		}
		lit1 := &ast.BasicLit{Value: "0", Kind: lit2.Kind}
		return s.evalBinOperator(lit1, lit2, nod.Op)
	case *ast.CallExpr:
		return s.evalCall(nod)
	case *ast.BadExpr:
		return nil, fmt.Errorf("Expression parsing failed")
	}

	return nil, fmt.Errorf("Expression %T not supported", n)
}

func (s *Solver) evalBinOperator(litX, litY *ast.BasicLit, op token.Token) (*ast.BasicLit, error) {
	kind := token.INT

	if litX.Kind == token.STRING || litY.Kind == token.STRING {
		kind = token.STRING
	} else if litX.Kind == token.FLOAT || litY.Kind == token.FLOAT || op == token.QUO {
		kind = token.FLOAT
	}

	switch kind {
	case token.INT:
		return s.evalIntBinOperator(op, litX, litY)
	case token.FLOAT:
		return s.evalFloatBinOperator(op, litX, litY)
	}

	return nil,
		fmt.Errorf("arithmetic operation not supported for type %s", kind.String())
}

func (s *Solver) evalIntBinOperator(op token.Token, litX, litY *ast.BasicLit) (*ast.BasicLit, error) {
	x, y := &big.Int{}, &big.Int{}
	if _, ok := x.SetString(litX.Value, 10); !ok {
		return nil, errors.New("int parsing failure")
	}

	if _, ok := y.SetString(litY.Value, 10); !ok {
		return nil, errors.New("int parsing failure")
	}

	switch op {
	case token.ADD:
		x.Add(x, y)
	case token.SUB:
		x.Sub(x, y)
	case token.MUL:
		x.Mul(x, y)
	case token.QUO:
		if y.Cmp(big.NewInt(0)) == 0 {
			return nil, errors.New("division by zero")
		}
		x.Div(x, y)
	default:
		return nil, errors.New("operation not supported")
	}

	return &ast.BasicLit{
		Value: x.Text(10),
		Kind:  token.INT,
	}, nil
}

func (s *Solver) evalFloatBinOperator(op token.Token, litX, litY *ast.BasicLit) (*ast.BasicLit, error) {
	x, y := &big.Float{}, &big.Float{}

	_, _, err := x.Parse(litX.Value, 10)
	if err != nil {
		return nil, err
	}

	_, _, err = y.Parse(litY.Value, 10)
	if err != nil {
		return nil, err
	}

	switch op {
	case token.ADD:
		x.Add(x, y)
	case token.SUB:
		x.Sub(x, y)
	case token.MUL:
		x.Mul(x, y)
	case token.QUO:
		if y.Cmp(big.NewFloat(0)) == 0 {
			return nil, errors.New("division by zero")
		}
		x.Quo(x, y)
	default:
		return nil, errors.New("operation not supported")
	}

	return &ast.BasicLit{
		Value: x.Text('f', -1),
		Kind:  token.FLOAT,
	}, nil
}

func (s *Solver) expandVariable(lit *ast.Ident) (*ast.BasicLit, error) {
	result, _, formulaErr, err := s.Solve(lit.Name)
	if err != nil {
		return nil, err
	} else if formulaErr != nil {
		return nil, formulaErr
	}

	resultLit := parser.ParseValue(result, lit.Name)
	basicLit, err := s.evalNode(resultLit)
	if err != nil {
		return nil, err
	}

	return basicLit, nil
}
