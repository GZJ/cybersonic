package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	flag "github.com/spf13/pflag"
)

var address string

func main() {
	var (
		sfx string
	)

	flag.StringVar(&sfx, "sfx", "", "sfx name")
	flag.StringVar(&address, "address", "127.0.0.1:49161", "server address")

	flag.Parse()

	a := flag.Args()
	if len(a) > 0 {
		sfx = a[0]
		getSfx(sfx)
	} else {
		getAll()
	}
}

func getSfx(name string) {
	params := url.Values{}
	params.Add("name", name)

	url := fmt.Sprintf("http://%s/sfx?%s", address, params.Encode())

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

}

func getAll() {
	url := fmt.Sprintf("http://%s/all", address)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}
