package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type JSONRequest struct {
	Celsius string `json:"celsius"`
}

type SOAPResponse struct {
	Fahrenheit string `json:"fahrenheit"`
}

func main() {
	http.HandleFunc("/convert", convertHandler)
	fmt.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	// Parse JSON request
	var jsonReq JSONRequest
	if err := json.NewDecoder(r.Body).Decode(&jsonReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Build SOAP request
	soapRequest := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
	<soap12:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
	  <soap12:Body>
	    <CelsiusToFahrenheit xmlns="https://www.w3schools.com/xml/">
	      <Celsius>%s</Celsius>
	    </CelsiusToFahrenheit>
	  </soap12:Body>
	</soap12:Envelope>`, jsonReq.Celsius)

	// Send SOAP request
	resp, err := http.Post("https://www.w3schools.com/xml/tempconvert.asmx", "application/soap+xml; charset=utf-8", bytes.NewBufferString(soapRequest))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read SOAP response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse SOAP response to extract Fahrenheit value
	fahrenheit := extractValueFromXML(string(body), "CelsiusToFahrenheitResult")

	// Convert SOAP response to JSON
	soapResp := SOAPResponse{Fahrenheit: fahrenheit}
	jsonResp, err := json.Marshal(soapResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with JSON
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResp)
}

func extractValueFromXML(xml, tag string) string {
	start := fmt.Sprintf("<%s>", tag)
	end := fmt.Sprintf("</%s>", tag)
	startIdx := bytes.Index([]byte(xml), []byte(start)) + len(start)
	endIdx := bytes.Index([]byte(xml), []byte(end))
	return string(xml[startIdx:endIdx])
}
