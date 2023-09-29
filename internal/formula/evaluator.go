package formula

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
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
	}

	return nil, fmt.Errorf("Expression %T not supported", n)
}

func (s *Solver) evalBinOperator(litX, litY *ast.BasicLit, op token.Token) (*ast.BasicLit, error) {
	kind := token.INT
	if litX.Kind == token.FLOAT || litY.Kind == token.FLOAT {
		kind = token.FLOAT
	}

	switch kind {
	case token.INT:
		return s.evalIntBinOperator(op, litX, litY)
	case token.FLOAT:
		return s.evalFloatBinOperator(op, litX, litY)
	}

	return nil,
		fmt.Errorf("binary operator not supported for type %s", litX.Kind.String())
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
		r = x / y
	default:
		return r, errors.New("Operation not supported")
	}

	return r, nil
}

func (s *Solver) expandVariable(lit *ast.Ident) (*ast.BasicLit, error) {
	result, _, formulaErr, err := s.solve(lit.Name)
	if err != nil {
		return nil, err
	} else if formulaErr != nil {
		return nil, formulaErr
	}

	expr, err := parser.ParseExpr(result)
	if err != nil {
		return createStringLit(result), nil
	}

	res, isLit := expr.(*ast.BasicLit)
	if !isLit || (res.Kind != token.INT && res.Kind != token.FLOAT) {
		return createStringLit(result), nil
	}

	return res, nil
}

func createStringLit(value string) *ast.BasicLit {
	return &ast.BasicLit{
		Value: value,
		Kind:  token.STRING,
	}
}
