package decisiontrees

import (
	"sync"
)

type CrossValidationFunc func(trainingSet Examples, testingSet Examples) interface{}

func RunCrossValidation(numFolds int, e Examples, f CrossValidationFunc) []interface{} {
	folds := e.crossValidationSamples(numFolds)
	crossValidatedResults := make([]interface{}, numFolds)
	w := sync.WaitGroup{}
	for i, _ := range folds {
		w.Add(1)
		go func(pos int) {
			testingSet := folds[pos]
			trainingSet := make([]*Example, 0, len(e)*(numFolds-1)/numFolds)
			for i, _ := range folds {
				if i != pos {
					trainingSet = append(trainingSet, folds[i]...)
				}
			}

			crossValidatedResults[pos] = f(trainingSet, testingSet)
			w.Done()
		}(i)
	}
	w.Wait()
	return crossValidatedResults
}
