package radix

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

type Payload struct {
	Handler http.Handler
}

type Node struct {
	path     string
	index    string
	children []Node
	Payload
}

type Tree struct {
	root Node
}

func firstbyte(dir string) byte {
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
	return node
}

func (node *Node) captureNode() *Node {
	if i := len(node.index); special(node.index[i-1]) {
		return &node.children[i-1]
	}
	return nil
}

func special(b byte) bool {
	return b == '*' || b == ':'
}

func (node *Node) append(c Node) *Node {
	if i := len(node.index) - 1; i >= 0 && special(node.index[i]) {
		last := node.children[i]
		node.index, node.children = node.index[:i], node.children[:i]
		node.append(c)
		node.append(last)
		return &node.children[len(node.children)-2]
	}
	node.children = append(node.children, c)
	node.index = node.index + string(firstbyte(c.path))
	return &node.children[len(node.children)-1]
}

func (node *Node) appendSplit(path string) *Node {
	ss := splitCompact(path)
	for i, s := range ss {
		child := newNode(s)
		if firstbyte(child.path) == '*' && i != len(ss)-1 {
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
		switch firstbyte(s) {
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
	old, replace = node.Payload, node.Handler != nil
	node.Payload = v
	node.setPath(node.path) // For root node
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
		b        = firstbyte(dir)
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

	if (b == ':' || b == '*') && node.captureNode() != nil {
		panic(fmt.Errorf(
			"radix: conflict parameter type: old=%q, new=%q",
			node.captureNode().path, dir,
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
			b        = firstbyte(path)
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

		if l := len(node.index) - 1; l >= 0 {
			switch node.index[l] {
			case ':':
				pos := strings.IndexByte(path, '/')
				if pos < 0 {
					pos = len(path)
				}
				path = path[pos:]
				node = &node.children[l]
				continue OUTER
			case '*':
				path = path[len(path):]
				node = &node.children[l]
				continue OUTER
			}
		}

		return nil
	}
	return node
}

func (node *Node) String() string {
	return fmt.Sprintf(
		"Node{dir: %q, #child: %d, index: %q}",
		node.path, len(node.children), string(node.index),
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
		path        string
	}
	var (
		f           func(*Node, int, []context)
		tree        bytes.Buffer
		annotations []string
		result      bytes.Buffer
		maxlen      int
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
		// line.WriteByte('"')
		line.WriteString(strIF(len(n.path) > 0, n.path, "."))
		// line.WriteByte('"')
		if line.Len() > maxlen {
			maxlen = line.Len()
		}
		line.WriteTo(&tree)
		tree.WriteByte('\n')

		path := ""
		if n.Handler != nil {
			for _, c := range ctx {
				path += "/" + c.path
			}
		}
		annotations = append(annotations, path)

		for i := range n.children {
			ctx = append(ctx, context{
				first:      i == 0,
				last:       i == len(n.children)-1,
				nextIsLast: i == len(n.children)-2,
				level:      level + 1,
				path:       n.children[i].path,
			})
			f(&n.children[i], level+1, ctx)
			ctx = ctx[:len(ctx)-1]
		}
	}

	ctx := make([]context, 0, 8)
	f(&t.root, 0, ctx)

	scanner := bufio.NewScanner(&tree)
	for scanner.Scan() {
		line := scanner.Bytes()
		result.Write(line)
		for i := maxlen + 2 - len(line); i > 0; i-- {
			result.WriteByte(' ')
		}
		result.WriteString(annotations[0])
		annotations = annotations[1:]
		result.WriteByte('\n')
	}
	return result.String()
}

func (t *Tree) Optimize() {
	var depthfirst, breadthfirst func(*Node, func(*Node))

	// NOTE: don't modify Node.children in function f.
	depthfirst = func(n *Node, f func(*Node)) {
		f(n)
		for i := range n.children {
			depthfirst(&n.children[i], f)
		}
	}
	// Modifying node.children in function f is safe.
	breadthfirst = func(n *Node, f func(*Node)) {
		queue := make([]*Node, 0, 16)
		queue = append(queue, n)
		for len(queue) > 0 {
			n, queue = queue[0], queue[1:]
			f(n)
			for i := range n.children {
				queue = append(queue, &n.children[i])
			}
		}
	}

	// Move *param or :param to the end of children list
	/*
		breadthfirst(&t.root, func(n *Node) {
			if n.icap <= 0 {
				return
			}
			i := n.icap - 1
			child := n.children[i]
			copy(n.children[i:], n.children[i+1:])
			n.children = n.children[:len(n.children)-1]
			n.children = append(n.children, child)
			n.icap = int32(len(n.children))
			n.index = ""
			for i := range n.children {
				n.index += string(firstbyte(n.children[i].path))
			}
		})
	*/

	// Count the number of nodes
	var count int
	depthfirst(&t.root, func(*Node) { count++ })

	// Move all nodes but root to a continuous memory segment
	nodes := make([]Node, 0, count-1)
	breadthfirst(&t.root, func(n *Node) {
		nodes = append(nodes, n.children...)
	})
	breadthfirst(&t.root, func(n *Node) {
		n.children = nodes[:len(n.children)]
		nodes = nodes[len(n.children):]
	})
	if len(nodes) != 0 {
		panic("radix: optimization failed")
	}

	// Move all index string to a continuous memory segment
	var buf bytes.Buffer
	breadthfirst(&t.root, func(n *Node) {
		buf.WriteString(n.index)
	})
	index := buf.String()
	breadthfirst(&t.root, func(n *Node) {
		n.index = index[:len(n.index)]
		index = index[len(n.index):]
	})
	if len(index) != 0 {
		panic("radix: optimization failed")
	}

	// Move all path string to a continuous memory segment (DFS)
	buf.Reset()
	depthfirst(&t.root, func(n *Node) {
		buf.WriteString(n.path)
	})
	path := buf.String()
	depthfirst(&t.root, func(n *Node) {
		n.path = path[:len(n.path)]
		path = path[len(n.path):]
	})
	if len(path) != 0 {
		panic("radix: optimization failed")
	}
}
