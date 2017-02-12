package kgb // import "go.tianon.xyz/kgb"

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client represents a single KGB endpoint (expected to accept JSON-RPC commands as outlined in https://kgb.alioth.debian.org/kgb-protocol.html).
type Client struct {
	Addr string
}

// Project represents the combination of a KGB endpoint with a specific project ID and password (which is the means of "authentication" for the KGB protocol).
type Project struct {
	Client

	ID       string
	Password string
}

// NewClient returns an instance of Client pointing at the referenced address.
func NewClient(addr string) *Client {
	return &Client{
		Addr: addr,
	}
}

// Project returns a Project instance using the given project ID and password.
func (c Client) Project(id, password string) *Project {
	return &Project{
		Client:   c,
		ID:       id,
		Password: password,
	}
}

// RelayMessage invokes the "relay_message" method in the context of a Project.
func (p Project) RelayMessage(msg string) error {
	res, err := p.jsonRPC("relay_message", msg)
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

// jsonRPC invokes the given JSON-RPC method, using the Project object to "authenticate" it via the "X-KGB-Auth" header.
func (p Project) jsonRPC(method string, params ...interface{}) (interface{}, error) {
	reqJSON, err := json.Marshal(map[string]interface{}{
		"method": method,
		"params": params,
		"id":     0,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", p.Client.Addr+"/json-rpc", bytes.NewReader(reqJSON))
	if err != nil {
		return nil, err
	}

	req.Header["X-KGB-Project"] = []string{p.ID}

	// X-KGB-Auth: sha1_hex( p.Password + p.ID + reqJSON )
	h := sha1.New()
	_, err = io.WriteString(h, p.Password+p.ID)
	if err != nil {
		return nil, err
	}
	_, err = h.Write(reqJSON)
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
