package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
)

func TestGetPostData(t *testing.T) {
	fmt.Println("start router")
	testRouter := SetupRouter()

	postValues := url.Values{}
	postValues.Set("data", rndStr(4))
	postDataStr := postValues.Encode()
	postDataBytes := []byte(postDataStr)
	postBytesReader := bytes.NewReader(postDataBytes)

	req, err := http.NewRequest("GET", "/data", nil)
	post, err := http.NewRequest("POST", "/data", postBytesReader)
	if err != nil {
		fmt.Println(err)
	}

	wg := &sync.WaitGroup{}
	for count := 0; count < 2; count++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			response := httptest.NewRecorder()
			testRouter.ServeHTTP(response, req)
			testRouter.ServeHTTP(response, post)
			fmt.Println(response.Body)
		}()
	}
	wg.Wait()
}