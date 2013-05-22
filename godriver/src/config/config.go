package config

import (
	"io/ioutil"
	"fmt"
	"httpserver"
	"encoding/json"
	"net/url"
	"net/http"
)

var configFile = "/home/alexander/netdriver/godriver/src/config/config.txt"

func ReadFile(){
	m := httpserver.JSONStruct{}
	b, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	_ = json.Unmarshal(b, &m)
	for _, value := range m.Mounted{
		values := make(url.Values)
		values.Set("command", "mount")
		values.Set("nbd", value.NbdDevice)
		values.Set("target", value.ImageName)

		resp, err := http.PostForm("http://localhost:8080/", values)
		
		if err != nil{
			fmt.Println("Error: ", err)
		}
		
		length := resp.ContentLength
		temp := make([]byte, length)
		resp.Body.Read(temp)
		fmt.Print(string(temp))
		fmt.Println("------------------------------------------------")
	}
}