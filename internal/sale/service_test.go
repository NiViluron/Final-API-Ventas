package sale

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_Create_Simple(t *testing.T) {
	mockHandler := http.NewServeMux()

	mockHandler.HandleFunc("/users/123", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id": "123"}`))
	})

	mockServer := httptest.NewServer(mockHandler)
	defer mockServer.Close()

	s := NewService(NewLocalStorage(), nil, mockServer.URL)

	input := &Sale{
		UserID: "123",
		Amount: 100.0,
	}

	s = NewService(&mockStorage{
		mockSet: func(sale *Sale) error {
			return errors.New("fake error trying to set sale")
		},
	}, nil, mockServer.URL)

	err := s.Create(input)
	require.NotNil(t, err)
	require.EqualError(t, err, "user not found")
}

type mockStorage struct {
	mockSet      func(sale *Sale) error
	mockRead     func(id string) (*Sale, error)
	mockGetSales func(id string, status string) ([]*Sale, error)
}

func (m *mockStorage) Set(sale *Sale) error {
	return m.mockSet(sale)
}

func (m *mockStorage) Read(id string) (*Sale, error) {
	return m.mockRead(id)
}

func (m *mockStorage) GetSales(user_id string, status string) ([]*Sale, error) {
	return m.mockGetSales(user_id, status)
}
