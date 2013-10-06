package main

import (
	"bytes"
	"code.google.com/p/goprotobuf/proto"
	"encoding/json"
	"flag"
	"fmt"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"strings"
)

var (
	forestPath = flag.String("forest", "forest.json", "")
	// if left/right probability deviates from 0.5 by more than
	// likelyThreshold, then we add a likely/unlikely annotation
	likelyThreshold = flag.Float64("likely_threshold", 0.4, "")

	compilerPath = flag.String("compiler", "/usr/bin/clang++", "")
)

func parseToProto(file string, protobuf proto.Message) error {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(f, protobuf)
}

type codeGenerator struct {
	forest *pb.Forest
}

type codeWriter struct {
	b           bytes.Buffer
	indentLevel int
}

func (c *codeWriter) WriteString(s string) {
	c.b.WriteString(fmt.Sprintf("%v%v", strings.Repeat("  ", c.indentLevel), s))
}

func getAnnotation(node *pb.TreeNode) string {
	if node.GetAnnotation() == nil || node.GetAnnotation().LeftFraction == nil {
		return ""
	}

	if node.GetAnnotation().GetLeftFraction() > 0.5+*likelyThreshold {
		return "LIKELY"
	}

	if node.GetAnnotation().GetLeftFraction() < 0.5-*likelyThreshold {
		return "UNLIKELY"
	}

	return ""
}

func printNode(node *pb.TreeNode, c *codeWriter) {
	if node.GetLeft() == nil && node.GetRight() == nil {
		c.WriteString(fmt.Sprintf("return %v;\n", node.GetLeafValue()))
		return
	}

	c.WriteString(fmt.Sprintf("if (%v(f[%v] < %v)) {\n", getAnnotation(node), node.GetFeature(), node.GetSplitValue()))
	{
		c.indentLevel++
		printNode(node.GetLeft(), c)
		c.indentLevel--
	}
	c.WriteString("} else {\n")
	{
		c.indentLevel++
		printNode(node.GetRight(), c)
		c.indentLevel--
	}
	c.WriteString("}\n")
}

func (c *codeGenerator) generateForest(i int, cw *codeWriter) {
	cw.WriteString(fmt.Sprintf("double evaluateTree%v(const double* f) {\n", i))
	{
		cw.indentLevel++
		printNode(c.forest.GetTrees()[i], cw)
		cw.indentLevel--
	}
	cw.WriteString("}\n")
}

func (c *codeGenerator) generate() string {
	cw := &codeWriter{}
	cw.WriteString("#define LIKELY(x)   (__builtin_expect((x), 1))\n")
	cw.WriteString("#define UNLIKELY(x) (__builtin_expect((x), 0))\n")
	cw.WriteString(`extern "C" {`)
	cw.WriteString("\n")

	for i := range c.forest.GetTrees() {
		c.generateForest(i, cw)
		cw.WriteString("\n")
	}

	// main routine
	cw.WriteString("double evaluate(const double* f) {\n")
	{
		cw.indentLevel++
		cw.WriteString("double result = 0.0;\n")
		cw.WriteString("{\n")
		{
			cw.indentLevel++
			for i := range c.forest.GetTrees() {
				cw.WriteString(fmt.Sprintf("result += evaluateTree%v(f);\n", i))
			}
			cw.indentLevel--
		}
		cw.WriteString("}\n")
		cw.WriteString("return result;\n")
		cw.indentLevel--
	}
	cw.WriteString("}\n")
	cw.WriteString("}\n")
	return cw.b.String()
}

func main() {
	flag.Parse()
	forest := &pb.Forest{}
	if err := parseToProto(*forestPath, forest); err != nil {
		glog.Fatal(err)
	}
	g := codeGenerator{forest}
	glog.Info(g.generate())

	soPath, err := compileTree(forest)
	if err != nil {
		glog.Fatal(err)
	}
	os.Stdout.WriteString(soPath)
}
