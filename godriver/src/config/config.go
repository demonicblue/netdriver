package config

import (
	"io/ioutil"
	"fmt"
	"httpserver"
	"encoding/json"
	"nethandler"
)

var configFile = "/home/alexander/netdriver/godriver/src/config/config.txt"

func ReadFile(){
	m := httpserver.ConfigStruct{}
	b, err := ioutil.ReadFile(configFile)

	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	_ = json.Unmarshal(b, &m)

	fmt.Println("config.txt found, will begin setup...")

	for key, value := range m.Mounted {
		temp := m.User[key]
		fmt.Println(value.ImageName, temp.Username, temp.Password, value.NbdDevice)

		httpserver.AddToMountList(value.NbdDevice, value.ImageName)
		
		httpserver.LinkedLogins[len(httpserver.LinkedLogins)+1], err = nethandler.SetupConnection(value.ImageName, temp.Username, temp.Password, value.NbdDevice)
		if err != nil {
			fmt.Println("Error: %g", err)
		}
	}
	fmt.Println("Setup complete!")
}
