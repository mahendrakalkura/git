package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fatih/color"
)

var colorsGreen = color.New(color.FgGreen).SprintfFunc()
var colorsRed = color.New(color.FgRed).SprintfFunc()
var colorsYellow = color.New(color.FgYellow).SprintfFunc()

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
	if !isDirectoryOrFile(path) {
		return nil
	}
	if !isValidDirectory(path) {
		return nil
	}
	directories := strings.Split(path, "/")
	length := len(directories) - 1
	if directories[length] != ".git" {
		return nil
	}
	if directories[length-1] == "hugo-agency-theme" {
		return nil
	}
	if directories[length-1] == "startbootstrap-sb-admin-2" {
		return nil
	}
	if directories[length-2] == "deps" {
		return nil
	}
	directories = directories[:length]
	path = strings.Join(directories, "/")
	waitGroup.Add(1)
	go process(path)
	return nil
}

func process(path string) {
	fmt.Println(path)
	defer waitGroup.Done()
	return

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
		return
	}

	two := "Your branch is ahead"
	if strings.Contains(outputString, two) {
		fmt.Printf("%38s: %s\n", colorsYellow(two), path)
		return
	}

	three := "Changes not staged for commit"
	if strings.Contains(outputString, three) {
		fmt.Printf("%38s: %s\n", colorsRed(three), path)
		return
	}

	fmt.Printf("%38s: %s\n", colorsGreen("All clear"), path)
}

func isDirectoryOrFile(path string) bool {
	stat, statErr := os.Stat(path)
	if statErr != nil {
		return false
	}
	mode := stat.Mode()
	if mode.IsRegular() {
		return true
	}
	if mode.IsDir() {
		return true
	}
	return false
}

func isValidDirectory(path string) bool {
	if strings.Contains(path, "/bitbucket.org/") {
		return true
	}
	if strings.Contains(path, "/cogitosys.com/") {
		return true
	}
	if strings.Contains(path, "/github.com/mahendrakalkura/") {
		return true
	}
	if strings.Contains(path, "/github.com/netenberg/") {
		return true
	}
	if strings.Contains(path, "/github.com/tweetTV/") {
		return true
	}
	return false
}
