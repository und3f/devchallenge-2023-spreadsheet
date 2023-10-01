package formula

import (
	"testing"

	"devchallenge.it/spreadsheet/internal/model"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func prepare() (*model.Dao, redismock.ClientMock) {
	rdb, mock := redismock.NewClientMock()
	dao := model.NewDao(rdb)

	return dao, mock
}

func TestRecursiveFormula(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	dao := model.NewDao(rdb)

	mock.ExpectHGet("devchallenge-xx", "var1").SetVal("=var2")
	mock.ExpectHGet("devchallenge-xx", "var2").SetVal("=var3")
	mock.ExpectHGet("devchallenge-xx", "var3").SetVal("=1")

	solver := NewSolver(dao, "devchallenge-xx")

	result, _, _, err := solver.Solve("var1")

	assert.NoError(t, err)
	assert.Equal(t, "1", result)
}

func TestCaseInsensitive(t *testing.T) {
	dao, mock := prepare()

	mock.ExpectHGet("devchallenge-xx", "var3").SetVal("=vAr1+VAR2")
	mock.ExpectHGet("devchallenge-xx", "var1").SetVal("1")
	mock.ExpectHGet("devchallenge-xx", "var2").SetVal("2")

	solver := NewSolver(dao, "devchallenge-xx")

	result, value, _, err := solver.Solve("var3")
	assert.NoError(t, err)
	assert.Equal(t, "=vAr1+VAR2", value)
	assert.Equal(t, "3", result)
}

func TestCache(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	dao := model.NewDao(rdb)

	mock.ExpectHGet("devchallenge-xx", "var1").SetVal("=var2+var3+var2")
	mock.ExpectHGet("devchallenge-xx", "var2").SetVal("=1")
	mock.ExpectHGet("devchallenge-xx", "var3").SetVal("=var2")

	solver := NewSolver(dao, "devchallenge-xx")

	result, _, _, err := solver.Solve("var1")

	assert.NoError(t, err)
	assert.Equal(t, "3", result)
}

func TestCycleDependency(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	dao := model.NewDao(rdb)

	mock.ExpectHGet("devchallenge-xx", "var1").SetVal("=var2+var3")
	mock.ExpectHGet("devchallenge-xx", "var2").SetVal("=var3")
	mock.ExpectHGet("devchallenge-xx", "var3").SetVal("=var1")

	solver := NewSolver(dao, "devchallenge-xx")

	result, _, formulaError, err := solver.Solve("var1")

	assert.NoError(t, err)
	if assert.Error(t, formulaError) {
		assert.Equal(t, CYCLE_DEPENDECY_ERROR, formulaError)
	}
	assert.Equal(t, ERROR, result)
}

func TestCycleDependency2(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	dao := model.NewDao(rdb)

	mock.ExpectHGet("devchallenge-xx", "var1").SetVal("=var1")

	solver := NewSolver(dao, "devchallenge-xx")

	result, _, formulaError, err := solver.Solve("var1")

	assert.NoError(t, err)
	if assert.Error(t, formulaError) {
		assert.Equal(t, CYCLE_DEPENDECY_ERROR, formulaError)
	}
	assert.Equal(t, "ERROR", result)
}

func TestInvalidFormula(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	dao := model.NewDao(rdb)

	mock.ExpectHGet("devchallenge-xx", "var1").SetVal("=(1*2")

	solver := NewSolver(dao, "devchallenge-xx")

	result, _, formulaError, err := solver.Solve("var1")

	assert.NoError(t, err)
	assert.Error(t, formulaError)
	assert.Equal(t, "ERROR", result)
}

func TestEmptyFormula(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	dao := model.NewDao(rdb)

	mock.ExpectHGet("devchallenge-xx", "var1").SetVal("")

	solver := NewSolver(dao, "devchallenge-xx")

	result, _, formulaError, err := solver.Solve("var1")

	assert.NoError(t, err)
	assert.NoError(t, formulaError)
	assert.Equal(t, "", result)
}
