package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Rec struct {
	XMLName xml.Name `xml:"Rec"`
	Value   []string `xml:"Value"`
}

type Recs struct {
	XMLName  xml.Name `xml:"Recs"`
	RecItems []Rec    `xml:"Rec"`
}

type Field struct {
	XMLName xml.Name `xml:"Field"`
	Name    string   `xml:"Name"`
}

type Fields struct {
	XMLName    xml.Name `xml:"Fields"`
	FieldItems []Field  `xml:"Field"`
}

type ResultSet struct {
	XMLName      xml.Name `xml:"Table"`
	Name         string   `xml:"Name"`
	Rows         int      `xml:"Rows"`
	ResultFields Fields   `xml:"Fields"`
	ResultRecs   Recs     `xml:"Recs"`
}

type Response struct {
	XMLName      xml.Name `xml:"QueryResponse"`
	Base64String string   `xml:"Data"`
}

func printResult(result ResultSet, printData bool) {
	if printData {
		fieldCnt := len(result.ResultFields.FieldItems)
		for f := 0; f < fieldCnt; f++ {
			fmt.Print(result.ResultFields.FieldItems[f].Name)
			if f+1 < fieldCnt {
				fmt.Print(",")
			}
		}
		fmt.Print("\n")

		for r := 0; r < result.Rows; r++ {
			var rec Rec = result.ResultRecs.RecItems[r]
			for f := 0; f < fieldCnt; f++ {
				fmt.Print(rec.Value[f])
				if f+1 < fieldCnt {
					fmt.Print(",")
				}
			}
			fmt.Print("\n")
		}
	} else {
		fmt.Println(result.Rows)
	}
}

func processResult(result []byte) (ResultSet, error) {

	var respData Response
	var resultSet ResultSet

	xml.Unmarshal(result, &respData)
	// fmt.Println("base64 data:", respData.Base64String)

	decodedXml, err := base64.StdEncoding.DecodeString(respData.Base64String)
	if err != nil {
		fmt.Println("error:", err)
		return resultSet, err
	}
	fmt.Printf("Decoded XML: %q\n", decodedXml)

	xml.Unmarshal(decodedXml, &resultSet)

	return resultSet, nil
}

func doRequest(sql []byte) []byte {
	// fmt.Printf("SQL: %s", sql)

	requestXmlTpl := `<Request>
<Command>DATATRANSFER</Command>
<DataTransferPassword>t1ck3t5!123</DataTransferPassword>
<XmlRS>1</XmlRS>
<DatabaseREQ>
  <Query>%s</Query>
</DatabaseREQ>
</Request>`
	requestXml := fmt.Sprintf(requestXmlTpl, sql)
	fmt.Printf("Request XML:\n%s", requestXml)

	host := "73753.formovietickets.com"
	port := "2235"
	url := fmt.Sprintf("https://%s:%s/rts.asp", host, port)

	req, err := http.NewRequest("POST", url, strings.NewReader(requestXml))
	if err != nil {
		log.Fatal(err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	return body
}

func main() {
	// fmt.Println(len(os.Args), os.Args)
	if len(os.Args) < 2 {
		log.Fatal("Pass SQL file as only parameter")
	}
	var sqlFile = os.Args[1]
	// fmt.Printf("Using SQL file: %s\n", sqlFile)

	printData := false
	if len(os.Args) > 2 {
		var errConv error
		printData, errConv = strconv.ParseBool(os.Args[2])
		if errConv != nil {
			fmt.Println(errConv, ", not printing data")
			printData = false
		}
	}

	dat, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		log.Fatal(err)
	}

	var resultRaw = doRequest(dat)
	var resultSet ResultSet
	resultSet, err = processResult(resultRaw)
	printResult(resultSet, printData)
}
