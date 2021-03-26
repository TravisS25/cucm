package app

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type CUCM struct {
	RootDir string `yaml:"root_dir" mapstructure:"root_dir"`
}

type Schema struct {
	XMLName  xml.Name  `xml:"schema"`
	Elements []Element `xml:"element"`
}

type Element struct {
	XMLName xml.Name `xml:"element"`
	Name    string   `xml:"name,attr"`
	Type    string   `xml:"type,attr"`
}

func (c *CUCM) Execute() error {
	var err error

	filesDir := c.RootDir + "files/"
	dirs, err := ioutil.ReadDir(filesDir)

	if err != nil {
		return errors.WithStack(err)
	}

	// tmpDir := c.RootDir + "tmp/"

	// if err = os.MkdirAll(tmpDir, os.ModePerm); err != nil {
	// 	return errors.WithStack(err)
	// }

	fmt.Printf("reading directories ....")

	for _, v := range dirs {
		if !v.IsDir() {
			return fmt.Errorf(`all entries in "files" directory should be directories`)
		}

		if _, err = strconv.ParseFloat(v.Name(), 32); err != nil {
			fmt.Printf(err.Error())
			return fmt.Errorf("folder names should only be float numbers i.e 10.0 or 10.5")
		}

		cucmVersionDir := filesDir + v.Name() + "/"
		files, err := ioutil.ReadDir(cucmVersionDir)

		if err != nil {
			return errors.WithStack(err)
		}

		if len(files) != 1 {
			return fmt.Errorf("directory '%s' should contain 1 file", v.Name())
		}

		if files[0].IsDir() {
			return fmt.Errorf("directory '%s' should not contain another directory", v.Name())
		}

		if files[0].Name() != "AXLSoap.xsd" {
			return fmt.Errorf("directory '%s' should contain file named '%s'", v.Name(), "AXLSoap.xsd")
		}

		axlFilePath := cucmVersionDir + "AXLSoap.xsd"
		packageName := "cucm" + strings.Replace(v.Name(), ".", "_", -1)
		goCodeDir := c.RootDir + packageName

		if err = os.MkdirAll(goCodeDir, os.ModePerm); err != nil {
			return errors.WithStack(err)
		}

		goCodeFile := goCodeDir + "/cucm.go"

		cmdStr := fmt.Sprintf("xsdgen -o %s -pkg %s %s", goCodeFile, packageName, axlFilePath)
		stdErr := &bytes.Buffer{}
		genCmd := exec.Command("/bin/sh", "-c", cmdStr)
		genCmd.Stderr = stdErr

		fmt.Printf("generating go file '%s.go' ....", v.Name())

		if err = genCmd.Run(); err != nil {
			return errors.WithStack(fmt.Errorf(stdErr.String()))
		}

		fmt.Printf("finished generating go file '%s.go' ....", v.Name())

		axlFile, err := os.Open(axlFilePath)

		if err != nil {
			return errors.WithStack(err)
		}

		defer axlFile.Close()

		axlBytes, err := ioutil.ReadAll(axlFile)

		if err != nil {
			return errors.WithStack(err)
		}

		var schema Schema

		if err = xml.Unmarshal(axlBytes, &schema); err != nil {
			return errors.WithStack(err)
		}

		goFile, err := os.OpenFile(goCodeFile, os.O_APPEND|os.O_WRONLY, os.ModePerm)

		if err != nil {
			return errors.WithStack(err)
		}

		defer goFile.Close()

		bufWriter := bufio.NewWriter(goFile)

		for _, v := range schema.Elements {
			wrapperTypeSlice := strings.Split(v.Type, ":")

			if len(wrapperTypeSlice) == 1 {
				continue
			}

			if wrapperTypeSlice[0] != "axlapi" {
				continue
			}

			line1 := fmt.Sprintf("\ntype %s struct { \n", strings.Title(v.Name))
			line2 := fmt.Sprintf("\tXMLName xml.Name `xml:\"ns:%s\"` \n", v.Name)
			line3 := "\t" + strings.Title(wrapperTypeSlice[1]) + "\n"
			line4 := "}\n"

			bufWriter.WriteString(line1)
			bufWriter.WriteString(line2)
			bufWriter.WriteString(line3)
			bufWriter.WriteString(line4)
		}

		if err = bufWriter.Flush(); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
