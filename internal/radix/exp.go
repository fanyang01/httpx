package radix

// This file contains some experimental implementations.

import "strings"

func (node *Node) append_v0(c Node) *Node {
	node.children = append(node.children, c)
	node.index = node.index + string(firstbyte(c.path))
	return &node.children[len(node.children)-1]
}

func (node *Node) insert_v0(ss []string, v Payload) (old Payload, replace bool) {
	if len(ss) == 0 {
		old, replace = node.Payload, node.typ.IsNotNil()
		node.Payload = v
		node.typ = staticNode | nonNilNode
		return old, replace
	}
	var (
		dir      = ss[0]
		b        = firstbyte(dir)
		index    = node.index
		i        = strings.IndexByte(index, b)
		children = node.children
	)
	for ; i >= 0; i = strings.IndexByte(index, b) {
		if next := &children[i]; next.path == dir {
			return next.insert_v0(ss[1:], v)
		}
		index, children = index[i+1:], children[i+1:]
	}
	next := node.append_v0(Node{
		path: dir,
		typ:  staticNode,
	})
	return next.insert_v0(ss[1:], v)
}

func (t *Tree) lookup_v0(path string) *Node {
	node := &t.root

OUTER:
	for pos := 0; len(path) > 0; path = path[pos:] {
		path = path[1:]

		if pos = strings.IndexByte(path, '/'); pos < 0 {
			pos = len(path)
		}

		var (
			dir      = path[:pos]
			b        = firstbyte(dir)
			index    = node.index
			i        = strings.IndexByte(index, b)
			children = node.children
		)
		for ; i >= 0; i = strings.IndexByte(index, b) {
			if next := &children[i]; next.path == dir {
				node = next
				continue OUTER
			}
			index, children = index[i+1:], children[i+1:]
		}

		return nil
	}

	return node
}

// Invariant: strings.SplitN(node.dir, "/", 2)[0] == newpath[0]
func (node *Node) insert_v1(newpath []string, v Payload) (old Payload, replace bool) {
	var (
		path = strings.Split(node.path, "/")
		n    = commonPrefix(path, newpath)
	)
	switch {
	case n < len(path): // Split current node
		l := len(strings.Join(newpath[:n], "/"))
		child := *node
		child.path = node.path[l+1:]
		*node = Node{
			path: node.path[:l],
			typ:  staticNode,
		}
		node.append_v0(child)

		if n == len(newpath) {
			return node.replace(v)
		}
		return node.append_v0(Node{
			path: strings.Join(newpath[n:], "/"),
		}).replace(v)

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
		index, children = index[i+1:], children[i+1:]
	}

	// Failed, append to the child list of current node
	return node.append_v0(Node{
		path: strings.Join(newpath[n:], "/"),
	}).replace(v)
}

func (t *Tree) lookup_v1(path string) *Node {
	var (
		node    = &t.root
		delayed = false
		np      = -1
	)
OUTER:
	for pos := 0; len(path) > 0; path = path[pos:] {
		path = path[1:]

		if pos = strings.IndexByte(path, '/'); pos < 0 {
			pos = len(path)
		}

		dir := path[:pos]

		switch {
		case !delayed:
			break
		case !strings.HasPrefix(node.path[np:], dir):
			return nil
		case len(node.path[np:]) == len(dir):
			delayed, np = false, -1
			continue OUTER
		case node.path[np+len(dir)] == '/':
			np += len(dir) + 1
			continue OUTER
		default:
			return nil
		}

		var (
			b        = firstbyte(dir)
			index    = node.index
			i        = strings.IndexByte(index, b)
			children = node.children
		)
		for ; i >= 0; i = strings.IndexByte(index, b) {
			if next := &children[i]; strings.HasPrefix(next.path, dir) {
				if len(next.path) == len(dir) {
					node = next
					continue OUTER
				}
				if next.path[len(dir)] == '/' {
					node = next
					delayed, np = true, len(dir)+1
					continue OUTER
				}
			}
			index, children = index[i+1:], children[i+1:]
		}
		return nil
	}
	return node
}
