package repository

type Pagination struct {
	Limit      int         `json:"Limit,omitempty;query:limit"`
	Page       int         `json:"Page,omitempty;query:page"`
	Sort       string      `json:"Sort,omitempty;query:sort"`
	TotalRows  int64       `json:"TotalRows"`
	TotalPages int         `json:"TotalPages"`
	Rows       interface{} `json:"Rows"`
}

const (
	DefaultPageSize = "10"
	DefaultPage     = "1"
)

func (p *Pagination) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

func (p *Pagination) GetLimit() int {
	if p.Limit == 0 {
		p.Limit = 10
	}
	return p.Limit
}

func (p *Pagination) GetPage() int {
	if p.Page == 0 {
		p.Page = 1
	}
	return p.Page
}

func (p *Pagination) GetSort() string {
	if p.Sort == "" {
		p.Sort = "Id desc"
	}
	return p.Sort
}
