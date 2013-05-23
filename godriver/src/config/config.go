package config

import (
	"io/ioutil"
	"fmt"
	"httpserver"
	"encoding/json"
	//"nethandler"
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

		httpserver.Listm[value.NbdDevice] = value.ImageName
		
		/* session, err := nethandler.SetupConnection(value.ImageName, value2.Username, value2.Password, value.NbdDevice)
		if err != nil {
			fmt.Println("Error: %g", err)
		}

		httpserver.Listm[value.NbdDevice] = value.ImageName

		httpserver.LinkedLogins.Id = session.Id
		httpserver.LinkedLogins.Image = session.Image
		httpserver.LinkedLogins.Username = session.Username
		httpserver.LinkedLogins.Passwd = session.Passwd
		httpserver.LinkedLogins.NbdPath = session.NbdPath
		*/
	}
	fmt.Println("Setup complete!")
}
