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
	command := fmt.Sprintf("cd %s && /usr/bin/git remote update && /usr/bin/git status", path)

	output_bytes, err := exec.Command("/bin/bash", "-c", command).Output()
	if err != nil {
		panic(err)
	}

	output_string := string(output_bytes)
	output_string = strings.Replace(output_string, "\n", "", -1)
	fmt.Println(output_string)

	one := "Changes not staged for commit"
	if strings.Contains(output_string, one) {
		fmt.Printf("%28s: %s\n", one, path)
	}

	two := "Your branch is ahead"
	if strings.Contains(output_string, two) {
		fmt.Printf("%28s: %s\n", two, path)
	}

	three := "Your branch is behind"
	if strings.Contains(output_string, three) {
		fmt.Printf("%28s: %s\n", three, path)
	}
}
