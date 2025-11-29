package charts

import (
	"fmt"
	"os"

	"github.com/prometheus/common/model"
	"golang.org/x/term"
)

type Charter interface {
	PrintQuery(model.Vector)
}

type ntCharts struct{}

func NewNtCharts() Charter {
	return &ntCharts{}
}

func (*ntCharts) PrintQuery(vector model.Vector) {
	width, _, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Printf("Error getting terminal size: %v\n", err)
		return
	}
	bc := Barchart(vector, width)
	fmt.Println(bc)
}
