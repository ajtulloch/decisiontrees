package decisiontrees

import (
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"sync"
)

type Evaluator interface {
	evaluate(features map[int64]float64) float64
}

type ForestEvaluator struct {
	forest *pb.Forest
}

type TreeEvaluator struct {
	tree *pb.TreeNode
}

func isLeaf(node *pb.TreeNode) bool {
	return node.LeafValue != nil
}

func (f *ForestEvaluator) evaluate(features map[int64]float64) float64 {
	sum := 0.0
	for _, t := range f.forest.GetTrees() {
		sum += (&TreeEvaluator{t}).evaluate(features)
	}
	return sum
}

func (t *TreeEvaluator) evaluate(features map[int64]float64) float64 {
	node := t.tree
	for !isLeaf(node) {
		if features[node.GetFeature()] < node.GetSplitValue() {
			node = node.GetLeft()
		} else {
			node = node.GetRight()
		}
	}
	return node.GetLeafValue()
}

const leafFeatureId = -1

type flatNode struct {
	value     float64
	feature   int64
	leftChild int
}

type FastTreeEvaluator struct {
	nodes []flatNode
}

func (f *FastTreeEvaluator) evaluate(features map[int64]float64) float64 {
	node := f.nodes[0]
	for node.feature != leafFeatureId {
		if features[node.feature] < node.value {
			node = f.nodes[node.leftChild]
		} else {
			node = f.nodes[node.leftChild+1]
		}
	}
	return node.value
}

func flattenTree(f *FastTreeEvaluator, current *pb.TreeNode, currentIndex int) {
	if isLeaf(current) {
		f.nodes[currentIndex] = flatNode{
			value:   current.GetLeafValue(),
			feature: leafFeatureId,
		}
		return
	}

	// append child nodes
	// since we push on N + 2 elements, we want index N + 1, hence len(f.nodes)
	leftChild := len(f.nodes)
	f.nodes = append(f.nodes, flatNode{}, flatNode{})

	f.nodes[currentIndex] = flatNode{
		value:     current.GetSplitValue(),
		feature:   current.GetFeature(),
		leftChild: leftChild,
	}

	flattenTree(f, current.GetLeft(), leftChild)
	flattenTree(f, current.GetRight(), leftChild+1)
}

func NewFastTreeEvaluator(t *pb.TreeNode) Evaluator {
	f := &FastTreeEvaluator{
		nodes: make([]flatNode, 1),
	}
	flattenTree(f, t, 0)
	return f
}

type FastForestEvaluator struct {
	trees []Evaluator
}

type ParallelForestEvaluator struct {
	trees []Evaluator
}

func (p *ParallelForestEvaluator) evaluate(features map[int64]float64) float64 {
	sum := 0.0
	wg := sync.WaitGroup{}
	for _, t := range p.trees {
		wg.Add(1)
		go func(t Evaluator) {
			sum += t.evaluate(features)
			wg.Done()
		}(t)
	}
	wg.Wait()
	return sum
}

func (f *FastForestEvaluator) evaluate(features map[int64]float64) float64 {
	sum := 0.0
	for _, t := range f.trees {
		sum += t.evaluate(features)
	}
	return sum
}

func NewParallelForestEvaluator(f *pb.Forest) Evaluator {
	e := &ParallelForestEvaluator{
		trees: make([]Evaluator, 0, len(f.GetTrees())),
	}
	for _, t := range f.GetTrees() {
		e.trees = append(e.trees, NewFastTreeEvaluator(t))
	}
	return e
}

func NewFastForestEvaluator(f *pb.Forest) Evaluator {
	e := &FastForestEvaluator{
		trees: make([]Evaluator, 0, len(f.GetTrees())),
	}
	for _, t := range f.GetTrees() {
		e.trees = append(e.trees, NewFastTreeEvaluator(t))
	}
	return e
}
