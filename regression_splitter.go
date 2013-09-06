package decisiontrees

import (
	"code.google.com/p/goprotobuf/proto"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"sort"
	"sync"
)

type lossState struct {
	averageLabel         float64
	sumSquaredDivergence float64
	numExamples          int
}

func constructLoss(e Examples) *lossState {
	l := &lossState{}
	for _, ex := range e {
		l.addExample(ex)
	}
	return l
}

func (l *lossState) addExample(e *Example) {
	l.numExamples += 1
	delta := e.Label - l.averageLabel
	l.averageLabel += delta / float64(l.numExamples)
	newDelta := e.Label - l.averageLabel
	l.sumSquaredDivergence += delta * newDelta
}

func (l *lossState) removeExample(e *Example) {
	l.numExamples -= 1
	delta := e.Label - l.averageLabel
	l.averageLabel -= delta / float64(l.numExamples)
	newDelta := e.Label - l.averageLabel
	l.sumSquaredDivergence -= delta * newDelta
}

type RegressionSplitter struct {
	lossFunction         LossFunction
	splittingConstraints pb.SplittingConstraints
}

func (c *RegressionSplitter) shouldSplit(examples Examples, bestGain float64, currentLevel int64) bool {
	maximumLevels := c.splittingConstraints.MaximumLevels
	if maximumLevels != nil && *maximumLevels > currentLevel {
		return false
	}

	minAverageGain := c.splittingConstraints.MinimumAverageGain
	if minAverageGain != nil && *minAverageGain > bestGain/float64(len(examples)) {
		return false
	}

	minSamplesAtLeaf := c.splittingConstraints.MinimumSamplesAtLeaf
	if minSamplesAtLeaf != nil && *minSamplesAtLeaf > int64(len(examples)) {
		return false
	}
	return true
}

func (c *RegressionSplitter) generateTree(examples Examples, currentLevel int64) *pb.TreeNode {
	bestGain, bestFeature, bestValue, bestIndex := 0.0, int64(0), float64(0.0), 0
	leftLoss, rightLoss, totalLoss := constructLoss(Examples{}), constructLoss(examples), constructLoss(examples)

	for _, feature := range examples.getFeatures() {
		// TODO(tulloch) - parallelize this sort
		sort.Sort(ExampleSorter{examples, feature})
		for index, example := range examples {
			func() {
				if index == 0 {
					return
				}

				previousValue := examples[index-1].Features[feature]
				currentValue := example.Features[feature]
				if previousValue == currentValue {
					return
				}

				gain := totalLoss.sumSquaredDivergence - leftLoss.sumSquaredDivergence - rightLoss.sumSquaredDivergence
				if gain > bestGain {
					bestGain = gain
					bestFeature = feature
					bestValue = 0.5 * (previousValue + currentValue)
					bestIndex = index
				}
			}()

			leftLoss.addExample(example)
			rightLoss.removeExample(example)
		}
	}

	if c.shouldSplit(examples, bestGain, currentLevel) {
		sort.Sort(ExampleSorter{examples, bestFeature})

		tree := &pb.TreeNode{
			Feature:    proto.Int64(bestFeature),
			SplitValue: proto.Float64(bestValue),
		}

		// Recur down the left and right branches in parallel
		w := sync.WaitGroup{}
		recur := func(child **pb.TreeNode, e Examples) {
			w.Add(1)
			go func() {
				*child = c.generateTree(e, currentLevel+1)
				w.Done()
			}()
		}

		recur(&tree.Left, examples[bestIndex:])
		recur(&tree.Right, examples[:bestIndex])
		w.Wait()
		return tree
	}

	// Otherwise, return the leaf
	leafWeight := c.lossFunction.GetLeafWeight(examples)
	prior := c.lossFunction.GetPrior(examples)
	return &pb.TreeNode{
		LeafValue: proto.Float64(leafWeight * prior),
	}
}

func (c *RegressionSplitter) GenerateTree(examples Examples) *pb.TreeNode {
	return c.generateTree(examples, 0)
}
