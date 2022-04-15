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
	Parent, ID uint
	L, R       *Node
	Matched    bool
}

// String is the string representation of a tree
func (n *Node) String() string {
	if n.Value == A {
		return "a"
	}

	if n.Value == B {
		return "b"
	}

	return fmt.Sprintf("(%s*%s)", n.L.String(), n.R.String())
}

// Copy copies the tree
func (n *Node) Copy() *Node {
	if n.Value > 0 {
		return n
	}

	new := &Node{}
	left := n.L.Copy()
	right := n.R.Copy()
	new.L = left
	new.R = right
	return new
}

// Apply applies the rule to node
func (n *Node) Apply() (bool, *Node) {
	if n.Value > 0 {
		return true, n
	}

	if n.Matched {
		new := &Node{}
		l, left := n.L.Apply()
		r, right := n.R.Apply()
		new.L = left
		new.R = right
		return l && r, new
	}

	new := &Node{}
	left := n.L.Copy()
	right := n.R.Copy()
	new.L = &Node{
		L: right,
		R: left,
	}
	new.R = right
	n.Matched = true
	return false, new
}

func main() {
	root := &Node{}
	root.L = &Node{
		Value: A,
	}
	root.R = &Node{
		Value: B,
	}

	apply := func(nodes []*Node) []*Node {
		children := make([]*Node, 0, 8)
		for _, node := range nodes {
			id := node.ID
			found, child := node.Apply()
			for !found {
				child.Parent = id
				children = append(children, child)
				found, child = node.Apply()
			}
		}
		return children
	}

	type ID struct {
		ID   uint
		Node *Node
	}

	var id uint
	graph, nodes, count, ids := make(map[uint]map[uint]uint), []*Node{root}, 0, make(map[string]ID)
	for i := 0; i < 9; i++ {
		for _, node := range nodes {
			s := node.String()
			i, ok := ids[s]
			if !ok {
				i = ID{
					ID:   id,
					Node: node,
				}
				ids[s] = i
				id++
			}
			node.ID = i.ID
			a, b := i.ID, node.Parent
			if a > b {
				a, b = b, a
			}
			parent, ok := graph[a]
			if !ok {
				parent = make(map[uint]uint)
			}
			parent[b]++
			graph[a] = parent
			count++
		}
		nodes = apply(nodes)
	}
	fmt.Println(id, len(graph), count)
}
