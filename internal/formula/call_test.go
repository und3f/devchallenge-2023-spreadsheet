package formula

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinCall(t *testing.T) {
	dao, mock := prepare()
	mock.ExpectHGet("devchallenge-xx", "var3").SetVal("=MIN(var1, var2)")
	mock.ExpectHGet("devchallenge-xx", "var1").SetVal("1")
	mock.ExpectHGet("devchallenge-xx", "var2").SetVal("2")

	solver := NewSolver(dao, "devchallenge-xx")
	result, value, formulaError, err := solver.Solve("var3")

	assert.NoError(t, formulaError)
	assert.NoError(t, err)
	assert.Equal(t, "=MIN(var1, var2)", value)
	assert.Equal(t, "1", result)
}

func TestMaxCall(t *testing.T) {
	dao, mock := prepare()
	mock.ExpectHGet("devchallenge-xx", "var3").SetVal("=MAX(var1, var2)")
	mock.ExpectHGet("devchallenge-xx", "var1").SetVal("1")
	mock.ExpectHGet("devchallenge-xx", "var2").SetVal("2")

	solver := NewSolver(dao, "devchallenge-xx")
	result, value, formulaError, err := solver.Solve("var3")

	assert.NoError(t, formulaError)
	assert.NoError(t, err)
	assert.Equal(t, "=MAX(var1, var2)", value)
	assert.Equal(t, "2", result)
}

func TestSumCall(t *testing.T) {
	dao, mock := prepare()
	mock.ExpectHGet("devchallenge-xx", "var3").SetVal("=SUM(var1, var2)")
	mock.ExpectHGet("devchallenge-xx", "var1").SetVal("1")
	mock.ExpectHGet("devchallenge-xx", "var2").SetVal("2")

	solver := NewSolver(dao, "devchallenge-xx")
	result, value, formulaError, err := solver.Solve("var3")

	assert.NoError(t, formulaError)
	assert.NoError(t, err)
	assert.Equal(t, "=SUM(var1, var2)", value)
	assert.Equal(t, "3", result)
}

func TestAvgCall(t *testing.T) {
	dao, mock := prepare()
	mock.ExpectHGet("devchallenge-xx", "var3").SetVal("=AVG(var1, var2)")
	mock.ExpectHGet("devchallenge-xx", "var1").SetVal("1")
	mock.ExpectHGet("devchallenge-xx", "var2").SetVal("2")

	solver := NewSolver(dao, "devchallenge-xx")
	result, value, formulaError, err := solver.Solve("var3")

	assert.NoError(t, formulaError)
	assert.NoError(t, err)
	assert.Equal(t, "=AVG(var1, var2)", value)
	assert.Equal(t, "1.5", result)
}
