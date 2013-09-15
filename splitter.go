package decisiontrees

import (
	pb "github.com/ajtulloch/decisiontrees/protobufs"
)

type splitter interface {
	GenerateTree(examples Examples) *pb.TreeNode
}
