package decisiontrees

import (
	"flag"
)

var (
	numFeatures = flag.Int("num_features", 100, "")
	numTrees    = flag.Int("num_trees", 5, "")
	numLevels   = flag.Int("num_levels", 2, "")
)
