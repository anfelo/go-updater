package datasource

import "github.com/go-resty/resty/v2"

type Datasource struct {
	Client *resty.Client
}

// NewSource - Creates a new database connection
func NewDatasource() *Datasource {
	return &Datasource{
		Client: resty.New(),
	}
}
