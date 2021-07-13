package main

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/shlex"
)

var RawLanguageList LanguageList
var Languages map[string]Language
var Tests []Test
var Cases TestCases

func IsTest(extension string) bool {
	for ext := range Languages {
		if ext == extension {
			return true
		}
	}

	return false
}

func GatherTests() {

	files, err := ioutil.ReadDir(".")

	if err != nil {
		Fatal("Error while reading file list in cwd", err)
	}

	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if !file.IsDir() && IsTest(ext) {

			lang := Languages[ext]
			code, _ := ioutil.ReadFile(file.Name())

			Tests = append(
				Tests,
				Test{
					FileName:  file.Name(),
					Extension: ext,

					Language: lang,

					CompletedCases:  0,
					Passed:          false,
					FinishedTesting: false,

					CodeLength: len(code),
					Time:       0,
				},
			)
		}
	}
}

func GatherLanguages() {
	file, err := ioutil.ReadFile("config/language-config.json")

	if err != nil {
		Fatal("Error while reading config/language-config.json", err)
	}

	err = json.Unmarshal(file, &RawLanguageList)

	if err != nil {
		Fatal("Error while parsing languages list", err)
	}

	Languages = make(map[string]Language, len(RawLanguageList.Languages))

	for _, lang := range RawLanguageList.Languages {
		Languages[lang.Extension] = lang
	}
}

func GatherCases() {
	file, err := ioutil.ReadFile("config/cases.json")

	if err != nil {
		Fatal("Error while reading test cases from config/cases.json", err)
	}

	Cases = TestCases{}
	err = json.Unmarshal(file, &Cases)

	if err != nil {
		Fatal("Error while parsing test cases", err)
	}
}

func CompileFiles() {

	for pos, file := range Tests {

		if file.Language.CompiledLanguage {

			execCommand := strings.ReplaceAll(file.Language.Command, "[f]", file.FileName)
			execCommand = strings.ReplaceAll(execCommand, "[e]", file.Extension)
			execCommand = strings.ReplaceAll(
				execCommand,
				"[n]",
				file.FileName[0:len(file.FileName)-len(file.Extension)],
			)

			Info("Compiling file", file.FileName, "-", execCommand)

			cmd, _ := shlex.Split(execCommand)

			compiler := exec.Command(
				cmd[0],
				cmd[1:]...,
			)

			compiler.Start()
			compiler.Wait()

			file.Compiled = compiler.ProcessState.ExitCode() == 0

			if !file.Compiled {
				Warning(
					"Could not compile file",
					file.FileName,
					"- process errored with",
					compiler.ProcessState.String(),
				)
			} else {
				Success("Successfully compiled", file.FileName)
			}

			Tests[pos] = file
		}
	}
}

func InitTests() {
	GatherLanguages()
	GatherTests()
	GatherCases()
	CompileFiles()
}
