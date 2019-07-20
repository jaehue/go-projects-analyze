package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	analyze(os.Args[1])
}

type DirInfo struct {
	ExportedStructs   int
	ExportedFuncs     int
	UnexportedStructs int
	UnexportedFuncs   int
}

func (d DirInfo) blank() bool {
	return d.ExportedStructs+d.ExportedFuncs+d.UnexportedStructs+d.UnexportedFuncs == 0
}

func (d *DirInfo) add(v DirInfo) {
	d.ExportedStructs += v.ExportedStructs
	d.ExportedFuncs += v.ExportedFuncs
	d.UnexportedStructs += v.UnexportedStructs
	d.UnexportedFuncs += v.UnexportedFuncs
}

func analyze(dir string) {
	// https://regex101.com/
	unexportedFunc := regexp.MustCompile(`func\s+(\((.)+\))?\s*[a-z]+`)
	exportedFunc := regexp.MustCompile(`func\s+(\((.)+\))?\s*[A-Z]+`)
	unexportedStruct := regexp.MustCompile(`type\s[a-z].+struct`)
	exportedStruct := regexp.MustCompile(`type\s[A-Z].+struct`)

	dirMap := map[string]*DirInfo{}
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if ignore(path) {
			return nil
		}

		f, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		unexportedFuncs := unexportedFunc.FindAll(f, -1)
		exportedFuncs := exportedFunc.FindAll(f, -1)
		unexportedStructs := unexportedStruct.FindAll(f, -1)
		exportedStructs := exportedStruct.FindAll(f, -1)

		dirName := filepath.Dir(path)
		dirName = dirName[len("/Users/jaehuejang/workspace/go/src/github.com/"):]
		if _, ok := dirMap[dirName]; !ok {
			dirMap[dirName] = &DirInfo{}
		}

		dirMap[dirName].add(DirInfo{
			ExportedStructs:   len(exportedStructs),
			ExportedFuncs:     len(exportedFuncs),
			UnexportedStructs: len(unexportedStructs),
			UnexportedFuncs:   len(unexportedFuncs),
		})

		return nil
	})
	display(dirMap)
}

func display(dirs map[string]*DirInfo) {

	const commonHeader1 string = "Public      Public            Private      Private"
	const commonHeader2 string = "Type        Method/Function   Type         Method/Function"
	const defaultOutputSeparator string = "-------------------------------------------------------------------------" +
		"-------------------------------------------------------------------------" +
		"-------------------------------------------------------------------------"
	var (
		maxPathLen int
		total      DirInfo
	)
	for dirName, dirInfo := range dirs {
		if curlen := len(dirName); maxPathLen < curlen {
			maxPathLen = curlen
		}
		total.add(*dirInfo)
	}

	headerLen := maxPathLen + 1
	rowLen := maxPathLen + len(commonHeader2) + 2
	header := "File"
	footer := fmt.Sprintf("AVERAGE(package count: %d)", len(dirs))

	fmt.Printf("%.[2]*[1]s\n", defaultOutputSeparator, rowLen)
	fmt.Printf("%-[2]*[1]s %[3]s\n", header, headerLen, commonHeader1)
	fmt.Printf("%-[2]*[1]s %[3]s\n", "", headerLen, commonHeader2)
	fmt.Printf("%.[2]*[1]s\n", defaultOutputSeparator, rowLen)

	for dirName, dirInfo := range dirs {
		if dirInfo.blank() {
			continue
		}
		fmt.Printf("%-[1]*[2]v %7[3]v %11[4]v %18[5]v %12[6]v\n",
			maxPathLen, dirName, dirInfo.ExportedStructs, dirInfo.ExportedFuncs, dirInfo.UnexportedStructs, dirInfo.UnexportedFuncs)
	}

	fmt.Printf("%.[2]*[1]s\n", defaultOutputSeparator, rowLen)

	fmt.Printf("%-[1]*[2]v %7[3]v %11[4]v %18[5]v %12[6]v\n",
		maxPathLen, footer, total.ExportedStructs/len(dirs), total.ExportedFuncs/len(dirs), total.UnexportedStructs/len(dirs), total.UnexportedFuncs/len(dirs))

	fmt.Printf("%.[2]*[1]s\n", defaultOutputSeparator, rowLen)
}

func ignore(path string) bool {
	if !strings.HasSuffix(path, ".go") {
		return true
	}
	if strings.HasSuffix(path, "test.go") {
		return true
	}

	if strings.Contains(path, "vendor") {
		return true
	}
	if strings.Contains(path, "cmd") {
		return true
	}

	return false
}
