package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type Product struct {
	price float64
	count int
}
type Store struct {
	products map[string]*Product
}

func NewStore() *Store {
	return &Store{
		products: make(map[string]*Product),
	}
}

func (s *Store) AddProduct(name string, price float64, count int) error {
	key := strings.ToLower(name)

	if _, ok := s.products[key]; ok {
		return fmt.Errorf("%s already exists", name)
	}

	if price <= 0 {
		return errors.New("price should be positive")
	}
	if count <= 0 {
		return errors.New("count should be positive")
	}
	s.products[key] = &Product{
		price: price,
		count: count,
	}
	return nil
}

func (s *Store) GetProductCount(name string) (int, error) {
	key := strings.ToLower(name)
	p, ok := s.products[key]
	if !ok {
		return 0, errors.New("invalid product name")
	}

	return p.count, nil
}

func (s *Store) GetProductPrice(name string) (float64, error) {
	key := strings.ToLower(name)
	p, ok := s.products[key]
	if !ok {
		return 0, errors.New("invalid product name")
	}
	return p.price, nil
}

func (s *Store) Order(name string, count int) error {
	key := strings.ToLower(name)
	p, ok := s.products[key]
	if count <= 0 {
		return errors.New("count should be positive")
	}
	if !ok {
		return errors.New("invalid product name")
	}
	if p.count == 0 {
		return fmt.Errorf("there is no %s in the store", name)
	}
	if count > p.count {
		return fmt.Errorf("not enough %s in the store. there are %d left", name, p.count)
	}
	p.count -= count
	return nil
}

func (s *Store) ProductsList() ([]string, error) {
	if len(s.products) == 0 {
		return nil, errors.New("store is empty")
	}

	listStore := make([]string, 0)
	for name, p := range s.products {
		if p.count > 0 {
			key := strings.ToLower(name)
			listStore = append(listStore, key)
		}
	}
	if len(listStore) == 0 {
		return nil, errors.New("store is empty")
	}
	sort.Strings(listStore)
	return listStore, nil
}
