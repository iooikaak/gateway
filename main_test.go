package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func BenchmarkSelect(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rpcGetUserInfo()
	}
}

func Req() {
	url := "http://warhorse-gateway.com/ping"
	method := "POST"

	payload := strings.NewReader(`{
   "user_id": 2
}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}

func GetUserInfo() {

	url := "http://warhorse-gateway.com/services/v1/microservice1/getUserinfo"
	method := "POST"

	payload := strings.NewReader(`{
    "user_id": 1
}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Authorization", "081BF2D2E206605EE073EE554732772C96F1A12A")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}

func rpcGetUserInfo(){
	url := "http://warhorse-gateway.com/services/v1/microservice1/rpcgetUserInfo"
	method := "POST"

	payload := strings.NewReader(`{
    "user_id": 1
}`)

	client := &http.Client {
	}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}