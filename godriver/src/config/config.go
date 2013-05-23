package config

import (
	"io/ioutil"
	"fmt"
	"httpserver"
	"encoding/json"
	"nethandler"
)

var configFile = "/home/alexander/netdriver/godriver/src/config/config.txt"
/*
 * Reads a configfile that needs to be specified in the variable above.
 * Then mounts the specified images to their corresponding device.
 */
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
		fmt.Println("Mounting", value.ImageName, "for user", temp.Username, "with password", temp.Password, "to device", value.NbdDevice,".")

		httpserver.AddToMountedList(value.NbdDevice, value.ImageName)
		
		httpserver.LinkedLogins[len(httpserver.LinkedLogins)+1], err = nethandler.SetupConnection(value.ImageName, temp.Username, temp.Password, value.NbdDevice)
		if err != nil {
			fmt.Println("Error: %g", err)
		}
	}
	fmt.Println("Setup complete!")
}
