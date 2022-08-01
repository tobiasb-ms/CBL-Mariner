package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gosuri/uiprogress"
)

type stampedFile struct {
	fileName     string
	currLine     int32
	totalLine    int32
	bar          *uiprogress.Bar
	lastStepDesc string
}

var (
	wg            sync.WaitGroup
	barWidth      = 30
	currProgress  = int32(0) // chose int32 instead of int to use the atomic package
	totalProgress = int32(24)
	isInit        = false // set to true the first time we detec the "init" file in build/timestamp folder.
	targetDir     string
	targetCSV     = []*stampedFile{
		&stampedFile{
			fileName:     "create_worker_chroot.csv",
			currLine:     0,
			totalLine:    5,
			lastStepDesc: "",
		},
		&stampedFile{
			fileName:     "imageconfigvalidator.csv",
			currLine:     0,
			totalLine:    2,
			lastStepDesc: "",
		},
		&stampedFile{
			fileName:     "imagepkgfetcher.csv",
			currLine:     0,
			totalLine:    9,
			lastStepDesc: "",
		},
		&stampedFile{
			fileName:     "imager.csv",
			currLine:     0,
			totalLine:    4,
			lastStepDesc: "",
		},
		&stampedFile{
			fileName:     "roast.csv",
			currLine:     0,
			totalLine:    3,
			lastStepDesc: "",
		},
	}
	// targetJSON = []string{} // for future version
)

func main() {
	fmt.Println("Starting dashboard")
	uiprogress.Start()
	bar := uiprogress.AddBar(int(totalProgress)).AppendCompleted()
	bar.Width = barWidth

	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("%15s: %-25s", "Total", "")
	})

	SetSubBar()

	wd, _ := os.Getwd()
	idx := strings.Index(wd, "CBL-Mariner")
	wd = wd[0 : idx+11]
	targetDir = wd + "/build/timestamp/"

	for {
		time.Sleep(1 * time.Second)
		// fmt.Printf("%d \n", int(currProgress))
		bar.Set(int(currProgress))

		// Check if the target directory exists. Assume we only need to check one directory for now.
		_, err := os.Stat(targetDir)
		if os.IsNotExist(err) {
			continue
		}

		// Check if build has started; if so, add a timestamp to init.csv
		if isInit == false {
			// fmt.Printf("incrementing in checkinit() \n")
			checkInit()
		}

		// Check update for each target file.
		for i, _ := range targetCSV {
			currFile := targetCSV[i]
			// Check if the file exists.
			currStat, err := os.Stat(targetDir + currFile.fileName)
			if os.IsNotExist(err) {
				continue
			}

			// If the file exists, check if there has been any updates since we last visited.
			go currFile.getUpdate(currStat)
		}
	}

}

// Check if the file has been updated, and get updated contents if it did.
// Assumption: the file and its parent directories of the file have been created.
func (file *stampedFile) getUpdate(currStat fs.FileInfo) {
	currNumLines := file.getNumLines()
	if currNumLines != file.currLine {
		atomic.AddInt32(&currProgress, currNumLines-file.currLine)

		file.currLine = currNumLines
		file.bar.Set(int(currNumLines))

		// pop one task off the queue when it's done
		if currNumLines == file.totalLine {
			wg.Done()
		}
	}
}

func (file *stampedFile) getNumLines() int32 {
	currfile, _ := os.Open(targetDir + file.fileName)
	fileScanner := bufio.NewScanner(currfile)
	count := int32(0)
	stepDesc := ""

	for fileScanner.Scan() {
		count += 1
		stepDesc = strings.Split(fileScanner.Text(), ",")[1]
		if count > file.currLine {
			// fmt.Printf("[%d / %d] in %s: %s \n", count, targetCSV[filepath][1], filepath, fileScanner.Text())
		}
	}

	file.lastStepDesc = stepDesc

	return count
}

// Checks if the build has started, and update the progress bar if it did start.
func checkInit() {
	_, err := os.Stat(targetDir + "/init")
	if os.IsNotExist(err) {
		return
	}
	isInit = true
	atomic.AddInt32(&currProgress, 1)
}

func SetSubBar() {
	for i, _ := range targetCSV {
		currFile := targetCSV[i]
		currFile.bar = uiprogress.AddBar(int(currFile.totalLine)).AppendCompleted()
		currFile.bar.Width = barWidth
		currFile.bar.PrependFunc(func(b *uiprogress.Bar) string {
			var tempFileName, tempLastStep string
			tempFileName = currFile.fileName[:len(currFile.fileName)-4]
			if len(currFile.fileName) > 15 {
				tempFileName = tempFileName[:13] + ".."
			}
			if len(currFile.lastStepDesc) > 25 {
				tempLastStep = currFile.lastStepDesc[:23] + ".."
			} else {
				tempLastStep = currFile.lastStepDesc
			}
			return fmt.Sprintf("%15s: %-25s", tempFileName, tempLastStep)
		})
		wg.Add(1)
	}
}