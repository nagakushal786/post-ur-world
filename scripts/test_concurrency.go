package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type UpdatePostPayload struct{
	Title *string `json:"title"`
	Content *string `json:"content"`
}

func updatePost(postID int, p UpdatePostPayload, wg *sync.WaitGroup){
	defer wg.Done()

	url:=fmt.Sprintf("http://localhost:8080/v1/posts/%d", postID)

	// Creating json payload
	b, _:=json.Marshal(p)

	req, err:=http.NewRequest("PATCH", url, bytes.NewBuffer(b))
	if err!=nil{
		fmt.Println(err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	// Sending the request
	client:=&http.Client{}
	resp, err:=client.Do(req)
	if err!=nil{
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	fmt.Println("Update response status:", resp.Status)
}

func main(){
	var wg sync.WaitGroup
	postID:=4

	wg.Add(2)
	content:="New content from user B"
	title:="New title from user A"

	go updatePost(postID, UpdatePostPayload{Title: &title}, &wg)
	go updatePost(postID, UpdatePostPayload{Content: &content}, &wg)
	wg.Wait()
}