package decisiontrees

import (
	"code.google.com/p/goprotobuf/proto"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"math"
	"sort"
	"sync"
)

func splitExamples(t *pb.TreeNode, e Examples) (left Examples, right Examples) {
	sort.Sort(ExampleSorter{e, t.GetFeature()})
	splitIndex := 0
	for i, ex := range e {
		splitIndex = i
		if ex.Features[t.GetFeature()] > t.GetSplitValue() {
			break
		}
	}
	left, right = e[:splitIndex], e[splitIndex:]
	return
}

// Returns the mapping of the current node to the new node.
// Returns mapped node and a boolean representing whether we should continue traversal
type TreeMapperFunc func(t *pb.TreeNode, e Examples) (*pb.TreeNode, bool)

func mapTree(t *pb.TreeNode, e Examples, m TreeMapperFunc) *pb.TreeNode {
	left, right := splitExamples(t, e)
	result, continueTraversal := m(t, e)
	if continueTraversal == false {
		return result
	}

	if result.GetLeft() != nil {
		result.Left, _ = m(t.GetLeft(), left)
	}

	if result.GetRight() != nil {
		result.Right, _ = m(t.GetRight(), right)
	}

	return result
}

func weakestLinkCostFunction(t *pb.TreeNode, e Examples) (float64, int) {
	left, right := splitExamples(t, e)
	if !isLeaf(t) {
		leftSquaredDivergence, leftNodes := weakestLinkCostFunction(t.GetLeft(), left)
		rightSquaredDivergence, rightNodes := weakestLinkCostFunction(t.GetRight(), right)
		return leftSquaredDivergence + rightSquaredDivergence, leftNodes + rightNodes
	} else {
		return constructLoss(e).sumSquaredDivergence, 1
	}
}

type Pruner struct {
	pruningConstraints pb.PruningConstraints
	lossFunction       LossFunction
}

type PrunedStage struct {
	alpha float64
	tree  *pb.TreeNode
}

func (p *Pruner) pruneTree(t *pb.TreeNode, e Examples) PrunedStage {
	bestNode, bestCost, bestLeaves := &pb.TreeNode{}, math.MaxFloat64, 0
	mapTree(t, e, TreeMapperFunc(func(n *pb.TreeNode, ex Examples) (*pb.TreeNode, bool) {
		nodeSquaredDivergence, nodeLeaves := weakestLinkCostFunction(n, ex)
		nodeCost := nodeSquaredDivergence / float64(nodeLeaves)
		if nodeCost < bestCost {
			bestNode = t
			bestCost = nodeCost
			bestLeaves = nodeLeaves
		}
		return proto.Clone(n).(*pb.TreeNode), true
	}))

	prunedTree := mapTree(t, e, TreeMapperFunc(func(n *pb.TreeNode, ex Examples) (*pb.TreeNode, bool) {
		if n != bestNode {
			return proto.Clone(n).(*pb.TreeNode), true
		}

		// Otherwise, return the leaf constructed by pruning all subtrees
		leafWeight := p.lossFunction.GetLeafWeight(ex)
		prior := p.lossFunction.GetPrior(ex)
		return &pb.TreeNode{
			LeafValue: proto.Float64(leafWeight * prior),
		}, false
	}))

	rootCost, rootLeaves := weakestLinkCostFunction(t, e)
	alpha := (rootCost - bestCost) / float64(rootLeaves-bestLeaves)
	return PrunedStage{
		alpha: alpha,
		tree:  prunedTree,
	}
}

func (p *Pruner) constructPrunedSequence(originalTree *pb.TreeNode, e Examples) []PrunedStage {
	sequence := make([]PrunedStage, 0)
	sequence = append(sequence, PrunedStage{0.0, originalTree})
	for {
		lastPruned := sequence[len(sequence)-1]
		if isLeaf(lastPruned.tree) {
			break
		}

		sequence = append(sequence, p.pruneTree(lastPruned.tree, e))
	}

	return sequence
}

func (p *Pruner) Prune(t *pb.TreeNode, trainingSet Examples, testingSet Examples) *pb.TreeNode {
	prunedSequence := p.constructPrunedSequence(t, trainingSet)
	result := make([]float64, 0, len(prunedSequence))
	w := sync.WaitGroup{}
	for i, _ := range prunedSequence {
		w.Add(1)
		go func(pos int) {
			rootCost, _ := weakestLinkCostFunction(prunedSequence[pos].tree, testingSet)
			result[i] = rootCost / float64(testingSet.Len())
		}(i)
	}
	w.Done()
	minCost, minCostTree := math.MaxFloat64, &pb.TreeNode{}
	for i, testingCost := range result {
		if testingCost < minCost {
			minCostTree = prunedSequence[i].tree
			minCost = testingCost
		}
	}
	return minCostTree
}
