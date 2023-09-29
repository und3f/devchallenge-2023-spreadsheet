package formula

import (
	"strconv"
	"testing"

	"devchallenge.it/spreadsheet/internal/model"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestSimpleSovle(t *testing.T) {
	dao, mock := prepare()

	mock.ExpectHGet("devchallenge-xx", "var3").SetVal("=var1+var2")
	mock.ExpectHGet("devchallenge-xx", "var1").SetVal("1")
	mock.ExpectHGet("devchallenge-xx", "var2").SetVal("2")

	solver := NewSolver(dao, "devchallenge-xx")

	result, value, err := solver.Solve("var3")
	assert.NoError(t, err)
	assert.Equal(t, "=var1+var2", value)
	assert.Equal(t, "3", result)
}

func TestCaseInsensitive(t *testing.T) {
	dao, mock := prepare()

	mock.ExpectHGet("devchallenge-xx", "var3").SetVal("=vAr1+VAR2")
	mock.ExpectHGet("devchallenge-xx", "var1").SetVal("1")
	mock.ExpectHGet("devchallenge-xx", "var2").SetVal("2")

	solver := NewSolver(dao, "devchallenge-xx")

	result, value, err := solver.Solve("var3")
	assert.NoError(t, err)
	assert.Equal(t, "=vAr1+VAR2", value)
	assert.Equal(t, "3", result)
}

func TestAllOperations(t *testing.T) {
	dao, mock := prepare()

	mock.ExpectHGet("devchallenge-xx", "var5").SetVal("=var1+(var2*var3+var4)/2")
	mock.ExpectHGet("devchallenge-xx", "var1").SetVal("1")
	mock.ExpectHGet("devchallenge-xx", "var2").SetVal("2")
	mock.ExpectHGet("devchallenge-xx", "var3").SetVal("3")
	mock.ExpectHGet("devchallenge-xx", "var4").SetVal("4")

	solver := NewSolver(dao, "devchallenge-xx")

	result, _, err := solver.Solve("var5")
	assert.NoError(t, err)
	assert.Equal(t, "6", result)
}

func TestFloat(t *testing.T) {
	dao, mock := prepare()

	mock.ExpectHGet("devchallenge-xx", "var1").SetVal("=var2+var3")
	mock.ExpectHGet("devchallenge-xx", "var2").SetVal("1.1")
	mock.ExpectHGet("devchallenge-xx", "var3").SetVal("2.2")

	solver := NewSolver(dao, "devchallenge-xx")
	result, _, err := solver.Solve("var1")

	assert.NoError(t, err)
	resultF, err := strconv.ParseFloat(result, 32)
	assert.NoError(t, err)
	assert.InDelta(t, 3.3, resultF, 0.01)
}

func TestIntAndFloatResultsFloat(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	dao := model.NewDao(rdb)

	mock.ExpectHGet("devchallenge-xx", "var1").SetVal("=1+2.3")

	solver := NewSolver(dao, "devchallenge-xx")

	result, _, err := solver.Solve("var1")

	assert.NoError(t, err)
	resultF, err := strconv.ParseFloat(result, 32)
	assert.NoError(t, err)
	assert.InDelta(t, 3.3, resultF, 0.01)
}
