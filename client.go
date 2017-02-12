package kgb // import "go.tianon.xyz/kgb"

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	Addr string
}

type Project struct {
	Client

	Id       string
	Password string
}

func NewClient(addr string) *Client {
	return &Client{
		Addr: addr,
	}
}

func (c Client) Project(id, password string) *Project {
	return &Project{
		Client:   c,
		Id:       id,
		Password: password,
	}
}

func (p Project) RelayMessage(msg string) error {
	res, err := p.jsonRpc("relay_message", msg)
	if err != nil {
		return err
	}

	switch str := res.(type) {
	case string:
		if str != "OK" {
			return fmt.Errorf(`result not "OK": %q`, str)
		}
		return nil
	default:
		return fmt.Errorf(`unexpected result: %v`, res)
	}
}

func (p Project) jsonRpc(method string, params ...interface{}) (interface{}, error) {
	reqJson, err := json.Marshal(map[string]interface{}{
		"method": method,
		"params": params,
		"id":     0,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", p.Client.Addr+"/json-rpc", bytes.NewReader(reqJson))
	if err != nil {
		return nil, err
	}

	req.Header["X-KGB-Project"] = []string{p.Id}

	// TODO X-KGB-Auth: sha1_hex( p.Password + p.Id + reqJson )
	h := sha1.New()
	_, err = io.WriteString(h, p.Password+p.Id)
	if err != nil {
		return nil, err
	}
	_, err = h.Write(reqJson)
	if err != nil {
		return nil, err
	}
	auth := fmt.Sprintf("%x", h.Sum(nil))
	req.Header["X-KGB-Auth"] = []string{auth}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", res.Status)
	}

	decoder := json.NewDecoder(res.Body)
	ret := map[string]interface{}{}
	err = decoder.Decode(&ret)
	if err != nil {
		return nil, err
	}

	if resErr, ok := ret["error"]; ok && resErr != nil {
		return nil, fmt.Errorf("error result: %v", resErr)
	}

	return ret["result"], nil
}
