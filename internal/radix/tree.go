package radix

import (
	"fmt"
	"net/http"
	"strings"
)

//go:generate stringer -type=nodeType

type nodeType int

func (i nodeType) IsNotNil() bool { return i&nonNilNode != 0 }

const (
	nonNilNode nodeType = 1 << iota
	staticNode
	capDirNode
	capAllNode
	nodeTypeMask = staticNode | capDirNode | capAllNode
)

type Data struct {
	Handler http.Handler
}

type Node struct {
	dir      string
	typ      nodeType
	index    string
	children []Node
	Data
}

type Tree struct {
	root Node
}

func byteForIdx(dir string) byte {
	if dir == "" {
		return '/'
	}
	return dir[0]
}

func (node *Node) append(c Node) *Node {
	node.children = append(node.children, c)
	node.index = node.index + string(byteForIdx(c.dir))
	return &node.children[len(node.children)-1]
}

func (t *Tree) Add(path string, v Data) (old Data, replace bool) {
	ss := strings.Split(path, "/")
	return t.root.insert(ss[1:], v)
}

func (node *Node) insert(ss []string, v Data) (old Data, replace bool) {
	if len(ss) == 0 {
		old, replace = node.Data, node.typ.IsNotNil()
		node.Data = v
		node.typ = staticNode | nonNilNode
		return old, replace
	}
	var (
		dir      = ss[0]
		b        = byteForIdx(dir)
		index    = node.index
		i        = strings.IndexByte(index, b)
		children = node.children
	)
	for ; i >= 0; i = strings.IndexByte(index, b) {
		if next := &children[i]; next.dir == dir {
			return next.insert(ss[1:], v)
		}
		index, children = index[i+1:], children[i+1:]
	}
	next := node.append(Node{
		dir: dir,
		typ: staticNode,
	})
	return next.insert(ss[1:], v)
}

func (t *Tree) Lookup(path string) *Node {
	node := &t.root

OUTER:
	for pos := 0; len(path) > 0; path = path[pos:] {
		path = path[1:]

		if pos = strings.IndexByte(path, '/'); pos < 0 {
			pos = len(path)
		}

		var (
			dir      = path[:pos]
			b        = byteForIdx(dir)
			index    = node.index
			i        = strings.IndexByte(index, b)
			children = node.children
		)
		for ; i >= 0; i = strings.IndexByte(index, b) {
			if next := &children[i]; next.dir == dir {
				node = next
				continue OUTER
			}
			index, children = index[i+1:], children[i+1:]
		}

		return nil
	}

	return node
}

func (node *Node) String() string {
	return fmt.Sprintf(
		"Node{dir: %q, #child: %d, index: %s, type: %s}",
		node.dir, len(node.children), string(node.index), node.typ,
	)
}
