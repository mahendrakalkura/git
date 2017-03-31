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

var waitGroup sync.WaitGroup

func main() {
	directory := flag.String("directory", "", "")
	flag.Parse()
	err := filepath.Walk(*directory, visit)
	if err != nil {
		panic(err)
	}
	waitGroup.Wait()
}

func visit(path string, fileInfo os.FileInfo, err error) error {
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
	if !strings.Contains(path, "/bitbucket.org/") && !strings.Contains(path, "/github.com/mahendrakalkura.com") {
		return nil
	}
	directories = directories[:length]
	path = strings.Join(directories, "/")
	waitGroup.Add(1)
	go process(path)
	return nil
}

func process(path string) {
	defer waitGroup.Done()

	command := fmt.Sprintf("cd %s && /usr/bin/git remote update && /usr/bin/git status", path)

	outputBytes, err := exec.Command("/bin/bash", "-c", command).Output()
	outputString := string(outputBytes)
	outputString = strings.Replace(outputString, "\n", "", -1)
	if err != nil {
		fmt.Println(command)
		fmt.Println(outputString)
		panic(err)
	}

	one := "Your branch is behind"
	if strings.Contains(outputString, one) {
		command := fmt.Sprintf("cd %s && /usr/bin/git pull", path)

		outputBytes, err := exec.Command("/bin/bash", "-c", command).Output()
		outputString := string(outputBytes)
		outputString = strings.Replace(outputString, "\n", "", -1)
		if err != nil {
			fmt.Println(command)
			fmt.Println(outputString)
			panic(err)
		}

		waitGroup.Add(1)
		go process(path)
	}

	two := "Your branch is ahead"
	if strings.Contains(outputString, two) {
		fmt.Printf("%29s: %s\n", two, path)
	}

	three := "Changes not staged for commit"
	if strings.Contains(outputString, three) {
		fmt.Printf("%29s: %s\n", three, path)
	}
}
