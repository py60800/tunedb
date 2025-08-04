// testIndex
package search

import (
	"fmt"
	"sort"
	"unicode"

	"github.com/mozillazg/go-unidecode"
)

var nodeIgnore = map[string]int{
	"the": 0,
	"an":  0,
}

func nodeSplit(s string) []string {
	s = unidecode.Unidecode(s)
	ns := ""
	found := make(map[string]int)
	for _, c := range s {
		if unicode.IsLetter(c) {
			ns = ns + string(unicode.ToLower(c))
		} else {
			if len(ns) > 1 {
				if _, ignorable := nodeIgnore[ns]; !ignorable {
					found[ns] = 0
				}
			}
			ns = ""
		}
	}
	if len(ns) > 1 {
		if _, ignorable := nodeIgnore[ns]; !ignorable {
			found[ns] = 0
		}
	}

	res := make([]string, len(found))
	i := 0
	for k := range found {
		res[i] = k
		i++
	}
	return res
}

type Node struct {
	next  map[rune]*Node
	match []int
}

func NodeNew() *Node {
	return &Node{
		next:  make(map[rune]*Node),
		match: make([]int, 0),
	}
}
func printNode(pfx string, node *Node) {
	fmt.Println(pfx, "Node")
	pfx = pfx + "..."
	fmt.Println(pfx, "Match:", node.match)
	for k, nex := range node.next {
		fmt.Println(pfx, "[", k, "]")
		printNode(pfx+"..", nex)

	}
}
func (node *Node) appendToNode(word []rune, idx int) {
	if len(node.match) > 0 && node.match[len(node.match)-1] == idx {
		//fmt.Println("Zarbi node", string(word))
	} else {
		node.match = append(node.match, idx)
	}
	if len(word) == 0 {
		return
	}
	ch := word[0]
	if nextNode, ok := node.next[ch]; ok {
		nextNode.appendToNode(word[1:], idx)
	} else {
		newnode := NodeNew()
		node.next[ch] = newnode
		newnode.appendToNode(word[1:], idx)
	}
}
func (node *Node) indexWord(word []rune, idx int) {
	for i := range word {
		node.appendToNode(word[i:], idx)
	}
}
func (node *Node) search(word []rune) []int {
	if len(word) == 0 {
		return node.match
	}
	if nextNode, ok := node.next[word[0]]; ok {
		return nextNode.search(word[1:])
	}
	return []int{}
}
func (node *Node) Search(what string) []int {
	words := nodeSplit(what)
	fmt.Println(words)
	lst := make(map[int]int)
	for i, w := range words {
		l := node.search([]rune(w))
		//sort.Ints(l)

		fmt.Println(w, ":", len(l), l[:min(10, len(l))])
		if i == 0 {
			for _, n := range l {
				lst[n] = 1
			}
		} else {
			for _, n := range l {
				if v, ok := lst[n]; ok {
					lst[n] = v + 1
				}
			}
		}
		fmt.Println(lst)
	}
	l := make([]int, 0)
	for k, v := range lst {
		if v >= len(words) {
			l = append(l, k)
		}
	}
	sort.Ints(l)
	return l
}
func (node *Node) IndexWords(what string, i int) {
	uword := nodeSplit(what)
	for _, w := range uword {
		node.indexWord([]rune(w), i)
	}
}
func (node *Node) Count() int {
	count := 0
	for _, n := range node.next {
		count += n.Count()
	}
	return count + 1
}
