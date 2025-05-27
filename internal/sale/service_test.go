package sale

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"resty.dev/v3"
)

func TestService_Create_Simple(t *testing.T) {
	mockHandler := http.NewServeMux()

	mockHandler.HandleFunc("/users/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`test`))
	})

	mockServer := httptest.NewServer(mockHandler)
	defer mockServer.Close()

	s := NewService(NewLocalStorage(), nil, mockServer.URL)

	fmt.Println("URL del mock server ", mockServer.URL)

	client := resty.New()
	resTest, errTest := client.R().Get(fmt.Sprintf("%s/users", mockServer.URL))

	if errTest != nil {
		fmt.Printf("tuve un error %s", errTest.Error())
		return
	}

	fmt.Printf("respuesta obtenida del api call al mock %v", resTest.String())

	input := &Sale{
		UserID: resTest.String(),
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
