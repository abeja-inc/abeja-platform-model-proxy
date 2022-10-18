package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/abeja-inc/platform-model-proxy/config"
	"github.com/abeja-inc/platform-model-proxy/convert"
)

func main() {
	http.HandleFunc("/organizations/1111111111111/deployments/2222222222222/results/req-id-1", handler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("unexpected error occurred", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	conf := config.NewConfiguration()
	cl, err := convert.ToContents(context.TODO(), r, &conf)
	if err != nil {
		fmt.Printf("unexpected error occurred: %s", err.Error())
	}
	fmt.Println()
	fmt.Printf("Request Method: %s\n", cl.Method)
	fmt.Printf("Request Content-Type: %s\n", cl.ContentType)
	fmt.Println("Request Headers:")
	for _, header := range cl.Headers {
		fmt.Printf("\t%s: %s\n", header.Key, strings.Join(header.Values[:], ", "))
	}
	fmt.Printf("content list size: %d\n", len(cl.Contents))
	for i, content := range cl.Contents {
		fmt.Printf("  Content %d:\n", i)
		if content.ContentType != nil {
			fmt.Printf("\tContent-Type: %s\n", *content.ContentType)
		}
		if content.FormName != nil {
			fmt.Printf("\tFormName: %s\n", *content.FormName)
		}
		if content.Path != nil {
			b, _ := ioutil.ReadFile(*content.Path)
			fmt.Printf("\tContent: %s\n", string(b))
		}
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("{\"status\":\"ok\"}")); err != nil {
		log.Fatal("unexpected error occurred", err)
	}
}
