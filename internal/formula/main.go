package formula

import (
	"errors"
	"strings"

	"devchallenge.it/spreadsheet/internal/formula/parser"
	"devchallenge.it/spreadsheet/internal/model"
	"github.com/redis/go-redis/v9"
)

const ERROR = "ERROR"

var CYCLE_DEPENDECY_ERROR = errors.New("Cycle dependency")
var NO_SUCH_CELL = errors.New("No such cellId")

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
		if err == redis.Nil {
			err = nil
			formulaError = NO_SUCH_CELL
		}
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
		return ERROR, value, CYCLE_DEPENDECY_ERROR, err
	}

	s.visited[cellId] = struct{}{}

	tr, formulaError := parser.ParseExpr(value[1:], cellId)
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
	return len(value) > 0 && value[0] == '='
}
