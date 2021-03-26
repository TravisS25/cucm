package app

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

func getCUCM(t *testing.T) *CUCM {
	var err error

	home, err := homedir.Dir()
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Search config in home directory with name ".cucm" (without extension).
	viper.AddConfigPath(home)
	viper.SetConfigName(".cucm")

	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf(err.Error())
	}

	var cucm *CUCM

	if err = viper.Unmarshal(&cucm); err != nil {
		t.Fatalf(err.Error())
	}

	return cucm
}

func TestExecute(t *testing.T) {
	var err error

	cucm := getCUCM(t)

	if err = cucm.Execute(); err != nil {
		t.Fatalf(err.Error())
	}
}

func TestReadXSDFile(t *testing.T) {
	cucm := getCUCM(t)
	xsdFile, err := os.Open(cucm.RootDir + "files/9.0/AXLSoap.xsd")

	if err != nil {
		t.Fatalf(err.Error())
	}

	defer xsdFile.Close()

	fileBytes, err := ioutil.ReadAll(xsdFile)

	if err != nil {
		t.Fatalf(err.Error())
	}

	var schema Schema

	if err = xml.Unmarshal(fileBytes, &schema); err != nil {
		t.Fatalf(err.Error())
	}

	for _, v := range schema.Elements {
		fmt.Printf("element: %+v\n", v)
	}

	t.Errorf("foo")
}

func TestPureBytes(t *testing.T) {
	var err error
	testBytes := []byte(
		`
		<xsd:schema xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:axlapi="http://www.cisco.com/AXL/API/11.0" attributeFormDefault="unqualified" elementFormDefault="unqualified" targetNamespace="http://www.cisco.com/AXL/API/11.0" version="11.0">
			<xsd:element name="addSipProfile" type="axlapi:AddSipProfileReq"/>
		</xsd:schema>
		`,
	)

	var schema Schema

	if err = xml.Unmarshal(testBytes, &schema); err != nil {
		t.Fatalf(err.Error())
	}

	t.Errorf("schema: %+v", schema)
}

// func TestVersion9(t *testing.T) {
// 	ap := cucm9_0.AddPhone{
// 		AddPhoneReq: cucm9_0.AddPhoneReq{
// 			Phone: cucm9_0.XPhone{
// 				Name: "SEP123456",
// 			},
// 		},
// 	}

// 	pr := cucm9_0.AddPhoneReq{
// 		Phone: cucm9_0.XPhone{
// 			Lines: cucm9_0.Anon387{
// 				Line: []cucm9_0.XPhoneLine{
// 					{
// 						Index: "1",
// 					},
// 				},
// 			},
// 		},
// 	}

// 	apBytes, err := xml.Marshal(ap)

// 	if err != nil {
// 		t.Fatalf(err.Error())
// 	}

// 	t.Errorf("ad phone: %s", string(apBytes))
// }
