package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
)

func TestGetCell(t *testing.T) {
	tctx := NewTestContext()

	tctx.mock.ExpectHGet("devchallenge-xx", "var1").SetVal("1")

	request, _ := http.NewRequest(http.MethodGet, "/devchallenge-xx/var1", nil)
	response := httptest.NewRecorder()
	tctx.router.ServeHTTP(response, request)

	var resp CellResponse
	json.NewDecoder(response.Body).Decode(&resp)

	assert.Equal(t, 200, response.Code)

	wantResp := CellResponse{
		Value:  "1",
		Result: "1",
	}

	if diff := deep.Equal(resp, wantResp); diff != nil {
		t.Error(diff)
	}

	assert.NoError(t, tctx.mock.ExpectationsWereMet())
}

func TestGetCellDoesntExists(t *testing.T) {
	tctx := NewTestContext()

	tctx.mock.ExpectHGet("devchallenge-xx", "var2").RedisNil()

	request, _ := http.NewRequest(http.MethodGet, "/devchallenge-xx/var2", nil)
	response := httptest.NewRecorder()
	tctx.router.ServeHTTP(response, request)

	var resp CellResponse
	json.NewDecoder(response.Body).Decode(&resp)

	assert.Equal(t, 404, response.Code)
	assert.NoError(t, tctx.mock.ExpectationsWereMet())
}

func TestGetCellInvalidCellId(t *testing.T) {
	tctx := NewTestContext()

	request, _ := http.NewRequest(http.MethodGet, "/devchallenge-xx/var+bc", nil)
	response := httptest.NewRecorder()
	tctx.router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestGetCellComplexName(t *testing.T) {
	tctx := NewTestContext()

	tctx.mock.ExpectHGet("devchallenge-xx", "說").SetVal("=á._+拿")
	tctx.mock.ExpectHGet("devchallenge-xx", "á._").SetVal("3")
	tctx.mock.ExpectHGet("devchallenge-xx", "拿").SetVal("2")

	request, _ := http.NewRequest(http.MethodGet, "/devchallenge-xx/說", nil)
	response := httptest.NewRecorder()
	tctx.router.ServeHTTP(response, request)

	var resp CellResponse
	json.NewDecoder(response.Body).Decode(&resp)

	assert.Equal(t, 200, response.Code)

	wantResp := CellResponse{
		Value:  "=á._+拿",
		Result: "5",
	}

	if diff := deep.Equal(resp, wantResp); diff != nil {
		t.Error(diff)
	}

	assert.NoError(t, tctx.mock.ExpectationsWereMet())
}
