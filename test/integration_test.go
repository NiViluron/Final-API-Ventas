package tests

import (
	"Final-API-Ventas/api"
	"Final-API-Ventas/internal/sale"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestIntegrationCreateAndGet(t *testing.T) {
	mockHandler := http.NewServeMux()

	mockHandler.HandleFunc("/users/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`test`))
	})

	mockServer := httptest.NewServer(mockHandler)
	defer mockServer.Close()

	fmt.Println("URL del mock server ", mockServer.URL)

	app := gin.Default()
	api.InitRoutes(app, mockServer.URL)

	// Create a new sale
	req, _ := http.NewRequest(http.MethodPost, "/sales", bytes.NewBufferString(`{
		"user_id": "test",
		"amount": 100.0	
	}`))

	res := fakeRequest(app, req)

	require.NotNil(t, res)
	require.Equal(t, http.StatusCreated, res.Code)

	var resSale *sale.Sale
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &resSale))
	require.Equal(t, "test", resSale.UserID)
	require.Equal(t, 100.0, resSale.Amount)
	require.Equal(t, 1, resSale.Version)
	require.NotEmpty(t, resSale.Status)
	require.NotEmpty(t, resSale.ID)
	require.NotEmpty(t, resSale.CreateAt)
	require.NotEmpty(t, resSale.UpdateAt)

	if resSale.Status == "pending" {
		req, _ = http.NewRequest(http.MethodPatch, "/sales/"+resSale.ID, bytes.NewBufferString(`{
			"status":"approved"
		}`))

		res = fakeRequest(app, req)
		require.NotNil(t, res)
		require.Equal(t, http.StatusOK, res.Code)

		var resSale2 *sale.Sale
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &resSale2))
		require.Equal(t, resSale.ID, resSale2.ID)
		require.Equal(t, "approved", resSale2.Status)
		require.Equal(t, 2, resSale2.Version)
		require.Equal(t, resSale.UserID, resSale2.UserID)
		require.Equal(t, resSale.Amount, resSale2.Amount)
		require.Equal(t, resSale.CreateAt, resSale2.CreateAt)
		require.NotEmpty(t, resSale2.UpdateAt)
	}

	// Get the sale
	req, _ = http.NewRequest(http.MethodGet, "/sales?user_id="+resSale.UserID+"&status=approved", nil)
	res = fakeRequest(app, req)
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)

	/*var resSales []*sale.Sale
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &resSales))

	for _, s := range resSales {
		require.Equal(t, resSale.UserID, s.UserID)
		require.Equal(t, "approved", s.Status)
	}*/
}

func fakeRequest(e *gin.Engine, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)

	return w
}
