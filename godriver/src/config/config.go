package config

import (
	"fmt"
	"httpserver"
	"encoding/json"
	"nethandler"
	"io/ioutil"
	"path/filepath"
	"os"
	"strings"

)

/*
 * Reads a configfile that needs to be specified in the variable above.
 * Then mounts the specified images to their corresponding device.
 */
func ReadFile(){
	m := httpserver.ConfigStruct{}
	
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
        if strings.Contains(path, "config.txt"){
        	bs, err := ioutil.ReadFile(path)
		    
		    if err != nil {
		        return nil
		    }
		   
		    fmt.Println("config.txt found, will begin setup...")
		    str := string(bs)
		    fmt.Println(str)

			_ = json.Unmarshal(bs, &m)
        }
        return nil
    })


	for key, value := range m.Mounted {
		temp := m.User[key]
		fmt.Println("Mounting", value.ImageName, "for user", temp.Username, "with password", temp.Password, "to device", value.NbdDevice,".")

		httpserver.AddToMountedList(value.NbdDevice, value.ImageName)
		
		httpserver.LinkedLogins[len(httpserver.LinkedLogins)+1], _ = nethandler.SetupConnection(value.ImageName, temp.Username, temp.Password, value.NbdDevice)
		//if err != nil {
		//	//fmt.Println("Error: %g", err)
		//}
	}
	fmt.Println("Setup complete!")
}
