package main

import (
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"io/ioutil"
	"os/exec"
)

// returns the filename of the shared object file
func compileTree(f *pb.Forest) (string, error) {
	c := codeGenerator{f}

	codeGenFile, _ := ioutil.TempFile("/tmp", "codegen_tree_file")
	cppO, _ := ioutil.TempFile("/tmp", "codegen_tree_o")
	cppSo, _ := ioutil.TempFile("/tmp", "codegen_tree_so")
	codeGenFile.WriteString(c.generate())

	commands := [][]string{
		{
			"mv",
			codeGenFile.Name(),
			codeGenFile.Name() + ".cpp",
		},
		{
			*compilerPath,
			"-O3",
			codeGenFile.Name() + ".cpp",
			"-c",
			"-o",
			cppO.Name(),
		},
		{
			*compilerPath,
			"-shared",
			cppO.Name(),
			"-dynamiclib",
			"-o",
			cppSo.Name(),
		},
	}

	for _, command := range commands {
		cmd := exec.Command(command[0], command[1:]...)
		err := cmd.Run()
		if err != nil {
			return "", err
		}
	}
	return cppSo.Name(), nil
}
