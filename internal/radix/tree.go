package radix

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

type nodeType int

func (i nodeType) IsNotNil() bool { return i&nonNilNode != 0 }

const (
	nonNilNode nodeType = 1 << iota
	staticNode
	capDirNode
	capAllNode
	nodeTypeMask = staticNode | capDirNode | capAllNode
)

var nodeTypeName = map[nodeType]string{
	0:                       "nil",
	staticNode:              "nil|static",
	capDirNode:              "nil|cap_dir",
	capAllNode:              "nil|cap_all",
	nonNilNode | staticNode: "static",
	nonNilNode | capDirNode: "cap_dir",
	nonNilNode | capAllNode: "cap_all",
}

func (i nodeType) String() string {
	if s := nodeTypeName[i]; s != "" {
		return s
	}
	return fmt.Sprintf("nodeType(%d)", i)
}

type Payload struct {
	Handler http.Handler
}

type Node struct {
	path     string
	typ      nodeType
	index    string
	capIdx   int // +1
	children []Node
	Payload
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

func newNode(path string) Node {
	var node Node
	node.setPath(path)
	return node
}

func (node *Node) setPath(path string) *Node {
	node.path = path
	node.typ &= ^nodeTypeMask
	switch byteForIdx(path) {
	case ':':
		node.typ |= capDirNode
	case '*':
		node.typ |= capAllNode
	default:
		node.typ |= staticNode
	}
	return node
}

func (node *Node) append(c Node) *Node {
	node.children = append(node.children, c)
	b := byteForIdx(c.path)
	node.index = node.index + string(b)
	if b == '*' || b == ':' {
		node.capIdx = len(node.index)
	}
	return &node.children[len(node.children)-1]
}

func (node *Node) appendSplit(path string) *Node {
	ss := splitCompact(path)
	for i, s := range ss {
		child := newNode(s)
		if child.typ&capAllNode != 0 && i != len(s)-1 {
			panic(fmt.Errorf(
				"radix: invalid pattern %q: capture-all can't be followed by path",
				strings.Join(ss[i:], "/"),
			))
		}
		node = node.append(child)
	}
	return node
}

func splitCompact(path string) (ss []string) {
	ss = strings.Split(path, "/")
	special := func(s string) bool {
		switch byteForIdx(s) {
		case ':', '*':
			return true
		}
		return false
	}

	for i := 0; i < len(ss)-1; {
		if !special(ss[i]) && !special(ss[i+1]) {
			ss[i+1] = strings.Join(ss[i:i+2], "/")
			copy(ss[i:], ss[i+1:])
			ss = ss[:len(ss)-1]
			continue
		}
		i++
	}
	return ss
}

func (t *Tree) Add(path string, v Payload) (old Payload, replace bool) {
	ss := strings.Split(path, "/")
	return t.root.insert(ss, v)
}

func commonPrefix(p1, p2 []string) (i int) {
	for i = 0; len(p1) > 0 && len(p2) > 0 && p1[0] == p2[0]; i++ {
		p1, p2 = p1[1:], p2[1:]
	}
	return i
}

func (node *Node) replace(v Payload) (old Payload, replace bool) {
	old, replace = node.Payload, node.typ.IsNotNil()
	node.Payload = v
	node.setPath(node.path) // For root node
	node.typ |= nonNilNode
	return old, replace
}

// Invariant: strings.SplitN(node.dir, "/", 2)[0] == newpath[0]
func (node *Node) insert(newpath []string, v Payload) (old Payload, replace bool) {
	var (
		path = strings.Split(node.path, "/")
		n    = commonPrefix(path, newpath)
	)
	switch {
	case n < len(path): // Split current node
		l := len(strings.Join(newpath[:n], "/"))
		child := *node
		child.path = node.path[l+1:]
		*node = newNode(node.path[:l])
		node.append(child)

		if n == len(newpath) {
			return node.replace(v)
		}
		return node.appendSplit(
			strings.Join(newpath[n:], "/"),
		).replace(v)

	case n == len(newpath): // Match current node
		return node.replace(v)
	}

	// Try to go deeper
	var (
		dir      = newpath[n]
		b        = byteForIdx(dir)
		index    = node.index
		i        = strings.IndexByte(index, b)
		children = node.children
	)
	for ; i >= 0; i = strings.IndexByte(index, b) {
		if next := &children[i]; strings.SplitN(next.path, "/", 2)[0] == dir {
			return next.insert(newpath[n:], v)
		}
		if b == ':' || b == '*' {
			panic(fmt.Errorf(
				"radix: conflict parameter name: old=%q, new=%q",
				children[i].path, dir,
			))
		}
		index, children = index[i+1:], children[i+1:]
	}

	if i = strings.IndexAny(node.index, ":*"); (b == ':' || b == '*') && i >= 0 {
		panic(fmt.Errorf(
			"radix: conflict parameter type: old=%q, new=%q",
			node.children[i].path, dir,
		))
	}

	// Failed, append to the child list of current node
	return node.appendSplit(
		strings.Join(newpath[n:], "/"),
	).replace(v)
}

func (t *Tree) Lookup(path string) *Node {
	node := &t.root

OUTER:
	for len(path) > 0 {
		// Invariant: path[0] = '/'
		path = path[1:]

		var (
			b        = byteForIdx(path)
			index    = node.index
			i        = strings.IndexByte(index, b)
			children = node.children
		)
		for ; i >= 0; i = strings.IndexByte(index, b) {
			var (
				child    = &children[i]
				pos, min int
			)
			if min = len(child.path); min > len(path) {
				min = len(path)
			}
			for ; pos < min && child.path[pos] == path[pos]; pos++ {
			}
			switch {
			case pos < len(child.path): // Not match
				break
			case pos == len(path): // Match
				return child
			case path[pos] != '/': // Not match
				break
			default: // Go deeper
				path = path[pos:]
				node = child
				continue OUTER
			}
			index, children = index[i+1:], children[i+1:]
		}

		if i = node.capIdx - 1; i >= 0 {
			switch node.index[i] {
			case ':':
				pos := strings.IndexByte(path, '/')
				if pos < 0 {
					pos = len(path)
				}
				path = path[pos:]
				node = &node.children[i]
				continue OUTER
			case '*':
				path = path[len(path):]
				node = &node.children[i]
				continue OUTER
			}
		}

		return nil
	}
	return node
}

func (node *Node) String() string {
	return fmt.Sprintf(
		"Node{dir: %q, #child: %d, index: %q, type: %s}",
		node.path, len(node.children), string(node.index), node.typ,
	)
}

func strIF(b bool, s1, s2 string) string {
	if b {
		return s1
	}
	return s2
}

func (t *Tree) String() string {
	const (
		blank    = "    "
		edge     = "|   "
		leaf     = "|---"
		sideLeaf = "+---"
	)
	type context struct {
		first, last bool
		nextIsLast  bool
		level       int
	}
	var (
		f      func(*Node, int, []context)
		result bytes.Buffer
	)
	f = func(n *Node, level int, ctx []context) {
		var line bytes.Buffer
		for _, c := range ctx {
			if c.last {
				line.WriteString(strIF(c.level < level, blank, sideLeaf))
			} else if c.first {
				line.WriteString(strIF(c.level < level, edge,
					strIF(c.nextIsLast, leaf, sideLeaf)))
			} else {
				line.WriteString(strIF(c.level < level, edge, leaf))
			}
		}

		line.WriteTo(&result)
		result.WriteByte('"')
		result.WriteString(n.path)
		result.WriteString(`"\n`)

		for i := range n.children {
			ctx = append(ctx, context{
				first:      i == 0,
				last:       i == len(n.children)-1,
				nextIsLast: i == len(n.children)-2,
				level:      level + 1,
			})
			f(&n.children[i], level+1, ctx)
			ctx = ctx[:len(ctx)-1]
		}
	}
	ctx := make([]context, 0, 8)
	f(&t.root, 0, ctx)
	return result.String()
}
