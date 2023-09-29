package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
)

func TestGetSpreadsheetExists(t *testing.T) {
	tctx := NewTestContext()

	tctx.mock.ExpectExists("devchallenge-xx").SetVal(1)
	tctx.mock.ExpectHKeys("devchallenge-xx").SetVal([]string{"var1"})
	tctx.mock.ExpectHGetAll("devchallenge-xx").SetVal(
		map[string]string{
			"var1": "1",
		})

	request, _ := http.NewRequest(http.MethodGet, "/devchallenge-xx", nil)
	response := httptest.NewRecorder()
	tctx.router.ServeHTTP(response, request)

	var resp SpreadsheetResponse
	json.NewDecoder(response.Body).Decode(&resp)

	assert.Equal(t, 200, response.Code)

	wantResp := SpreadsheetResponse{
		"var1": CellResponse{
			Value:  "1",
			Result: "1",
		},
	}
	if diff := deep.Equal(resp, wantResp); diff != nil {
		t.Error(diff)
	}

	assert.NoError(t, tctx.mock.ExpectationsWereMet())
}

func TestGetSpreadsheetDoesntExists(t *testing.T) {
	tctx := NewTestContext()

	tctx.mock.ExpectExists("devchallenge-xx").SetVal(0)

	request, _ := http.NewRequest(http.MethodGet, "/devchallenge-xx", nil)
	response := httptest.NewRecorder()
	tctx.router.ServeHTTP(response, request)

	var resp SpreadsheetResponse
	json.NewDecoder(response.Body).Decode(&resp)

	assert.Equal(t, 404, response.Code)
	assert.NoError(t, tctx.mock.ExpectationsWereMet())
}

func TestGetSpreadsheetPreload(t *testing.T) {
	tctx := NewTestContext()

	tctx.mock.ExpectExists("devchallenge-xx").SetVal(1)
	tctx.mock.ExpectHKeys("devchallenge-xx").SetVal(
		[]string{
			"var1", "var2", "var3", "var4"})
	tctx.mock.ExpectHGetAll("devchallenge-xx").SetVal(map[string]string{
		"var1": "=var2+var3",
		"var2": "=var3 + var3 - var3 - var3 + var4",
		"var3": "=var4-var4+0",
		"var4": "1",
	})

	request, _ := http.NewRequest(http.MethodGet, "/devchallenge-xx", nil)
	response := httptest.NewRecorder()
	tctx.router.ServeHTTP(response, request)

	var resp SpreadsheetResponse
	json.NewDecoder(response.Body).Decode(&resp)

	assert.Equal(t, 200, response.Code)

	wantResp := SpreadsheetResponse{
		"var1": CellResponse{
			Value:  "=var2+var3",
			Result: "1",
		},
		"var2": CellResponse{
			Value:  "=var3 + var3 - var3 - var3 + var4",
			Result: "1",
		},
		"var3": CellResponse{
			Value:  "=var4-var4+0",
			Result: "0",
		},
		"var4": CellResponse{
			Value:  "1",
			Result: "1",
		},
	}
	if diff := deep.Equal(resp, wantResp); diff != nil {
		t.Error(diff)
	}

	assert.NoError(t, tctx.mock.ExpectationsWereMet())
}
