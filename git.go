package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var wait_group sync.WaitGroup

func main() {
	directory := flag.String("directory", "", "")
	flag.Parse()
	err := filepath.Walk(*directory, visit)
	if err != nil {
		panic(err)
	}
	wait_group.Wait()
}

func visit(path string, file_info os.FileInfo, err error) error {
	directories := strings.Split(path, "/")
	length := len(directories) - 1
	if directories[length] != ".git" {
		return nil
	}
	if directories[length-1] == "hugo-agency-theme" {
		return nil
	}
	if directories[length-2] == "deps" {
		return nil
	}
	if strings.Contains(path, "github.com") && !strings.Contains(path, "mahendrakalkura") {
		return nil
	}
	if strings.Contains(path, "golang.org") {
		return nil
	}
	if strings.Contains(path, "gopkg.in") {
		return nil
	}
	directories = directories[:length]
	path = strings.Join(directories, "/")
	wait_group.Add(1)
	go process(path)
	return nil
}

func process(path string) {
	defer wait_group.Done()

	command := fmt.Sprintf("cd %s && /usr/bin/git remote update && /usr/bin/git status", path)

	output_bytes, err := exec.Command("/bin/bash", "-c", command).Output()
	output_string := string(output_bytes)
	output_string = strings.Replace(output_string, "\n", "", -1)
	if err != nil {
		fmt.Println(command)
		fmt.Println(output_string)
		panic(err)
	}

	one := "Your branch is behind"
	if strings.Contains(output_string, one) {
		command := fmt.Sprintf("cd %s && /usr/bin/git pull", path)

		output_bytes, err := exec.Command("/bin/bash", "-c", command).Output()
		output_string := string(output_bytes)
		output_string = strings.Replace(output_string, "\n", "", -1)
		if err != nil {
			fmt.Println(command)
			fmt.Println(output_string)
			panic(err)
		}

		process(path)
	}

	two := "Your branch is ahead"
	if strings.Contains(output_string, two) {
		fmt.Printf("%29s: %s\n", two, path)
	}

	three := "Changes not staged for commit"
	if strings.Contains(output_string, three) {
		fmt.Printf("%29s: %s\n", three, path)
	}
}
