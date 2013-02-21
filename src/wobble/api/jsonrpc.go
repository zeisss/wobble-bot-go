package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

// Performs a JSON-RPC 2.0 call to the endpoint
// Copied from https://github.com/ThePiachu/Go-HTTP-JSON-RPC/blob/master/httpjsonrpc/httpjsonrpcClient.go
// Modified by Stephan Zeissler
func JsonRpcCall(address, method string, id interface{}, params interface{}) (interface{}, error) {
	data, err := json.Marshal(map[string]interface{}{
		"version": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	})

	if err != nil {
		return nil, err
	}
	resp, err := http.Post(
		address,
		"application/json",
		strings.NewReader(string(data)),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("ResponseCode is != 200: " + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if result["jsonrpc"] != "2.0" {
		return nil, errors.New("Result is not in JSONRPC 2.0 format")
	}
	if result["error"] != nil {
		var error map[string]interface{} = result["error"].(map[string]interface{})
		code := int(error["code"].(float64))
		msg := error["message"].(string)
		return nil, WobbleApiError{code, msg}
	}

	return result["result"], nil
}
