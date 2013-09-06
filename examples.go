package decisiontrees

import (
	"math/rand"
)

type Example struct {
	Label    float64
	Features map[int64]float64
}

func (e Example) asBool() bool {
	if e.Label > 0 {
		return true
	}
	return false
}

type Examples []*Example

func (e Examples) Len() int {
	return len(e)
}

func (e Examples) boostrapSample(size int) Examples {
	subsample := make([]*Example, size)
	for i, _ := range subsample {
		subsample[i] = e[i]
	}

	for i := size + 1; i < len(e); i++ {
		j := int(rand.Int31n(int32(i)))
		if j < size {
			subsample[j] = e[i]
		}
	}
	return Examples(subsample)
}

func (e Examples) crossValidationSamples(folds int) []Examples {
	crossValidatedSamples := make([]Examples, folds)
	for i, _ := range crossValidatedSamples {
		crossValidatedSamples[i] = make([]*Example, 0, e.Len()/folds)
	}

	// Do a Fischer-Yates shuffle of the input array
	for i := range e {
		j := rand.Intn(i + 1)
		e[i], e[j] = e[j], e[i]
	}

	for i, ex := range e {
		fold := i % len(crossValidatedSamples)
		crossValidatedSamples[fold] = append(crossValidatedSamples[fold], ex)
	}
	return crossValidatedSamples
}

func (e Examples) boostrapFeatures(size int) []int64 {
	subsample := make([]int64, size)
	allFeatures := e.getFeatures()
	for i, _ := range subsample {
		subsample[i] = allFeatures[i]
	}

	for i := size + 1; i < len(allFeatures); i++ {
		j := int(rand.Int31n(int32(i)))
		if j < size {
			subsample[j] = allFeatures[i]
		}
	}
	return subsample
}

func (e Examples) Swap(i int, j int) {
	e[i], e[j] = e[j], e[i]
}

type ExampleSorter struct {
	Examples
	featureIndex int64
}

func (e ExampleSorter) Less(i int, j int) bool {
	return e.Examples[i].Features[e.featureIndex] < e.Examples[j].Features[e.featureIndex]
}

func (e Examples) getFeatures() []int64 {
	vals := make(map[int64]bool)
	for _, example := range e {
		for k, _ := range example.Features {
			vals[k] = true
		}
	}
	res := make([]int64, 0, len(vals))
	for k, _ := range vals {
		res = append(res, k)
	}
	return res
}
