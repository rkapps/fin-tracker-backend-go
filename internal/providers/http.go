package providers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func RunHTTPGet(url string, in interface{}) error {

	// log.Printf("url: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	err1 := json.Unmarshal(body, &in)
	if err1 != nil {
		return err1
	}

	return nil

}
