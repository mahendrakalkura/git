package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
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

var file = "git.json"

// Timestamps ...
type Timestamps struct {
	sync.RWMutex
	values map[string]int
}

// Get ...
func (timestamps *Timestamps) Get(key string) int {
	timestamps.RLock()
	defer timestamps.RUnlock()
	return timestamps.values[key]
}

// Set ...
func (timestamps *Timestamps) Set(key string, value int) {
	timestamps.Lock()
	defer timestamps.Unlock()
	timestamps.values[key] = value
}

var timestamps = Timestamps{
	values: make(map[string]int),
}

var waitGroup sync.WaitGroup

func init() {
	_, statErr := os.Stat(file)
	if os.IsNotExist(statErr) {
		return
	}
	readFile, readFileErr := ioutil.ReadFile(file)
	if readFileErr != nil {
		panic(readFileErr)
	}
	json.Unmarshal(readFile, &timestamps.values)
}

func term() {
	marshal, marshalErr := json.Marshal(timestamps.values)
	if marshalErr != nil {
		panic(marshalErr)
	}
	ioutil.WriteFile(file, marshal, 0644)
}

func main() {
	defer term()
	directory := flag.String("directory", "", "")
	flag.Parse()
	err := filepath.Walk(*directory, visit)
	if err != nil {
		panic(err)
	}
	waitGroup.Wait()
}

func visit(path string, fileInfo os.FileInfo, err error) error {
	if !strings.Contains(path, "/bitbucket.org/") && !strings.Contains(path, "/github.com/mahendrakalkura/") {
		return nil
	}
	if !isDirectoryOrFile(path) {
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
	defer waitGroup.Done()

	timestampsNew := getTimestampsNew(path)
	timestampsOld := getTimestampsOld(path)
	if timestampsNew <= timestampsOld {
		fmt.Printf("%38s: %s\n", colorsGreen("OK"), path)
		return
	}

	timestamps.Set(path, timestampsNew)

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

	fmt.Printf("%38s: %s\n", colorsGreen("Processed"), path)
}

func getTimestampsNew(path string) int {
	timestamp := 0
	visit := func(path string, fileInfo os.FileInfo, err error) error {
		if strings.HasSuffix(path, "/.git") {
			return nil
		}
		if strings.Contains(path, "/.git/") {
			return nil
		}
		if strings.HasSuffix(path, "/ssh") {
			return nil
		}
		if strings.Contains(path, "/ssh/0_master-") {
			return nil
		}
		if !isDirectoryOrFile(path) {
			return nil
		}
		stat, err := os.Stat(path)
		if err != nil {
			panic(err)
		}
		modTime := stat.ModTime()
		secondsInt64 := modTime.Unix()
		secondsInt := int(secondsInt64)
		if secondsInt > timestamp {
			timestamp = secondsInt
		}
		return nil
	}
	err := filepath.Walk(path, visit)
	if err != nil {
		panic(err)
	}
	return timestamp
}

func getTimestampsOld(path string) int {
	timestamp := timestamps.Get(path)
	return timestamp
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
