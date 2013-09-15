package decisiontrees

import (
	"code.google.com/p/goprotobuf/proto"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"github.com/golang/glog"
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

func (l *lossState) addExample(e *pb.Example) {
	l.numExamples += 1
	delta := e.GetWeightedLabel() - l.averageLabel
	l.averageLabel += delta / float64(l.numExamples)
	newDelta := e.GetWeightedLabel() - l.averageLabel
	l.sumSquaredDivergence += delta * newDelta
}

func (l *lossState) removeExample(e *pb.Example) {
	l.numExamples -= 1
	delta := e.GetWeightedLabel() - l.averageLabel
	l.averageLabel -= delta / float64(l.numExamples)
	newDelta := e.GetWeightedLabel() - l.averageLabel
	l.sumSquaredDivergence -= delta * newDelta
}

type RegressionSplitter struct {
	lossFunction         LossFunction
	splittingConstraints *pb.SplittingConstraints
	shrinkageConfig      *pb.ShrinkageConfig
}

func (c *RegressionSplitter) shouldSplit(
	examples Examples,
	bestSplit split,
	currentLevel int64) bool {
	if len(examples) <= 1 {
		glog.Infof("Num examples is %v, terminating", len(examples))
		return false
	}

	if bestSplit.index == 0 || bestSplit.index == len(examples) {
		glog.Infof("Empty branch with bestSplit = %v, numExamples = %v, terminating", bestSplit, len(examples))
		return false
	}

	maximumLevels := c.splittingConstraints.MaximumLevels
	if maximumLevels != nil && *maximumLevels < currentLevel {
		glog.Infof("Maximum levels is %v < %v currentLevel", *maximumLevels, currentLevel)
		return false
	}

	minAverageGain := c.splittingConstraints.MinimumAverageGain
	if minAverageGain != nil && *minAverageGain > bestSplit.gain/float64(len(examples)) {
		return false
	}

	minSamplesAtLeaf := c.splittingConstraints.MinimumSamplesAtLeaf
	if minSamplesAtLeaf != nil && *minSamplesAtLeaf > int64(len(examples)) {
		return false
	}
	return true
}

type split struct {
	feature int
	index   int
	gain    float64
}

func getBestSplit(examples Examples, feature int) split {
	examplesCopy := make([]*pb.Example, len(examples))
	if copy(examplesCopy, examples) != len(examples) {
		glog.Fatal("Failed copying all examples for sorting")
	}

	By(func(e1, e2 *pb.Example) bool {
		return e1.Features[feature] < e2.Features[feature]
	}).Sort(Examples(examplesCopy))

	leftLoss := constructLoss(Examples{})
	rightLoss := constructLoss(examplesCopy)
	totalLoss := constructLoss(examplesCopy)
	bestSplit := split{
		feature: feature,
	}
	for index, example := range examplesCopy {
		func() {
			if index == 0 {
				return
			}

			previousValue := examplesCopy[index-1].Features[feature]
			currentValue := example.Features[feature]
			if previousValue == currentValue {
				return
			}

			gain := totalLoss.sumSquaredDivergence -
				leftLoss.sumSquaredDivergence -
				rightLoss.sumSquaredDivergence
			if gain > bestSplit.gain {
				bestSplit.gain = gain
				bestSplit.index = index
			}
		}()

		leftLoss.addExample(example)
		rightLoss.removeExample(example)
	}
	return bestSplit
}

func (c *RegressionSplitter) generateTree(examples Examples, currentLevel int64) *pb.TreeNode {
	glog.Infof("Generating tree at level %v with %v examples", currentLevel, len(examples))
	glog.V(2).Infof("Generating with examples %+v", currentLevel, examples)

	features := examples.getFeatures()
	candidateSplits := make(chan split, len(features))
	for _, feature := range features {
		go func(feature int) {
			candidateSplits <- getBestSplit(examples, feature)
		}(feature)
	}

	bestSplit := split{}
	for _, _ = range features {
		candidateSplit := <-candidateSplits
		if candidateSplit.gain > bestSplit.gain {
			bestSplit = candidateSplit
		}
	}

	if c.shouldSplit(examples, bestSplit, currentLevel) {
		glog.Infof("Splitting at level %v with split %v", currentLevel, bestSplit)
		By(func(e1, e2 *pb.Example) bool {
			return e1.Features[bestSplit.feature] < e2.Features[bestSplit.feature]
		}).Sort(examples)

		bestValue := 0.5 * (examples[bestSplit.index-1].Features[bestSplit.feature] +
			examples[bestSplit.index].Features[bestSplit.feature])
		tree := &pb.TreeNode{
			Feature:    proto.Int64(int64(bestSplit.feature)),
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

		recur(&tree.Left, examples[bestSplit.index:])
		recur(&tree.Right, examples[:bestSplit.index])
		w.Wait()
		return tree
	}

	glog.Infof("Terminating at level %v with %v examples", currentLevel, len(examples))
	glog.V(2).Infof("Terminating with examples: %v", examples)
	// Otherwise, return the leaf
	leafWeight := c.lossFunction.GetLeafWeight(examples)
	shrinkage := 1.0
	if c.shrinkageConfig != nil && c.shrinkageConfig.Shrinkage != nil {
		shrinkage = c.shrinkageConfig.GetShrinkage()
	}

	glog.Infof("Leaf weight: %v, shrinkage: %v", leafWeight, shrinkage)
	return &pb.TreeNode{
		LeafValue: proto.Float64(leafWeight * shrinkage),
	}
}

func (c *RegressionSplitter) GenerateTree(examples Examples) *pb.TreeNode {
	return c.generateTree(examples, 0)
}
