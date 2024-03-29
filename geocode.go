// Geocode reads in approximate addresses and outputs Google's corrected address.
package main

import (
	"crypto/hmac"
	"encoding/base64"
	"encoding/csv"
	"os"
	"strings"
	//	"flag"
	"crypto/sha1"
	"errors"
	"fmt"
	"hash"
	"net/url"
	"testing"
)

const (
	GeocodingAPIVersion = "V2"
	HostURL             = "https://maps.googleapis.com/maps/api/geocode/json"
)

type config struct {
	QPS      int //queries per second
	Key      string
	ClientID string
}

// GetURL generates an unsigned query URL for the Google Maps API.
// See https://developers.google.com/maps/documentation/geocoding/
func GetURL(address string, sensor bool, client string) (*url.URL, error) {
	var result *url.URL
	result, err := url.Parse(HostURL)
	if err != nil {
		panic(err) // baseURL is a const so this should never fail.
	}
	var query url.Values = result.Query()
	query.Set("address", url.QueryEscape(address))
	query.Set("sensor", url.QueryEscape(fmt.Sprint(sensor)))
	if client != "" {
		if client[:3] != "gme-" {
			return nil, errors.New("invalid client id")
		}
		query.Set("client", url.QueryEscape(client))
	}
	result.RawQuery = query.Encode()
	return result, nil
}

// SignURL uses HMAC+SHA1 to sign the path+query of a URL with a string key.
// See https://developers.google.com/maps/documentation/business/webservices
func SignURL(toSign *url.URL, key string) error {
	var decodedKey []byte
	decodedKey, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		return err
	}
	var urlHash hash.Hash = hmac.New(sha1.New, decodedKey)
	_, err = urlHash.Write([]byte(toSign.RequestURI()))
	if err != nil {
		return err
	}
	var signature string = base64.URLEncoding.EncodeToString(urlHash.Sum(nil))
	var query url.Values = toSign.Query()
	query.Set("signature", signature)
	toSign.RawQuery = query.Encode()
	return nil
}

func TestSignURL(t *testing.T) {
	t.Parallel()
	const (
		egURL           = "http://maps.googleapis.com/maps/api/geocode/json?address=New+York&sensor=false&client=clientID"
		egKey           = "vNIXE0xscrmjlyV-12Nj_BvUPaw="
		egPortionToSign = "/maps/api/geocode/json?address=New+York&sensor=false&client=clientID"
		egSignature     = "KrU1TzVQM7Ur0i8i7K3huiw3MsA="
		egSignedURL     = "http://maps.googleapis.com/maps/api/geocode/json?address=New+York&sensor=false&client=clientID&signature=KrU1TzVQM7Ur0i8i7K3huiw3MsA="
	)
	var result *url.URL
	result, err := url.Parse(egURL)
	if err != nil {
		t.Error(err)
	}
	err = SignURL(result, egKey)
	if err != nil {
		t.Error(err)
	}
	var resultQ url.Values = result.Query()
	if resultQ.Get("signature") != egSignature {
		t.Fail()
	}
	if result.String() != egSignedURL {
		t.Fail()
	}
}

/*
func launchQuery(in <-chan url.ULR, out chan<- string, control <-chan float32) {
	var delay float32
	for {
		select {
		case delay = <-control:
			time.Sleep(delay)
		case default:
			next, ok := <-in
			if !ok {
				return
			}
			go doQuery(next, out)
		}
	}
}


type record struct {
     id string
     address string
}

func newCSVReader(filename, delimiter string) (*csv.Reader, *os.PathError) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrailingComma = true
	reader.TrimLeadingSpace = true
	reader.Comma = delimiter
	return reader, err
}

// readIn takes the next file from the input channel and placing the next
// record struct on the output channel. Set delimiter to control the delimter
// in the input data.
func readIn(delimiter string, files <-chan string, records chan-> string) {
	for filename := range files {
		reader, err := newCSVReader(filename)
		if err != nil {
			return err // Can't just return an err?
			// Also need to `defer file.Close()` somehow.

	}
}
*/

// Record structs hold information about the source and content of an address
// record.
type Record struct {
	Source  string // file name
	ID      string
	Sensor  bool
	Address string
}

// NewRecord returns a new Record correctly set based on a row from a CSV like
// unique-id,sensor,address,comp,on,ent,s. filename and n give the origin of
// the record.
func NewRecord(row []string, filename string, n int) (Record, error) {
	var record Record = Record{}
	if len(row) < 3 {
		return record, fmt.Errorf("Not enough columns in file %s after %d records",
			filename, n)
	}
	if row[1] == "1" || strings.ToLower(row[1]) == "true" {
		record.Sensor = true
	} else if row[1] != "0" && strings.ToLower(row[1]) != "false" {
		return record, fmt.Errorf("Expected 'true' or 'false' in file %s after %d records",
			filename, n)
	}
	for _, col := range row[2:] {
		record.Address += col
	}
	record.Source = filename
	record.ID = row[0]
	return record, nil
}

// NewLazyCSVReader returns a new *csv.Reader with many of the checks turned off
func NewLazyCSVReader(f *os.File, delimiter rune) *csv.Reader {
	var reader *csv.Reader
	reader = csv.NewReader(f)
	reader.LazyQuotes = true
	reader.TrailingComma = true
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1
	reader.Comma = delimiter
	return reader
}

// ReadRecords takes a given open CSV file and reads in records and writes them
// to and output channel.
func ReadRecords(f *os.File, delim rune, output chan<- Record) (n int, err error) {
	if output == nil {
		panic("ReadRecords output channel argument is nil.")
	}
	for reader := NewLazyCSVReader(f, delim); row, err := reader.Read(); n++ {
		if err != nil {
			return n, err
		}
		record, err := NewRecord(row, f.Name(), n)
		if err != nil {
			return n, err
		}
		output <- record
	}
	return
}
