package katsubushi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Client struct {
	server string
}

func NewClient(server string) *Client {
	return &Client{
		server: server,
	}
}

func (c *Client) New() (*IdInfo, error) {
	resp, err := http.PostForm(c.server, url.Values{})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	idInfo := &IdInfo{}
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(idInfo)

	return idInfo, nil
}

func (c *Client) Update(id int) (*IdInfo, error) {
	request, err := http.NewRequest("PUT", fmt.Sprintf("%s/%d", c.server, id), &bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	idInfo := &IdInfo{}
	decoder := json.NewDecoder(resp.Body)
	decoder.Decode(idInfo)

	return idInfo, nil
}

func (c *Client) Delete(id int) error {
	request, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%d", c.server, id), &bytes.Buffer{})
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
