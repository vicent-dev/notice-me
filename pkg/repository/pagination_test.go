package repository

import "testing"

func TestGetLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit int
		want  int
	}{
		{name: "uses provided limit", limit: 5, want: 5},
		{name: "defaults when zero", limit: 0, want: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pagination{Limit: tt.limit}
			if got := p.GetLimit(); got != tt.want {
				t.Fatalf("GetLimit() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetPage(t *testing.T) {
	tests := []struct {
		name string
		page int
		want int
	}{
		{name: "uses provided page", page: 3, want: 3},
		{name: "defaults when zero", page: 0, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pagination{Page: tt.page}
			if got := p.GetPage(); got != tt.want {
				t.Fatalf("GetPage() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetSort(t *testing.T) {
	tests := []struct {
		name string
		sort string
		want string
	}{
		{name: "uses provided sort", sort: "name asc", want: "name asc"},
		{name: "defaults when empty", sort: "", want: "Id desc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pagination{Sort: tt.sort}
			if got := p.GetSort(); got != tt.want {
				t.Fatalf("GetSort() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetOffset(t *testing.T) {
	tests := []struct {
		name  string
		limit int
		page  int
		want  int
	}{
		{name: "first page has no offset", limit: 10, page: 1, want: 0},
		{name: "second page", limit: 10, page: 2, want: 10},
		{name: "third page custom size", limit: 25, page: 3, want: 50},
		{name: "zero values fall back to defaults", limit: 0, page: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pagination{Limit: tt.limit, Page: tt.page}
			if got := p.GetOffset(); got != tt.want {
				t.Fatalf("GetOffset() = %d, want %d", got, tt.want)
			}
		})
	}
}
