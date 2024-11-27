package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	flag "github.com/spf13/pflag"
)

var address string

func main() {
	var (
		sfx     string
		name    string
		id      int
		message string
		verbose bool
	)
	flag.StringVar(&sfx, "sfx", "", "sfx name")
	flag.StringVar(&name, "name", "", "name to send")
	flag.IntVar(&id, "id", 0, "id to send")
	flag.StringVar(&message, "message", "", "message to send")
	flag.StringVar(&address, "address", "127.0.0.1:49161", "server address")
	flag.BoolVar(&verbose, "verbose", false, "enable verbose output")
	flag.Parse()
	
	args := flag.Args()
	if len(args) > 0 {
		sfx = args[0]
	}
	if len(args) > 1 {
		name = args[1]
	}
	if len(args) > 2 {
		id, _ = fmt.Sscanf(args[2], "%d", &id)
	}
	if len(args) > 3 {
		message = args[3]
	}
	
	if sfx != "" {
		postSfx(sfx, name, id, message, verbose)
	} else {
		getAll()
	}
}

func postSfx(sfxName, name string, id int, message string, verbose bool) {
	body := map[string]interface{}{
		"name":      name,
		"id":        id,
		"message":   message,
		"timestamp": time.Now().Format(time.RFC3339Nano),
	}
	jsonData, err := json.Marshal(body)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}
	url := fmt.Sprintf("http://%s/sfx?name=%s", address, sfxName)
	
	if verbose {
		fmt.Println("Sending POST request:")
		fmt.Printf("URL: %s\n", url)
		fmt.Println("Headers:")
		fmt.Println("  Content-Type: application/json")
		fmt.Println("Body:")
		fmt.Printf("  SFX Name: %s\n", sfxName)
		fmt.Printf("  Name: %s\n", name)
		fmt.Printf("  ID: %d\n", id)
		fmt.Printf("  Message: %s\n", message)
		fmt.Printf("  Timestamp: %s\n", body["timestamp"])
		fmt.Println("JSON Payload:")
		fmt.Println(string(jsonData))
		fmt.Println("-----------------------------------------------------------")
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()
	if verbose {
		fmt.Println("Response:")
		fmt.Printf("Status: %s\n", resp.Status)
		fmt.Println("Headers:")
		for k, v := range resp.Header {
			fmt.Printf("  %s: %v\n", k, v)
		}
		fmt.Println("Body:")
		fmt.Println("-----------------------------------------------------------")
	}
	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
	}
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
