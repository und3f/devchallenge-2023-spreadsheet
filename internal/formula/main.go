package formula

import (
	"errors"
	"go/ast"
	"go/parser"
	"strings"

	"devchallenge.it/spreadsheet/internal/model"
)

const ERROR = "ERROR"

type Solver struct {
	dao         *model.Dao
	spreadsheet string

	visited map[string]struct{}
	values  map[string]string
	cache   map[string]string
}

func NewSolver(dao *model.Dao, spreadsheet string) *Solver {
	return &Solver{
		dao:         dao,
		spreadsheet: spreadsheet,

		visited: make(map[string]struct{}),
		values:  make(map[string]string),
		cache:   make(map[string]string),
	}
}

func (s *Solver) LoadAllKeys() (err error) {
	data, err := s.dao.GetAllCells(s.spreadsheet)
	if err != nil {
		return err
	}

	for cellId, value := range data {
		s.values[cellId] = value
	}

	return nil
}

func (s *Solver) SetCell(cellId string, value string) {
	cellId = strings.ToLower(cellId)
	s.values[cellId] = value
}

func (s *Solver) Solve(cellId string) (result string, value string, formulaError error, err error) {
	cellId = strings.ToLower(cellId)

	value, err = s.getValue(cellId)
	if err != nil {
		return
	}

	if !IsFormula(value) {
		result = value
		return
	}

	if result, exists := s.cache[cellId]; exists {
		return result, value, nil, nil
	}

	if _, exists := s.visited[cellId]; exists {
		return ERROR, value, errors.New("Cycle dependency"), err
	}

	s.visited[cellId] = struct{}{}

	tr, formulaError := parser.ParseExpr(value[1:])
	if formulaError != nil {
		result = ERROR
		return
	}

	resultLit, formulaError := s.evalNode(tr)
	if formulaError != nil {
		result = ERROR
		return
	}

	result = resultLit.Value

	s.cache[cellId] = result

	return
}

func (s *Solver) getValue(cellId string) (string, error) {
	cellId = strings.ToLower(cellId)
	if value, exists := s.values[cellId]; exists {
		return value, nil
	}

	value, err := s.dao.GetCell(s.spreadsheet, cellId)
	if err != nil {
		return "", err
	}

	s.values[cellId] = value
	return value, nil
}

func IsFormula(value string) bool {
	return value[0] == '='
}

// TODO: support URI compliant variables
func IsVariable(value string) bool {
	tr, err := parser.ParseExpr(value)
	if err != nil {
		return false
	}

	_, ok := tr.(*ast.Ident)

	return ok
}
