package decisiontrees

import (
	pb "github.com/ajtulloch/decisiontrees/protobufs"
)

type Splitter interface {
	GenerateTree(examples Examples) *pb.TreeNode
}
