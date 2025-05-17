package query

type query struct {
	executor DBExecutor
}

func NewQuery(executor DBExecutor) *query {
	return &query{
		executor: executor,
	}
}
