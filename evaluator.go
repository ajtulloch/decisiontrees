package decisiontrees

import (
	"fmt"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"github.com/golang/glog"
	"math"
)

// Evaluator implements the evaluator of a decision tree given
// a feature vector
type Evaluator interface {
	Evaluate(features []float64) float64
}

// EvaluatorFunc implements the Evaluator interface
type EvaluatorFunc func(features []float64) float64

// Evaluate is the implementation of Evaluator interface
func (f EvaluatorFunc) Evaluate(features []float64) float64 {
	return f(features)
}

type forestEvaluator struct {
	forest *pb.Forest
}

type treeEvaluator struct {
	tree *pb.TreeNode
}

func isLeaf(node *pb.TreeNode) bool {
	return node.LeafValue != nil
}

func (f *forestEvaluator) Evaluate(features []float64) float64 {
	sum := 0.0
	for _, t := range f.forest.GetTrees() {
		sum += (&treeEvaluator{t}).Evaluate(features)
	}
	return sum
}

func (t *treeEvaluator) Evaluate(features []float64) float64 {
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

const leafFeatureID = -1

type flatNode struct {
	value     float64
	feature   int64
	leftChild int
}

type fastTreeEvaluator struct {
	nodes []flatNode
}

func validateTree(t *pb.TreeNode) error {
	if isLeaf(t) {
		if t.GetLeft() != nil || t.GetRight() != nil {
			return fmt.Errorf("leaf has non-zero children: %v", t)
		}
		return nil
	}

	// not a leaf - must have both children
	if t.GetLeft() == nil || t.GetRight() == nil {
		return fmt.Errorf("branch has nil children: %v", t.String())
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

func (f *fastTreeEvaluator) Evaluate(features []float64) float64 {
	node := f.nodes[0]
	for node.feature != leafFeatureID {
		if features[node.feature] < node.value {
			node = f.nodes[node.leftChild]
		} else {
			node = f.nodes[node.leftChild+1]
		}
	}
	return node.value
}

func flattenTree(f *fastTreeEvaluator, current *pb.TreeNode, currentIndex int) {
	glog.Infof("Flattening tree at index %v", currentIndex)
	if isLeaf(current) {
		f.nodes[currentIndex] = flatNode{
			value:   current.GetLeafValue(),
			feature: leafFeatureID,
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

func newFastTreeEvaluator(t *pb.TreeNode) (Evaluator, error) {
	err := validateTree(t)
	if err != nil {
		return nil, err
	}

	f := &fastTreeEvaluator{
		nodes: make([]flatNode, 1),
	}
	flattenTree(f, t, 0)
	return f, nil
}

type fastForestEvaluator struct {
	trees []Evaluator
}

func (f *fastForestEvaluator) Evaluate(features []float64) float64 {
	sum := 0.0
	for _, t := range f.trees {
		sum += t.Evaluate(features)
	}
	return sum
}

// NewRescaledFastForestEvaluator returns an evalator for a tree
// that automatically corrects for various scaling factors required
// for a given evaluation
func NewRescaledFastForestEvaluator(f *pb.Forest) (Evaluator, error) {
	e := &fastForestEvaluator{
		trees: make([]Evaluator, 0, len(f.GetTrees())),
	}

	for _, t := range f.GetTrees() {
		evaluator, err := newFastTreeEvaluator(t)
		if err != nil {
			return nil, err
		}
		e.trees = append(e.trees, evaluator)
	}

	switch f.GetRescaling() {
	case pb.Rescaling_NONE:
		return e, nil
	case pb.Rescaling_AVERAGING:
		return EvaluatorFunc(func(features []float64) float64 {
			return e.Evaluate(features) / float64(len(e.trees))
		}), nil
	case pb.Rescaling_LOG_ODDS:
		return EvaluatorFunc(func(features []float64) float64 {
			return 1.0 / (1.0 + math.Exp(-2.0*e.Evaluate(features)))
		}), nil
	}

	return nil, fmt.Errorf("unknown rescaling method: %v", f.GetRescaling)
}

// NewFastForestEvaluator returns a flattened tree representation
// used for efficient evaluation
func newUnscaledFastForestEvaluator(f *pb.Forest) (Evaluator, error) {
	e := &fastForestEvaluator{
		trees: make([]Evaluator, 0, len(f.GetTrees())),
	}

	for _, t := range f.GetTrees() {
		evaluator, err := newFastTreeEvaluator(t)
		if err != nil {
			return nil, err
		}
		e.trees = append(e.trees, evaluator)
	}
	return e, nil
}
