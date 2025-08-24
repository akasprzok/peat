package commands

import (
	"fmt"

	"github.com/akasprzok/peat/internal/prometheus"
)

type FormatQueryCmd struct {
	Query string `arg:"" name:"query" help:"Query to format." required:"true"`
}

func (f *FormatQueryCmd) Run(_ *Context) error {
	fmt.Println(prometheus.FormatQuery(f.Query))
	return nil
}
