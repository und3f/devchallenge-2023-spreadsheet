package formula

import (
	"fmt"
	"go/ast"
	"go/token"
	"math"
	"math/big"
	"strconv"
	"strings"
)

type FormulaFun func(s *Solver, args []*ast.BasicLit) (*ast.BasicLit, error)

var formulaFunctions = map[string]FormulaFun{
	"SUM": evalSum,
	"AVG": evalAvg,
	"MIN": evalMin,
	"MAX": evalMax,
}

func (s *Solver) evalCall(call *ast.CallExpr) (*ast.BasicLit, error) {
	funIdent, ok := call.Fun.(*ast.Ident)
	if !ok {
		return nil, fmt.Errorf("Invalid function name literal type: %T", call.Fun)
	}
	funName := strings.ToUpper(funIdent.Name)

	fun, exists := formulaFunctions[funName]

	if !exists {
		return nil, fmt.Errorf("Unknown function %q", funName)
	}

	args := make([]*ast.BasicLit, len(call.Args))

	for i := range call.Args {
		lit, err := s.evalNode(call.Args[i])
		if err != nil {
			return nil, err
		}

		args[i] = lit
	}

	return fun(s, args)
}

func parseLitFloatValue(lit *ast.BasicLit) (*big.Float, error) {
	v := &big.Float{}
	if _, _, err := v.Parse(lit.Value, 10); err != nil {
		return nil, err
	}

	return v, nil
}

func evalSum(s *Solver, args []*ast.BasicLit) (*ast.BasicLit, error) {
	sum := args[0]
	for _, lit := range args[1:] {
		newSum, err := s.evalBinOperator(sum, lit, token.ADD)
		if err != nil {
			return nil, err
		}
		sum = newSum
	}

	return sum, nil
}

func evalAvg(s *Solver, args []*ast.BasicLit) (*ast.BasicLit, error) {
	sum, err := evalSum(s, args)
	if err != nil {
		return nil, err
	}

	n := len(args)

	return s.evalBinOperator(sum, &ast.BasicLit{
		Kind:  token.INT,
		Value: strconv.Itoa(n),
	}, token.QUO)
}

func evalMin(s *Solver, args []*ast.BasicLit) (*ast.BasicLit, error) {
	minI := -1
	minVal := big.NewFloat(math.Inf(1))

	for i := range args {
		v, err := parseLitFloatValue(args[i])
		if err != nil {
			return nil, err
		}
		if v.Cmp(minVal) < 0 {
			minI = i
			minVal = v
		}

	}

	return args[minI], nil
}

func evalMax(s *Solver, args []*ast.BasicLit) (*ast.BasicLit, error) {
	maxI := -1
	maxVal := big.NewFloat(math.Inf(-1))

	for i := range args {
		v, err := parseLitFloatValue(args[i])
		if err != nil {
			return nil, err
		}
		if v.Cmp(maxVal) > 0 {
			maxI = i
			maxVal = v
		}

	}

	return args[maxI], nil
}
