package formula

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

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
	case *ast.BadExpr:
		return nil, fmt.Errorf("Expression parsing failed")
	}

	return nil, fmt.Errorf("Expression %T not supported", n)
}

func (s *Solver) evalBinOperator(litX, litY *ast.BasicLit, op token.Token) (*ast.BasicLit, error) {
	kind := token.INT

	if litX.Kind == token.STRING || litY.Kind == token.STRING {
		kind = token.STRING
	} else if litX.Kind == token.FLOAT || litY.Kind == token.FLOAT {
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
	x, err := strconv.ParseInt(litX.Value, 10, 64)
	if err != nil {
		return nil, err
	}

	y, err := strconv.ParseInt(litY.Value, 10, 64)
	if err != nil {
		return nil, err
	}

	r, err := evalNumericOperator(op, x, y)
	if err != nil {
		return nil, err
	}

	return &ast.BasicLit{
		Value: strconv.FormatInt(r, 10),
		Kind:  token.INT,
	}, nil
}

func (s *Solver) evalFloatBinOperator(op token.Token, litX, litY *ast.BasicLit) (*ast.BasicLit, error) {
	x, err := strconv.ParseFloat(litX.Value, 64)
	if err != nil {
		return nil, err
	}

	y, err := strconv.ParseFloat(litY.Value, 64)
	if err != nil {
		return nil, err
	}

	r, err := evalNumericOperator(op, x, y)
	if err != nil {
		return nil, err
	}

	return &ast.BasicLit{
		Value: strconv.FormatFloat(float64(r), 'f', -1, 64),
		Kind:  token.FLOAT,
	}, nil
}

func evalNumericOperator[V int64 | float64](op token.Token, x V, y V) (V, error) {
	var r V

	switch op {
	case token.ADD:
		r = x + y
	case token.SUB:
		r = x - y
	case token.MUL:
		r = x * y
	case token.QUO:
		if y == 0 {
			return r, errors.New("division by zero")
		}
		r = x / y
	default:
		return r, errors.New("operation not supported")
	}

	return r, nil
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
