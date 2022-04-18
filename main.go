// Copyright 2022 The Ruliad Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"math/cmplx"
	"math/rand"
	"os"
	"sort"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"

	"github.com/pointlander/pagerank"
)

type Value uint64

const (
	// None is the none value
	None Value = iota
	// A is the a value
	A
	// B is the b value
	B
)

var (
	// FlagTruther truther mode
	FlagTruther = flag.Bool("truther", false, "truther mode")
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

// Reduction reduces the matrix
func Reduction(name string, ranks *mat.Dense) {
	var pc stat.PC
	ok := pc.PrincipalComponents(ranks, nil)
	if !ok {
		panic("PrincipalComponents failed")
	}
	k := 2
	var proj mat.Dense
	var vec mat.Dense
	pc.VectorsTo(&vec)
	_, c := ranks.Caps()
	proj.Mul(ranks, vec.Slice(0, c, 0, k))

	fmt.Printf("\n")
	points := make(plotter.XYs, 0, 8)
	for i := 0; i < c; i++ {
		fmt.Println(proj.At(i, 0), proj.At(i, 1))
		points = append(points, plotter.XY{X: proj.At(i, 0), Y: proj.At(i, 1)})
	}

	p := plot.New()

	p.Title.Text = "x vs y"
	p.X.Label.Text = "x"
	p.Y.Label.Text = "y"

	scatter, err := plotter.NewScatter(points)
	if err != nil {
		panic(err)
	}
	scatter.GlyphStyle.Radius = vg.Length(3)
	scatter.GlyphStyle.Shape = draw.CircleGlyph{}
	p.Add(scatter)

	err = p.Save(8*vg.Inch, 8*vg.Inch, fmt.Sprintf("%s.png", name))
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()
	rand.Seed(1)

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

	g := pagerank.NewGraph64()
	adjacency := mat.NewDense(int(id), int(id), nil)
	for a, links := range graph {
		for b, weight := range links {
			g.Link(uint64(a), uint64(b), float64(weight))
			g.Link(uint64(b), uint64(a), float64(weight))
			adjacency.Set(int(a), int(b), float64(weight))
			adjacency.Set(int(b), int(a), float64(weight))
		}
	}
	inverse := make(map[uint]*Node)
	for _, node := range ids {
		inverse[node.ID] = node.Node
	}
	type Rank struct {
		Node *Node
		Rank float64
	}
	ranks := make([]Rank, 0, 8)
	g.Rank(0.85, 0.000001, func(node uint64, rank float64) {
		ranks = append(ranks, Rank{
			Node: inverse[uint(node)],
			Rank: rank,
		})
	})
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].Rank > ranks[j].Rank
	})
	for i := 0; i < 10; i++ {
		fmt.Println(ranks[i].Rank, ranks[i].Node.String())
	}

	if *FlagTruther {
		var eig mat.Eigen
		ok := eig.Factorize(adjacency, mat.EigenRight)
		if !ok {
			panic("Eigendecomposition failed")
		}

		values := eig.Values(nil)
		for i, value := range values {
			fmt.Println(i, value, cmplx.Abs(value), cmplx.Phase(value))
		}

		vectors := mat.CDense{}
		eig.VectorsTo(&vectors)

		ranks := mat.NewDense(int(id), int(id), nil)
		for i := 0; i < int(id); i++ {
			for j := 0; j < int(id); j++ {
				ranks.Set(i, j, real(vectors.At(i, j)))
			}
		}

		Reduction("truther", ranks)
	} else {
		var pc stat.PC
		ok := pc.PrincipalComponents(adjacency, nil)
		if !ok {
			panic("PrincipalComponents failed")
		}
		k := 2
		var proj mat.Dense
		var vec mat.Dense
		pc.VectorsTo(&vec)
		proj.Mul(adjacency, vec.Slice(0, int(id), 0, k))

		points := make(plotter.XYs, 0, 8)
		for i := 0; i < int(id); i++ {
			points = append(points, plotter.XY{X: proj.At(i, 0), Y: proj.At(i, 1)})
		}

		p := plot.New()

		p.Title.Text = "x vs y"
		p.X.Label.Text = "x"
		p.Y.Label.Text = "y"

		scatter, err := plotter.NewScatter(points)
		if err != nil {
			panic(err)
		}
		scatter.GlyphStyle.Radius = vg.Length(1)
		scatter.GlyphStyle.Shape = draw.CircleGlyph{}
		p.Add(scatter)

		err = p.Save(8*vg.Inch, 8*vg.Inch, "adjacency.png")
		if err != nil {
			panic(err)
		}

		output, err := os.Create("adjacency.dat")
		if err != nil {
			panic(err)
		}
		defer output.Close()
		for _, point := range points {
			fmt.Fprintf(output, "%f %f\n", point.X, point.Y)
		}
	}
}
