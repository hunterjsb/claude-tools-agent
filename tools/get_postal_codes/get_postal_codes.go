package main

import (
	"errors"
	"fmt"
	"net/http"
)

func GET_POSTAL_CODES(params map[string]any) error {
	postalCode, ok := params["postal_code"]
	if !ok {
		return errors.New("must provide postal_code")
	}
	url := fmt.Sprintf("http://localhost:8280/zipcodes/%s", postalCode)
	fmt.Println("URL", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	fmt.Println("GOT RESPONSE! YES!", resp.StatusCode)
	return nil
}
