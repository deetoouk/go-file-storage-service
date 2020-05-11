package repository

// ListOptions allows you to customise your results
type ListOptions struct {
	Limit          int64
	Page           int64
	OrderBy        string
	OrderDirection int8
}
