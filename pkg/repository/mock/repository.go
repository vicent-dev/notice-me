package mock

import (
	"errors"
	"notice-me-server/pkg/repository"
	"strconv"
)

type Repository[T repository.Entity] struct {
	entities map[string]*T

	FindErr          error
	FindPaginatedErr error
	CreateErr        error
	UpdateErr        error
	DeleteErr        error
	FindByErr        error
}

func NewRepository[T repository.Entity]() repository.Repository[T] {
	return NewErrorRepository[T]()
}

// NewErrorRepository returns the concrete mock so tests can inject failures
// through its exported error fields while still satisfying repository.Repository[T].
func NewErrorRepository[T repository.Entity]() *Repository[T] {
	return &Repository[T]{
		entities: make(map[string]*T),
	}
}

func (r Repository[T]) Find(id string) (*T, error) {

	if r.FindErr != nil {
		return nil, r.FindErr
	}

	if _, ok := r.entities[id]; !ok {
		return nil, errors.New("entity not found")
	}

	return r.entities[id], nil
}

func (r Repository[T]) FindPaginated(pageSize, page int) (*repository.Pagination, error) {
	if r.FindPaginatedErr != nil {
		return nil, r.FindPaginatedErr
	}

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
	if r.FindByErr != nil {
		return nil, r.FindByErr
	}

	//we are not going to test this since it uses gorm features
	return r.entitiesSlice(), nil
}

func (r Repository[T]) FindBy(fs ...repository.Field) ([]*T, error) {
	if r.FindByErr != nil {
		return nil, r.FindByErr
	}

	//we are not going to test this since it uses gorm features
	return r.entitiesSlice(), nil
}

func (r Repository[T]) FindFirstBy(fs ...repository.Field) (*T, error) {
	if r.FindByErr != nil {
		return nil, r.FindByErr
	}

	//we are not going to test this since it uses gorm features
	es := r.entitiesSlice()
	if len(es) == 0 {
		return nil, errors.New("entity not found")
	}

	return es[0], nil
}

func (r Repository[T]) Create(t *T) error {
	if r.CreateErr != nil {
		return r.CreateErr
	}

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
	if r.UpdateErr != nil {
		return r.UpdateErr
	}

	//we are not going to test this since it uses gorm features
	return nil
}

func (r Repository[T]) Delete(t *T) error {
	if r.DeleteErr != nil {
		return r.DeleteErr
	}

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
