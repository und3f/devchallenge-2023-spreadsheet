package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"devchallenge.it/spreadsheet/internal/model"
	"github.com/go-redis/redismock/v9"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type TestContext struct {
	service *Service
	router  *mux.Router
	dao     *model.Dao
	mock    redismock.ClientMock
}

func NewTestContext() *TestContext {
	r := mux.NewRouter()

	rdb, mock := redismock.NewClientMock()
	dao := model.NewDao(rdb)

	s := NewService(r, dao)

	return &TestContext{
		service: s,
		router:  r,
		dao:     dao,
		mock:    mock,
	}
}

func CreateUpsertPayload(value string) *bytes.Reader {
	jsonBody, _ := json.Marshal(UpsertPayload{value})
	return bytes.NewReader(jsonBody)
}

func TestUpsertSimpleVarSuccess(t *testing.T) {
	tctx := NewTestContext()

	tctx.mock.
		ExpectHSet(
			"devchallenge-xx",
			map[string]string{
				"var1": "0",
			},
		).SetVal(1)

	request, _ := http.NewRequest(
		http.MethodPost,
		"/devchallenge-xx/var1",
		CreateUpsertPayload("0"),
	)
	response := httptest.NewRecorder()

	tctx.router.ServeHTTP(response, request)

	var resp CellResponse
	json.NewDecoder(response.Body).Decode(&resp)

	assert.Equal(t, 201, response.Code)
	assert.Equal(t, "0", resp.Value)
	assert.Equal(t, "0", resp.Result)

	assert.NoError(t, tctx.mock.ExpectationsWereMet())
}

func TestUpsertInvalidCellIdFail(t *testing.T) {
	tctx := NewTestContext()

	request, _ := http.NewRequest(
		http.MethodPost,
		"/devchallenge-xx/var%bc",
		CreateUpsertPayload("0"),
	)
	response := httptest.NewRecorder()

	tctx.router.ServeHTTP(response, request)

	var resp CellResponse
	json.NewDecoder(response.Body).Decode(&resp)

	assert.Equal(t, http.StatusBadRequest, response.Code)

	assert.NoError(t, tctx.mock.ExpectationsWereMet())
}

func TestUpsertCaseInsensitiveSuccess(t *testing.T) {
	tctx := NewTestContext()

	tctx.mock.
		ExpectHSet(
			"devchallenge-xx",
			map[string]string{
				"var2": "1",
			},
		).SetVal(1)

	request, _ := http.NewRequest(
		http.MethodPost,
		"/DevChallenge-XX/VAR2",
		CreateUpsertPayload("1"),
	)
	response := httptest.NewRecorder()

	tctx.router.ServeHTTP(response, request)

	var resp CellResponse
	json.NewDecoder(response.Body).Decode(&resp)

	assert.Equal(t, 201, response.Code)
	assert.Equal(t, "1", resp.Value)
	assert.Equal(t, "1", resp.Result)

	assert.NoError(t, tctx.mock.ExpectationsWereMet())
}

func TestUpsertFormulaSuccess(t *testing.T) {
	tctx := NewTestContext()

	tctx.mock.ExpectHGet("devchallenge-xx", "var1").SetVal("1")
	tctx.mock.ExpectHGet("devchallenge-xx", "var2").SetVal("2")

	tctx.mock.
		ExpectHSet(
			"devchallenge-xx",
			map[string]string{
				"var3": "=var1+var2",
			},
		).SetVal(1)

	request, _ := http.NewRequest(
		http.MethodPost,
		"/devchallenge-xx/var3",
		CreateUpsertPayload("=var1+var2"),
	)
	response := httptest.NewRecorder()

	tctx.router.ServeHTTP(response, request)

	var resp CellResponse
	json.NewDecoder(response.Body).Decode(&resp)

	assert.Equal(t, 201, response.Code)
	assert.Equal(t, "=var1+var2", resp.Value)
	assert.Equal(t, "3", resp.Result)

	assert.NoError(t, tctx.mock.ExpectationsWereMet())
}

func TestPostFormulaError(t *testing.T) {
	tctx := NewTestContext()

	tctx.mock.ExpectHGet("devchallenge-xx", "var2").SetVal("2")

	request, _ := http.NewRequest(
		http.MethodPost,
		"/devchallenge-xx/var1",
		CreateUpsertPayload("=var2+var1"),
	)
	response := httptest.NewRecorder()

	tctx.router.ServeHTTP(response, request)

	var resp CellResponse
	json.NewDecoder(response.Body).Decode(&resp)

	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	assert.Equal(t, "=var2+var1", resp.Value)
	assert.Equal(t, "ERROR", resp.Result)

	assert.NoError(t, tctx.mock.ExpectationsWereMet())
}
