package mock

import (
	"errors"
	"notice-me-server/pkg/repository"
	"strconv"
)

type Repository[T repository.Entity] struct {
	entities map[string]*T
}

func NewRepository[T repository.Entity]() repository.Repository[T] {
	return Repository[T]{
		entities: make(map[string]*T),
	}
}

func (r Repository[T]) Find(id string) (*T, error) {

	if _, ok := r.entities[id]; !ok {
		return nil, errors.New("entity not found")
	}

	return r.entities[id], nil
}

func (r Repository[T]) FindPaginated(pageSize, page int) (*repository.Pagination, error) {
	es := r.entitiesSlice()
	//we are not going to test this since it uses gorm features
	return &repository.Pagination{
		Limit:     pageSize,
		Page:      page,
		Sort:      "",
		Rows:      es,
		TotalRows: int64(len(es)),
	}, nil
}

func (r Repository[T]) FindWithRelations(id string) (*T, error) {
	return r.Find(id)
}

func (r Repository[T]) FindByWithRelations(fs ...repository.Field) ([]*T, error) {
	//we are not going to test this since it uses gorm features
	return r.entitiesSlice(), nil
}

func (r Repository[T]) FindBy(fs ...repository.Field) ([]*T, error) {
	//we are not going to test this since it uses gorm features
	return r.entitiesSlice(), nil
}

func (r Repository[T]) FindFirstBy(fs ...repository.Field) (*T, error) {
	//we are not going to test this since it uses gorm features
	return r.entitiesSlice()[0], nil
}

func (r Repository[T]) Create(t *T) error {
	r.entities[strconv.Itoa(len(r.entities))] = t
	return nil
}

func (r Repository[T]) CreateBulk(ts []T) error {
	for _, t := range ts {
		r.Create(&t)
	}

	return nil
}

func (r Repository[T]) Update(t *T, fs ...repository.Field) error {
	//we are not going to test this since it uses gorm features
	return nil
}

func (r Repository[T]) Delete(t *T) error {
	var key string

	for k, e := range r.entities {
		if e == t {
			key = k
			break
		}
	}

	if key == "" {
		return errors.New("entity not found")
	}

	delete(r.entities, key)
	return nil
}

func (r Repository[T]) entitiesSlice() []*T {
	var es []*T

	for _, e := range r.entities {
		es = append(es, e)
	}

	return es
}
