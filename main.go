// Copyright 2022 The Ruliad Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

type Value uint64

const (
	// None is the none value
	None Value = iota
	// A is the a value
	A
	// B is the b value
	B
)

// Node is the node in a binary tree
type Node struct {
	Value
	L, R    *Node
	Matched bool
}

func main() {
	root := &Node{}
	root.L = &Node{
		Value: A,
	}
	root.R = &Node{
		Value: B,
	}

	var iterate func(n *Node, matched bool) (bool, *Node)
	iterate = func(n *Node, matched bool) (bool, *Node) {
		if n.Value > 0 {
			return true, n
		}

		if !matched {
			if n.Matched {
				new := &Node{}
				l, left := iterate(n.L, false)
				r, right := iterate(n.R, false)
				new.L = left
				new.R = right
				return l && r, new
			} else {
				new := &Node{}
				_, left := iterate(n.L, true)
				_, right := iterate(n.R, true)

				new.L = &Node{
					L: right,
					R: left,
				}
				new.R = right
				n.Matched = true
				return false, new
			}
		}

		new := &Node{}
		l, left := iterate(n.L, true)
		r, right := iterate(n.R, true)
		new.L = left
		new.R = right
		return l && r && n.Matched, new
	}

	var toString func(n *Node) string
	toString = func(n *Node) string {
		if n.Value == A {
			return "a"
		}

		if n.Value == B {
			return "b"
		}

		return fmt.Sprintf("(%s*%s)", toString(n.L), toString(n.R))
	}

	nodes := make([]*Node, 0, 8)
	nodes = append(nodes, root)
	apply := func(nodes []*Node) []*Node {
		children := make([]*Node, 0, 8)
		for _, node := range nodes {
			found, child := iterate(node, false)
			for !found {
				children = append(children, child)
				found, child = iterate(node, false)
			}
		}
		return children
	}

	next := apply(nodes)
	for i := 0; i < 8; i++ {
		tmp := apply(next)
		nodes = append(nodes, next...)
		next = tmp
	}
	for _, node := range nodes {
		fmt.Println(toString(node))
	}
}
