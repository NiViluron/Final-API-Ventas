package sale

import "errors"

var ErrNotFound = errors.New("venta no encontrada")
var ErrEmptyID = errors.New("ID de venta vacía")

// storage define la interfaz para el almacenamiento de ventas
type Storage interface {
	Set(sale *Sale) error
	Read(id string) (*Sale, error)
	GetSales(id string, status string) ([]*Sale, error)
}

// LocalStorage es una implementación en memoria del almacenamiento de ventas
type LocalStorage struct {
	m map[string]*Sale
}

// NewLocalStorage instantiates a new LocalStorage with an empty map.
func NewLocalStorage() *LocalStorage {
	return &LocalStorage{
		m: map[string]*Sale{},
	}
}

func (l *LocalStorage) Set(sale *Sale) error {
	if sale.ID == "" {
		return ErrEmptyID
	}

	l.m[sale.ID] = sale
	return nil
}

// Read retrieves a sale from the local storage by ID.
// Returns ErrNotFound if the sale is not found.
func (l *LocalStorage) Read(id string) (*Sale, error) {
	s, ok := l.m[id]
	if !ok {
		return nil, ErrNotFound
	}

	return s, nil
}
func (l *LocalStorage) GetSales(user_id string, status string) ([]*Sale, error) {
	sales := []*Sale{}

	if status == "" {
		for _, v := range l.m {
			if v.UserID == user_id {
				sales = append(sales, v)
			}
		}
	} else {
		for _, v := range l.m {
			if v.UserID == user_id && v.Status == status {
				sales = append(sales, v)
			}
		}
	}

	return sales, nil
}
