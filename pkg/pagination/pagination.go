package pagination

// Pagination represents technical pagination parameters for database queries
type Pagination struct {
	Limit  int
	Offset int
}

// NewPagination creates a new Pagination with limit and offset
func NewPagination(limit, offset int) *Pagination {
	return &Pagination{
		Limit:  limit,
		Offset: offset,
	}
}

// NewDefaultPagination creates a new Pagination with default values
func NewDefaultPagination() *Pagination {
	return &Pagination{
		Limit:  10,
		Offset: 0,
	}
}
