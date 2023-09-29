package formula

import (
	"strings"
	"testing"

	"devchallenge.it/spreadsheet/internal/model"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestStrings(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	dao := model.NewDao(rdb)

	mock.ExpectHGet("devchallenge-xx", "var1").SetVal("=var2")
	mock.ExpectHGet("devchallenge-xx", "var2").SetVal("=((var3))")
	mock.ExpectHGet("devchallenge-xx", "var3").SetVal("abc!@_*.%á")

	solver := NewSolver(dao, "devchallenge-xx")

	result, _, _, err := solver.Solve("var1")

	assert.NoError(t, err)
	assert.Equal(t, "abc!@_*.%á", result)
}

func TestNoBinaryOperatorsForStrings(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	dao := model.NewDao(rdb)

	mock.ExpectHGet("devchallenge-xx", "var1").SetVal("=var2+1")
	mock.ExpectHGet("devchallenge-xx", "var2").SetVal("Some string")

	solver := NewSolver(dao, "devchallenge-xx")

	result, _, _, err := solver.Solve("var1")

	assert.NoError(t, err)
	assert.Equal(t, "ERROR", result)
}

const LATIN = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz01233456789"

func FuzzString(f *testing.F) {
	rdb, mock := redismock.NewClientMock()
	dao := model.NewDao(rdb)

	f.Add(LATIN)
	f.Add("最近有什么新鲜事吗？")
	f.Add(strings.Repeat(LATIN, 1000))

	f.Fuzz(func(t *testing.T, s string) {
		mock.ExpectHGet("devchallenge-xx", "var1").SetVal(s)

		solver := NewSolver(dao, "devchallenge-xx")

		result, _, _, err := solver.Solve("var1")

		if err != nil || result != s {
			t.Errorf("%q, %v", result, err)
		}
	})
}
