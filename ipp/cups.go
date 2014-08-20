package ipp

import (
	"os"
	//"io"
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/binary"
	"bytes"
	//"bufio"
	//"code.google.com/p/go.net/html/charset"
)

type CupsServer struct {
	uri            string
	username       string
	password       string
	requestCounter int32
	printername	   string
}

type Response_Header struct {
	Version uint16
	Status 	uint16
	ReqID	uint32
}

type Attribute_Header struct {
	Tag		uint8
	Name 	uint16
	Value 	uint16
}

type Attribute_Name_Length struct {
	Tag		uint8
	NameLen	uint16
}

func (c *CupsServer) SetServer(server string) {
	c.uri = server
}

func (c *CupsServer) SetServerUserName(name string) {
	c.username = name
}

func (c *CupsServer) SetServerUserPassword(password string) {
	c.password = password
}


func (c *CupsServer) SetPrinterName(name string) {
	c.printername = name
}


func (c *CupsServer) CreateRequest(operationId uint16) Message {
	m := newMessage(operationId)
	return m
}
/*
 Octets           Symbolic Value               Protocol field

 0x0100           1.0                          version-number
 0x000A           Get-Jobs                     operation-id
 0x00000123       0x123                        request-id
 0x01             start operation-attributes   operation-attributes-tag
 0x47             charset type                 value-tag
 */

func (c *CupsServer) GetPrinters()(Message, error) {
	m := c.CreateRequest(CUPS_GET_PRINTERS)
	m.AddAttribute(TAG_CHARSET, "attributes-charset", charset("utf-8"))
	m.AddAttribute(TAG_LANGUAGE, "attributes-natural-language", naturalLanguage("en-us"))
	m.AddAttribute(TAG_KEYWORD, "requested-attributes", keyword("printer-name"))
	m.AddAttribute(TAG_ENUM, "printer-type", enum(0))
	m.AddAttribute(TAG_ENUM, "printer-type-mask", enum(1))
	return c.DoRequest(m)
	
}

func (c *CupsServer) GetPrinterAttributes() {
	m := c.CreateRequest(GET_PRINTER_ATTRIBUTES)
	m.AddAttribute(TAG_CHARSET, "attributes-charset", charset("utf-8"))
	m.AddAttribute(TAG_LANGUAGE, "attributes-natural-language", naturalLanguage("en-us"))
	m.AddAttribute(TAG_URI, "printer-uri", uri("ipp://"+c.uri+":631/printers/Canon_iP2700_series"))
	
	a := NewAttribute()
	a.AddValue(TAG_KEYWORD, "requested-attributes", keyword("copies-supported"))
	a.AddValue(TAG_KEYWORD, "", keyword("document-format-supported"))
	a.AddValue(TAG_KEYWORD, "", keyword("printer-is-accepting-jobs"))
	a.AddValue(TAG_KEYWORD, "", keyword("printer-state"))
	a.AddValue(TAG_KEYWORD, "", keyword("printer-state-message"))
	a.AddValue(TAG_KEYWORD, "", keyword("printer-state-reasons"))
	
	m.AppendAttribute(a)
	
	c.DoRequest(m)
	
}

func (c *CupsServer) PrintTestPage(data []byte) {
	m := c.CreateRequest(PRINT_JOB)
	m.AddAttribute(TAG_CHARSET, "attributes-charset", charset("utf-8"))
	m.AddAttribute(TAG_LANGUAGE, "attributes-natural-language", naturalLanguage("en-us"))
	m.AddAttribute(TAG_URI, "printer-uri", uri("ipp://"+c.uri+":631/printers/"+c.printername))
	m.AddAttribute(TAG_KEYWORD, "requesting-user-name", keyword([]byte(c.username)))
	m.Data = data

	c.DoRequest(m)
	
}

func (c *CupsServer) DoRequest(m Message)(Message, error) {

    fii, _ := os.Create("./printer/aC")
	defer fii.Close()
	s := m.marshallMsg()
	fii.Write(s.Bytes())
	// "http://192.168.1.8:631/ipp/printer" "application/ipp"


	//reader := bytes.NewReader(s.Bytes())


	resp, err := http.Post("http://"+c.uri+":631/printers/"+c.printername, "application/ipp", s)
	if err != nil {
		fmt.Println("err: ",err)
	}
  	body, errr := ioutil.ReadAll(resp.Body)
	if errr != nil {
		fmt.Println("errr:   ", errr)
	}
  fmt.Println("Response Body: ", string(body))
  //fmt.Println("Response Body bytes[]: ", body)
  fmt.Println("End Tag: ", TAG_END)
  fmt.Println("Header: ", resp.Header)

	/*****************************************/
	var data_header = Response_Header{}
	binary.Read(bytes.NewReader(body),binary.BigEndian,&data_header)
	//fmt.Println("Response Header:", data_header)

	var data_tag uint8
	binary.Read(bytes.NewReader(body[8:]),binary.BigEndian,&data_tag)
	//fmt.Println("Response Body tag:", data_tag)

	var attrib_header = Attribute_Header{}
	binary.Read(bytes.NewReader(body[8+1:]),binary.BigEndian,&attrib_header)
	//fmt.Println("Attribute Header:", attrib_header)

	var name_len = Attribute_Name_Length{}
	binary.Read(bytes.NewReader(body[8+1+5:]),binary.BigEndian,&name_len)
	//fmt.Println("Attribute Name Length:", name_len)

	count := 8+1+5
	for count < len(body) {
		count = count + 3

		//binary.Read(bytes.NewReader(body[count:]),binary.BigEndian,&name_len)
		//fmt.Println("Attribute Name Length:", name_len)
	}



  	x, eerr := ParseMessage(body)



	for _, ag := range x.attributeGroups {
		for _, ab := range ag.attributes {
			for _, val := range ab.values {
				fmt.Println(val.name, "<->", val.valueTagStr, "<->", val.String())
			}
		}
	}

	//fmt.Println(x.attributeGroups[0].attributes[0].values)
	//fmt.Println(x.attributeGroups[0].attributes[0].values[0].String())
  	//fmt.Println("Message: ", x.attributeGroups[1].attributes[0], len(x.attributeGroups[1].attributes[0].values), "eerr: ", eerr)

	return x, eerr

}
