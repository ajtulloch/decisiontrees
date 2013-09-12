package decisiontrees

import (
	"errors"
	"fmt"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"github.com/golang/glog"
)

type Evaluator interface {
	evaluate(features []float64) float64
}

type EvaluatorFunc func(features []float64) float64

func (f EvaluatorFunc) evaluate(features []float64) float64 {
	return f(features)
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

func (f *ForestEvaluator) evaluate(features []float64) float64 {
	sum := 0.0
	for _, t := range f.forest.GetTrees() {
		sum += (&TreeEvaluator{t}).evaluate(features)
	}
	return sum
}

func (t *TreeEvaluator) evaluate(features []float64) float64 {
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

func validateTree(t *pb.TreeNode) error {
	if isLeaf(t) {
		if t.GetLeft() != nil || t.GetRight() != nil {
			return errors.New(fmt.Sprintf("Leaf has non-zero children: %v", t))
		}
		return nil
	}

	// not a leaf - must have both children
	if t.GetLeft() == nil || t.GetRight() == nil {
		return errors.New(fmt.Sprintf("Branch has nil children: %v", t.String()))
	}

	err := validateTree(t.GetLeft())
	if err != nil {
		return err
	}

	err = validateTree(t.GetRight())
	if err != nil {
		return err
	}
	return nil
}

func (f *FastTreeEvaluator) evaluate(features []float64) float64 {
	glog.Info("Evaluating fast tree")
	node := f.nodes[0]
	for node.feature != leafFeatureId {
		glog.Info("Looping inside fast tree")
		if features[node.feature] < node.value {
			node = f.nodes[node.leftChild]
		} else {
			node = f.nodes[node.leftChild+1]
		}
	}
	glog.Info("Finished fast tree")
	return node.value
}

func flattenTree(f *FastTreeEvaluator, current *pb.TreeNode, currentIndex int) {
	glog.Infof("Flattening tree at index %v", currentIndex)
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

func NewFastTreeEvaluator(t *pb.TreeNode) (Evaluator, error) {
	err := validateTree(t)
	if err != nil {
		return nil, err
	}

	f := &FastTreeEvaluator{
		nodes: make([]flatNode, 1),
	}
	flattenTree(f, t, 0)
	return f, nil
}

type FastForestEvaluator struct {
	trees []Evaluator
}

func (f *FastForestEvaluator) evaluate(features []float64) float64 {
	sum := 0.0
	for _, t := range f.trees {
		sum += t.evaluate(features)
	}
	return sum
}

func NewFastForestEvaluator(f *pb.Forest) (Evaluator, error) {
	e := &FastForestEvaluator{
		trees: make([]Evaluator, 0, len(f.GetTrees())),
	}

	for _, t := range f.GetTrees() {
		evaluator, err := NewFastTreeEvaluator(t)
		if err != nil {
			return nil, err
		}
		e.trees = append(e.trees, evaluator)
	}
	return e, nil
}
