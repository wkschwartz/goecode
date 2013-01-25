// Geocode reads in approximate addresses and outputs Google's corrected address.
package main

import (
	"flag"
	"net/url"
	"encoding/csv"
	"encoding/base64"
	"crypto/hmac"
	"testing"
)

const (
	GeocodingAPIVersion = "V2"
 HostURL string = "https://maps.googleapis.com/maps/api/geocode/json"
)

/*
type config struct {
	qps int //queries per second
	key string
	id string
	files []string
}
*/
// GetURL generates an unsigned query URL for the Google Maps API.
// See https://developers.google.com/maps/documentation/geocoding/
func GetURL(address string, sensor bool, client string) (url.URL, error) {
	var result url.URL
	result, err = url.URL.Parse(HostURL)
	if err != nil {
		panic(err) // baseURL is a const so this should never fail.
	}
	var query url.Values = result.Query()
	query.Set("address", url.QueryEscape(address))
	query.Set("sensor", url.QueryEscape(fmt.Sprint(sensor)))
	if client != "" {
		if client[:3] != "gme-" {
			return nil, error.New("invalid client id")
		}
		query.Set("client", url.QueryEscape(client))
	}
	result.RawQuery = query.Encode()
	return result, nil
}

// SignURL uses HMAC+SHA1 to sign the path+query of a URL with a string key.
// See https://developers.google.com/maps/documentation/business/webservices
func SignURL(toSign *url.URL, key string) error {
	var decodeKey []byte
	decodeKey, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		return err
	}
	var urlHash hash.Hash = hmac.New(sha1.New, decodedKey)
	_, err := urlHash.Write(toSign.RequestURI())
	if err != nil {
		return err
	}
	var signature string = base64.URLEncoding.EncodeToString(urlHash.Sum(nil))
	var query url.Values = toSign.Query()
	query.Set("signature", signature)
	toSign.RawQuery = query.Encode()
}

func TestSignURL(t *testing.T) {
	t.Parallel()
	const (
		egURL = "http://maps.googleapis.com/maps/api/geocode/json?address=New+York&sensor=false&client=clientID"
		egKey = "vNIXE0xscrmjlyV-12Nj_BvUPaw="
		egPortionToSign = "/maps/api/geocode/json?address=New+York&sensor=false&client=clientID"
		egSignature = "KrU1TzVQM7Ur0i8i7K3huiw3MsA="
		egSignedURL = "http://maps.googleapis.com/maps/api/geocode/json?address=New+York&sensor=false&client=clientID&signature=KrU1TzVQM7Ur0i8i7K3huiw3MsA="
	)
	var result *url.URL = &url.Parse(egURL)
	err := SignURL(result, egKey)
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