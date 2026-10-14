package repository

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// testEntity is a dialect-neutral entity (integer PK) used to exercise the
// generic GORM adapter without depending on MariaDB specifics.
type testEntity struct {
	gorm.Model
	Name string
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	// A single connection keeps the shared in-memory database alive.
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("underlying sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(&testEntity{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	return db
}

func TestNewGormCreateAndFind(t *testing.T) {
	db := newTestDB(t)
	repo := NewGorm[testEntity](db)

	created := &testEntity{Name: "alpha"}
	if err := repo.Create(created); err != nil {
		t.Fatalf("create: %v", err)
	}

	found, err := repo.Find("1")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if found.Name != "alpha" {
		t.Fatalf("Name = %q, want %q", found.Name, "alpha")
	}
}

func TestGormFindNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewGorm[testEntity](db)

	if _, err := repo.Find("999"); err == nil {
		t.Fatal("expected error for missing record, got nil")
	}
}

func TestGormCreateBulk(t *testing.T) {
	db := newTestDB(t)
	repo := NewGorm[testEntity](db)

	if err := repo.CreateBulk([]testEntity{{Name: "a"}, {Name: "b"}, {Name: "c"}}); err != nil {
		t.Fatalf("create bulk: %v", err)
	}

	var count int64
	db.Model(&testEntity{}).Count(&count)
	if count != 3 {
		t.Fatalf("count = %d, want 3", count)
	}
}

func TestGormFindBy(t *testing.T) {
	db := newTestDB(t)
	repo := NewGorm[testEntity](db)

	repo.CreateBulk([]testEntity{{Name: "keep"}, {Name: "drop"}, {Name: "keep"}})

	got, err := repo.FindBy(Field{Column: "name", Value: "keep"})
	if err != nil {
		t.Fatalf("find by: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
}

func TestGormFindByWithRelations(t *testing.T) {
	db := newTestDB(t)
	repo := NewGorm[testEntity](db)

	repo.Create(&testEntity{Name: "x"})

	got, err := repo.FindByWithRelations(Field{Column: "name", Value: "x"})
	if err != nil {
		t.Fatalf("find by with relations: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
}

func TestGormFindFirstBy(t *testing.T) {
	db := newTestDB(t)
	repo := NewGorm[testEntity](db)

	repo.CreateBulk([]testEntity{{Name: "one"}, {Name: "two"}})

	got, err := repo.FindFirstBy(Field{Column: "name", Value: "two"})
	if err != nil {
		t.Fatalf("find first by: %v", err)
	}
	if got.Name != "two" {
		t.Fatalf("Name = %q, want %q", got.Name, "two")
	}
}

func TestGormFindFirstByEmpty(t *testing.T) {
	db := newTestDB(t)
	repo := NewGorm[testEntity](db)

	if _, err := repo.FindFirstBy(Field{Column: "name", Value: "nope"}); err == nil {
		t.Fatal("expected error when nothing matches, got nil")
	}
}

func TestGormFindWithRelations(t *testing.T) {
	db := newTestDB(t)
	repo := NewGorm[testEntity](db)

	repo.Create(&testEntity{Name: "solo"})

	got, err := repo.FindWithRelations("1")
	if err != nil {
		t.Fatalf("find with relations: %v", err)
	}
	if got.Name != "solo" {
		t.Fatalf("Name = %q, want %q", got.Name, "solo")
	}
}

func TestGormFindPaginated(t *testing.T) {
	db := newTestDB(t)
	repo := NewGorm[testEntity](db)

	repo.CreateBulk([]testEntity{
		{Name: "1"}, {Name: "2"}, {Name: "3"}, {Name: "4"}, {Name: "5"},
	})

	p, err := repo.FindPaginated(2, 2)
	if err != nil {
		t.Fatalf("find paginated: %v", err)
	}

	if p.TotalRows != 5 {
		t.Fatalf("TotalRows = %d, want 5", p.TotalRows)
	}
	if p.TotalPages != 3 {
		t.Fatalf("TotalPages = %d, want 3", p.TotalPages)
	}

	rows := p.Rows.([]*testEntity)
	if len(rows) != 2 {
		t.Fatalf("len(Rows) = %d, want 2", len(rows))
	}
	// page 2 with limit 2 => offset 2 => rows 3 and 4.
	if rows[0].Name != "3" || rows[1].Name != "4" {
		t.Fatalf("page 2 rows = [%s,%s], want [3,4]", rows[0].Name, rows[1].Name)
	}
}

func TestGormUpdate(t *testing.T) {
	db := newTestDB(t)
	repo := NewGorm[testEntity](db)

	created := &testEntity{Name: "before"}
	repo.Create(created)

	if err := repo.Update(created, Field{Column: "name", Value: "after"}); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, _ := repo.Find("1")
	if got.Name != "after" {
		t.Fatalf("Name = %q, want %q", got.Name, "after")
	}
}

func TestGormDelete(t *testing.T) {
	db := newTestDB(t)
	repo := NewGorm[testEntity](db)

	created := &testEntity{Name: "doomed"}
	repo.Create(created)

	if err := repo.Delete(created); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if _, err := repo.Find("1"); err == nil {
		t.Fatal("expected deleted record to be gone, got nil")
	}
}
