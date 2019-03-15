package splunksearch

import (
	"net/http"
	"net/url"
	"bytes"
	"fmt"
	"encoding/xml"
	"io/ioutil"
)

type Response struct {
	Feed
	MessagesResponse
	httpResponse *http.Response
}

type Search map[string]SType

func (s Search) toStringMap() map[string]string {
	ret := make(map[string]string)
	for k, v := range s {
		switch {
		case v.List != nil, v.Map != nil:
			fmt.Println("don't really know what to do with this one:", k)
		default:
			ret[k] = v.Str
		}
	}
	return ret
}

func (s Search) Encode() []byte {
	ret := url.Values{}
	for k, v := range s {
		switch {
		case v.List != nil:
			for _, i := range v.List {
				switch {
				case i.List != nil, i.Map != nil:
					fmt.Println("don't really know what to do with this one:", k)
				default:
					ret.Add(k, i.Str)
				}
			}
		case v.Map != nil:
			fmt.Println("don't really know what to do with this one:", k)
		default:
			ret.Set(k, v.Str)
		}
	}
	return []byte(ret.Encode())
}

type SError struct {
	StatusCode int
	Status string
	Messages []Message
}

func (err SError) Error() string {
	return fmt.Sprintf("service returned %s: %s", err.Status, err.Messages)
}

type Client struct {
	Username string
	Password string
	Endpoint string
	ApiPath  string
	Client   *http.Client
}

func (c Client) makeRequest(method string, path string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, c.Endpoint+c.ApiPath+path, bytes.NewReader(body))
	req.SetBasicAuth(c.Username, c.Password)

	if err != nil {
		return nil, err
	}
	return req, nil
}

func (c Client) execRequest(r *http.Request) (*Response, error) {
	resp, err := c.Client.Do(r)
	if err != nil {
		return nil, err
	}

	fmt.Println(resp.Status)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	res := Response{httpResponse: resp}
	if err := xml.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *Client) getSearches(r *Response) ([]Search, error) {
	if r.Entries == nil {
		return nil, SError{
			StatusCode: r.httpResponse.StatusCode,
			Status: r.httpResponse.Status,
			Messages: r.Messages,
		}
	}

	fmt.Println(len(r.Entries))
	s0 := Search(r.Entries[0].Content.Map)
	fmt.Println("search[0] = ", s0.toStringMap())

	searches := make([]Search, len(r.Entries))
	for i, e := range r.Entries {
		ss := Search(e.Content.Map)
		ss["name"] = SType{Str: e.Title}
		searches[i] = ss
	}

	return searches, nil
}

func (c Client) ListSearches() ([]Search, error) {
	req, err := c.makeRequest("GET", "/saved/searches", nil)
	if err != nil {
		return nil, err
	}

	res, err := c.execRequest(req)
	if err != nil {
		return nil, err
	}

	searches, err := c.getSearches(res)
	if err != nil {
		return nil, err
	}

	return searches, nil
}

func (c Client) GetSearch(name string) (Search, error) {
	req, err := c.makeRequest("GET", "/saved/searches/"+name, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.execRequest(req)
	if err != nil {
		return nil, err
	}

	searches, err := c.getSearches(res)
	if err != nil {
		return nil, err
	}

	return searches[0], nil
}

func (c Client) DeleteSearch(name string) (*Response, error) {
	req, err := c.makeRequest("DELETE", "/saved/searches/"+name, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.execRequest(req)
	if res.httpResponse.StatusCode == 404 {
		return res, SError{
			Status: res.httpResponse.Status,
			StatusCode: res.httpResponse.StatusCode,
			Messages: res.Messages,
		}
	}
	return res, nil
}

func (c Client) SetSearch(search Search) (Search, error) {
	req, err := c.makeRequest("HEAD", "/saved/searches/"+search["name"].Str, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		fmt.Println("Error doing HEAD")
		return nil, err
	}
	switch resp.StatusCode {
	case 404:
		fmt.Println("Search", search["name"].Str, "doesn't exist, creating new one")
		return c.NewSearch(search)
	case 200:
		fmt.Println("Search", search["name"].Str, "already exists, updating it")
		return c.UpdateSearch(search)
	default:
		return nil, SError{
			Status: resp.Status,
			StatusCode: resp.StatusCode,
		}
	}
}

func (c Client) NewSearch(search Search) (Search, error) {
	req, err := c.makeRequest("POST", "/saved/searches", search.Encode())
	if err != nil {
		return nil, err
	}

	res, err := c.execRequest(req)
	if err != nil {
		return nil, err
	}

	searches, err := c.getSearches(res)
	if err != nil {
		return nil, err
	}

	return searches[0], nil
}

func (c Client) UpdateSearch(search Search) (Search, error) {
	// if the body contains "name" you get an error, so copy out everything except "name"
	body := Search{}
	for k, v := range search {
		if k != "name" {
			body[k] = v
		}
	}

	req, err := c.makeRequest("POST", "/saved/searches/"+search["name"].Str, body.Encode())
	if err != nil {
		return nil, err
	}

	res, err := c.execRequest(req)
	if err != nil {
		return nil, err
	}

	searches, err := c.getSearches(res)
	if err != nil {
		return nil, err
	}

	return searches[0], nil
}

