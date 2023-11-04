package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscribeCell(t *testing.T) {
	tctx := NewTestContext()

	tctx.mock.ExpectIncr("subscription:counter").SetVal(10)
	tctx.mock.ExpectHSet("subscription:a", map[string]string{
		"spreadsheetId": "devchallenge-xx",
		"cellId":        "var1",
	}).SetVal(1)

	request, _ := http.NewRequest(http.MethodPost, "/devchallenge-xx/var1/subscribe", nil)
	response := httptest.NewRecorder()
	tctx.router.ServeHTTP(response, request)

	var resp SubsribeResponse
	json.NewDecoder(response.Body).Decode(&resp)

	assert.Equal(t, 201, response.Code)

	assert.Equal(t, SubsribeResponse{
		WebhookUrl: "http:///sub/a",
	}, resp)

	assert.NoError(t, tctx.mock.ExpectationsWereMet())
}
