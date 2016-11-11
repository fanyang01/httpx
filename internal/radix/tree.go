package radix

import (
	"bytes"
	"fmt"
	"strings"
)

type NodeType int

const (
	TypeNil   NodeType = 0
	TypeEmpty NodeType = 1 << iota
	TypeSimple
	TypeMatch
	TypeCapture
	TypeHasValue = TypeSimple | TypeMatch | TypeCapture
)

type Node struct {
	dir   string
	child []Node
	index []byte
	typ   NodeType
	v     int
}

const None = -1

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
	node.child = append(node.child, c)
	node.index = append(node.index, byteForIdx(c.dir))
	return &node.child[len(node.child)-1]
}

func (t *Tree) Add(path string, v int) (ov int, replace bool) {
	if t.root.typ == TypeNil {
		t.root.dir = ""
		t.root.typ = TypeEmpty
		t.root.v = None
	}
	ss := strings.Split(path, "/")
	return t.root.insert(ss[1:], v)
}

func (node *Node) insert(ss []string, v int) (ov int, replace bool) {
	if len(ss) == 0 {
		ov, replace = node.v, node.typ&TypeHasValue != 0
		node.v = v
		node.typ = TypeSimple
		return ov, replace
	}
	var (
		dir   = ss[0]
		b     = byteForIdx(dir)
		index = node.index
		i     = bytes.IndexByte(index, b)
	)
	for ; i >= 0; i = bytes.IndexByte(index, b) {
		if next := &node.child[i]; next.dir == dir {
			return next.insert(ss[1:], v)
		}
		index = index[i+1:]
	}
	next := node.append(Node{
		dir: dir,
		typ: TypeEmpty,
		v:   None,
	})
	return next.insert(ss[1:], v)
}

func (t *Tree) Lookup(path string) (v int, exact bool) {
	if t.root.typ == TypeNil {
		return None, false
	}

	node := &t.root

OUTER:
	for ptr := 0; len(path) > 0; path, ptr = path[ptr+1:], 0 {
		fmt.Printf("node{dir: %q, #child: %d, index: %s, type: %d}\n", node.dir, len(node.child), string(node.index), node.typ)
		if ptr = strings.IndexByte(path, '/'); ptr < 0 {
			// TODO
			ptr = len(path)
		}

		var (
			dir   = path[:ptr]
			b     = byteForIdx(dir)
			index = node.index
			i     = bytes.IndexByte(index, b)
		)
		for ; i >= 0; i = bytes.IndexByte(index, b) {
			if next := &node.child[i]; next.dir == dir {
				node = next
				continue OUTER
			}
			index = index[i+1:]
		}

		return node.v, false
	}

	return node.v, true
}
