package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	directory := flag.String("directory", "", "")
	flag.Parse()
	err := filepath.Walk(*directory, visit)
	if err != nil {
		panic(err)
	}
}

func visit(path string, file_info os.FileInfo, err error) error {
	directories := strings.Split(path, "/")
	length := len(directories) - 1
	if directories[length] != ".git" {
		return nil
	}
	if directories[length-2] == "deps" {
		return nil
	}
	directories = directories[:length]
	path = strings.Join(directories, "/")
	go process(path)
	return nil
}

func process(path string) {
	command := fmt.Sprintf("cd %s && git status", path)
	output_bytes, err := exec.Command("/bin/bash", "-c", command).Output()
	if err != nil {
		panic(err)
	}
	output_string := string(output_bytes)
	output_string = strings.Replace(output_string, "\n", "", -1)
	if strings.Contains(output_string, "nothing to commit, working tree clean") {
		return
	}
	fmt.Println(path)
}