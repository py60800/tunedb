// Node_test.go
package search

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"
)

//go:embed titles.txt
var data string

//go:embed samples.txt
var samples string

func TestSearch(t *testing.T) {
	dataTest := strings.Split(data, "\n")
	tData := strings.Split(samples, "\n")

	node := NodeNew()
	for i, t := range dataTest {
		node.IndexWords(t, i)
	}
	fmt.Printf("DataSet:%v Nodes:%v\n", len(dataTest), node.Count())
	for _, t := range tData {
		fmt.Println("Search:", t)
		found := node.Search(t)
		fmt.Println("Found:", len(found))

		for _, l := range found {
			fmt.Printf("\t<%v>\n", dataTest[l])
		}
	}

	//	fmt.Println(node)

}
