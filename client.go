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
	Address string
}

// Project represents the combination of a KGB endpoint with a specific project ID and password (which is the means of "authentication" for the KGB protocol).
type Project struct {
	Client

	ID       string
	Password string
}

// NewClient returns an instance of Client pointing at the referenced address.
func NewClient(address string) *Client {
	return &Client{
		Address: address,
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

// RelayMessage invokes the "relay_message" method in the context of a Project (see https://kgb.alioth.debian.org/kgb-protocol.html#relay_message_message).
func (p Project) RelayMessage(msg string) error {
	return p.jsonRPCExpectOK("relay_message", msg)
}

// CommitV4Args is for Project.CommitV4 (see https://kgb.alioth.debian.org/kgb-protocol.html#commit_v4_arguments).
type CommitV4Args struct {
	// A string identifying the commit in the version control system. Git (short) hash, Subversion revision number, this kind of thing.
	CommitId string `json:"commit_id"`

	// A string to prepend to the commit ID when displaying on IRC. r is particularly useful for Subversion repositories.
	RevPrefix string `json:"rev_prefix"`

	// A string representing the commit author.
	Author string `json:"author"`

	// A string representing the commit branch.
	Branch string `json:"branch"`

	// A string representing the commit module or sub-project.
	Module string `json:"module"`

	// The commit message.
	CommitLog string `json:"commit_log"`

	// List of changes files/directories in the commit. Each string is a path, optionaly prepended with (A) for added paths, (M) for modified paths and (D) for deleted paths. If no prefix is given modification is assumed. An additional plus sign flags property changes (Specific to Subversion term), e.g. (M+).
	Changes []string `json:"changes"`

	// A map with additional parameters. Currently supported members are:
	//
	// web_link: A URL with commit details (e.g. gitweb or viewvc).
	//
	// use_irc_notices: A flag whether to use IRC notices instead of regular messages.
	//
	// use_color: A flag whether to use colors when sending commit notifications. Defaults to 1.
	Extra map[string]interface{} `json:"extra"`
}

// Commitv4 invokes the "commit_v4" method in the context of a Project (see https://kgb.alioth.debian.org/kgb-protocol.html#commit_v4_arguments).
func (p Project) CommitV4(args CommitV4Args) error {
	return p.jsonRPCExpectOK("commit_v4", args)
}

// jsonRPCExpectOK invokes the given JSON-RPC method (as in Project.jsonRPC), but also validates that the result is the string "OK".
func (p Project) jsonRPCExpectOK(method string, params ...interface{}) error {
	res, err := p.jsonRPC(method, params...)
	if err != nil {
		return err
	}

	switch str := res.(type) {
	case string:
		// "relay_message" returns "OK"
		// "commit_v4" returns ""
		if str == "OK" || str == "" {
			return nil
		}
		return fmt.Errorf(`result not "OK": %q`, str)
	default:
		return fmt.Errorf(`unexpected result: %v`, res)
	}
}

// jsonRPC invokes the given JSON-RPC method, using the Project object to "authenticate" it via the "X-KGB-Auth" header.
func (p Project) jsonRPC(method string, params ...interface{}) (interface{}, error) {
	reqJSON, err := json.Marshal(map[string]interface{}{
		"method": method,
		"params": params,
		"id":     42,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", p.Client.Address+"/json-rpc", bytes.NewReader(reqJSON))
	if err != nil {
		return nil, err
	}
	req.Header["Content-Type"] = []string{"application/json"}

	// X-KGB-Project: p.ID
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
	req.Header["X-KGB-Project"] = []string{p.ID}
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
