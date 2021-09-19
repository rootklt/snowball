package api

type Queryer interface {
	Query() error
}

func DoQuery(q Queryer) error {
	return q.Query()
}
