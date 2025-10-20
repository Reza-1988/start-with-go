package main

type Store struct {
	//TODO
}

func NewStore() *Store {
	return nil
}

func (s *Store) AddProduct(name string, price float64, count int) error {
	return nil
}

func (s *Store) GetProductCount(name string) (int, error) {
	return 0, nil
}

func (s *Store) GetProductPrice(name string) (float64, error) {
	return 0, nil
}

func (s *Store) Order(name string, count int) error {
	return nil
}

func (s *Store) ProductsList() ([]string, error) {
	return nil, nil
}
