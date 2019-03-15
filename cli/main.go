package main

import (
	"fmt"
	spk "github.com/shua/splunksearch"
	"net/http"
	"os"
)

func usage() {
	fmt.Println("usage: ", os.Args[0], `cmd args...

cmd can be one of:
list-search	list available searches
get-search	get search by name
set-search	set search with values (convenience wrapper for new/update)
new-search	create a new search
update-search	change an existing search`,
	)
}

func main() {
	client := &spk.Client{
		Username: os.Getenv("SPLUNK_USERNAME"),
		Password: os.Getenv("SPLUNK_PASSWORD"),
		Endpoint: os.Getenv("SPLUNK_ENDPOINT"),
		ApiPath:  os.Getenv("SPLUNK_APIPATH"), // eg "/servicesNS/<user>/<index>"
		Client:   &http.Client{},
	}

	fmt.Println(client)

	if len(os.Args) < 2 {
		usage()
		return
	}

	switch os.Args[1] {
	case "list-search", "ls":
		ss, err := client.ListSearches()
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		fmt.Println("len(ss) =", len(ss))
		for _, s := range ss {
			fmt.Println(s)
		}

	case "get-search", "gs":
		s, err := client.GetSearch(os.Args[2])
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		fmt.Println(s)

	case "delete-search", "ds":
		r, err := client.DeleteSearch(os.Args[2])
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		fmt.Println(r)

	case "new-search", "ns":
		search := spk.Search{
			"name":        spk.SType{Str: os.Args[2]},
			"search":      spk.SType{Str: os.Args[3]},
			"description": spk.SType{Str: os.Args[4]},
		}
		s, err := client.NewSearch(search)
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		fmt.Println(s)

	case "update-search", "us":
		search := spk.Search{
			"name":        spk.SType{Str: os.Args[2]},
			"search":      spk.SType{Str: os.Args[3]},
			"description": spk.SType{Str: os.Args[4]},
		}
		s, err := client.UpdateSearch(search)
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		fmt.Println(s)

	case "set-search", "ss":
		search := spk.Search{
			"name":        spk.SType{Str: os.Args[2]},
			"search":      spk.SType{Str: os.Args[3]},
			"description": spk.SType{Str: os.Args[4]},
		}
		s, err := client.SetSearch(search)
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		fmt.Println(s)

	default:
		fmt.Println(os.Args)
	}
}
