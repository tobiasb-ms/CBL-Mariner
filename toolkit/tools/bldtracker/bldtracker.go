// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// Tool to initialize a CSV file or record a new timestamp
// for shell scripts during the image-building process.

package main

import (
	"os"

	"github.com/microsoft/CBL-Mariner/toolkit/tools/internal/exe"
	"github.com/microsoft/CBL-Mariner/toolkit/tools/internal/logger"
	"github.com/microsoft/CBL-Mariner/toolkit/tools/internal/timestamp"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	initializeMode = "init"
	recordMode     = "record"
	stopMode       = "stop"
	finishMode     = "finish"
	watchMode      = "watch"
)

var (
	app                     = kingpin.New("bldtracker", "A tool that helps track build time of different steps in the makefile.")
	scriptName              = app.Flag("script-name", "The name of the current tool.").Required().String()
	stepPath                = app.Flag("step-path", "A '/' separated path of steps").Default("").String()
	expectedWeight          = app.Flag("expected-weight", "Expected number of substeps/weight for this new step").Default("0").Float64()
	weight                  = app.Flag("weight", "How much weight will this step have relative to others, defaults to 1.0").Default("1.0").Float64()
	outPath                 = app.Flag("out-path", "The file that stores timestamp CSVs.").Required().String() // currently must be absolute
	logLevel                = exe.LogLevelFlag(app)
	validModes              = []string{initializeMode, recordMode, stopMode, finishMode, watchMode}
	mode                    = app.Flag("mode", "The mode of this tool. Could be 'initialize' ('i') or 'record' ('r').").Required().Enum(validModes...)
	createIfMIssing         = app.Flag("create-if-missing", "Regardless of mode, create a file if its missing").Bool()
	createIfMissingExpected = app.Flag("create-if-missing-expected", "How much weight should the missing root expect").Default("0").Float64()
)

func main() {
	app.Version(exe.ToolkitVersion)
	kingpin.MustParse(app.Parse(os.Args[1:]))
	setupLogger(*logLevel)

	//logger.Log.Warnf("Printing log to '%s'", *outPath)

	// Perform different actions based on the input "mode".
	switch *mode {
	case initializeMode:
		initialize(*outPath, *scriptName, *expectedWeight)

	case recordMode:
		record(*outPath, *scriptName, *stepPath, *expectedWeight, *weight, *createIfMIssing, *createIfMissingExpected)

	case stopMode:
		stop(*outPath, *scriptName, *stepPath)

	case finishMode:
		finish(*outPath, *scriptName)

	case watchMode:
		watch(*outPath)

	default:
		logger.Log.Warnf("Invalid call. Mode must be 'n' for initialize or 'r' for record. ")
	}
}

func setupLogger(logLevelStr string) {
	if logLevelStr == "" {
		logLevelStr = "info"
	}
	logger.InitStderrLog()
	logger.SetStderrLogLevel(logLevelStr)
}

// Creates a JSON specifically for the shell script mentioned in "scriptName".
func initialize(completePath, toolName string, expectedWeight float64) {
	timestamp.BeginTiming(toolName, completePath)
	timestamp.FlushAndCleanUpResources()
}

// Records a new timestamp to the specific JSON for the specified shell script.
func record(completePath, toolName, path string, expectedWeight float64, weight float64, createIfMIssing bool, createIfMissingExpected float64) {
	fullPath := toolName + "/" + path
	timestamp.ResumeTiming(toolName, completePath)
	defer timestamp.FlushAndCleanUpResources()
	timestamp.StartEventByPath(fullPath)
}

func stop(completePath, toolName, path string) {
	fullPath := toolName + "/" + path
	timestamp.ResumeTiming(toolName, completePath)
	defer timestamp.FlushAndCleanUpResources()
	timestamp.StopEventByPath(fullPath)
}

func finish(completePath, toolName string) {
	timestamp.ResumeTiming(toolName, completePath)
	timestamp.CompleteTiming()
}

func watch(completePath string) {
}