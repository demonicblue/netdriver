package config

import (
	"encoding/json"
	"fmt"
	"httpserver"
	"nethandler"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

/*
 * Reads a config-file that needs to be in the same directory as the ivbsnetd-binary.
 * Then mounts the specified images to their corresponding device.
 */
func ReadFile() {
	m := httpserver.ConfigStruct{}

	pid := os.Getpid()
	link, _ := os.Readlink("/proc/" + strconv.Itoa(pid) + "/exe")

	dir := filepath.Dir(link)
	dir = strings.Replace(dir, "\\", "/", -1)

	bs, err := os.Open(dir + "/config.txt")

	if err != nil {
		fmt.Println("config.txt was not found!")
		return
	}

	fmt.Println("config.txt found, will begin setup...")

	stat, err := bs.Stat()

	if err != nil {
		fmt.Println(err)
		return
	}

	b := make([]byte, stat.Size())
	bs.Read(b)
	_ = json.Unmarshal(b, &m)

	for key, value := range m.Mounted {
		temp := m.User[key]
		fmt.Println("Mounting", value.ImageName, "for user", temp.Username, "to device", value.NbdDevice+".")

		httpserver.AddToMountedList(value.NbdDevice, value.ImageName)

		httpserver.LinkedLogins[len(httpserver.LinkedLogins)+1], _ = nethandler.SetupConnection(value.ImageName, temp.Username, temp.Password, value.NbdDevice)
		if err != nil {
			fmt.Println("Error: %g", err)
		}
	}
	fmt.Println("Setup complete!")
}
