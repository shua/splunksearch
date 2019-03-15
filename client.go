package splunksearch

import (
	"net/http"
	"net/url"
	"bytes"
	"fmt"
	//"flag"
	"encoding/xml"
	"errors"
	"io/ioutil"
)

type SplunkSearch map[string]SType

func (s *SplunkSearch) toStringMap() map[string]string {
	ret := make(map[string]string)
	for k, v := range *s {
		switch {
		case v.List != nil, v.Map != nil:
			fmt.Println("don't really know what to do with this one:", k)
		default:
			ret[k] = v.Str
		}
	}
	return ret
}

func (s *SplunkSearch) Encode() []byte {
	ret := url.Values{}
	for k, v := range *s {
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

type SplunkClient struct {
	Username string
	Password string
	Endpoint string
	ApiPath  string
	Client   *http.Client
}

func (c SplunkClient) makeRequest(method string, path string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, c.Endpoint+c.ApiPath+path, bytes.NewReader(body))
	req.SetBasicAuth(c.Username, c.Password)

	if err != nil {
		return nil, err
	}
	return req, nil
}

func (c SplunkClient) execRequest(r *http.Request) (*SplunkResponse, error) {
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

	var res SplunkResponse
	if err := xml.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *SplunkClient) getSplunkSearches(r *SplunkResponse) ([]SplunkSearch, error) {
	if r.Entries == nil {
		return nil, errors.New(fmt.Sprint(r.Response))
	}

	fmt.Println(len(r.Entries))
	s0 := SplunkSearch(r.Entries[0].Content.Map)
	fmt.Println("search[0] = ", s0.toStringMap())

	searches := make([]SplunkSearch, len(r.Entries))
	for i, e := range r.Entries {
		ss := SplunkSearch(e.Content.Map)
		ss["name"] = SType{Str: e.Title}
		searches[i] = ss
	}

	return searches, nil
}

func (c SplunkClient) ListSearches() ([]SplunkSearch, error) {
	req, err := c.makeRequest("GET", "/saved/searches", nil)
	if err != nil {
		return nil, err
	}

	res, err := c.execRequest(req)
	if err != nil {
		return nil, err
	}

	searches, err := c.getSplunkSearches(res)
	if err != nil {
		return nil, err
	}

	return searches, nil
}

func (c SplunkClient) GetSearch(name string) (*SplunkSearch, error) {
	req, err := c.makeRequest("GET", "/saved/searches/"+name, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.execRequest(req)
	if err != nil {
		return nil, err
	}

	searches, err := c.getSplunkSearches(res)
	if err != nil {
		return nil, err
	}

	return &searches[0], nil
}

func (c SplunkClient) DeleteSearch(name string) (*SplunkResponse, error) {
	req, err := c.makeRequest("DELETE", "/saved/searches/"+name, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.execRequest(req)
	if res.Entries == nil {
		return nil, errors.New(fmt.Sprint(res.Response))
	}

	return res, nil
}

func (c SplunkClient) SetSearch(search SplunkSearch) (*SplunkSearch, error) {
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
		return nil, errors.New(fmt.Sprint(resp))
	}
}

func (c SplunkClient) NewSearch(search SplunkSearch) (*SplunkSearch, error) {
	req, err := c.makeRequest("POST", "/saved/searches", search.Encode())
	if err != nil {
		return nil, err
	}

	res, err := c.execRequest(req)
	if err != nil {
		return nil, err
	}

	searches, err := c.getSplunkSearches(res)
	if err != nil {
		return nil, err
	}

	return &searches[0], nil
}

func (c SplunkClient) UpdateSearch(search SplunkSearch) (*SplunkSearch, error) {
	// if the body contains "name" you get an error, so copy out everything except "name"
	body := SplunkSearch{}
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

	searches, err := c.getSplunkSearches(res)
	if err != nil {
		return nil, err
	}

	return &searches[0], nil
}

