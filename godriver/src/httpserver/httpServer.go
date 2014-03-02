package httpserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"nethandler"
	"strconv"
	"strings"
)

var httpAlive = make(chan int)
var AvailableList []string
var MountedList map[string]string
var LinkedLogins map[int]nethandler.IVBSSession

const lenPath = len("/status") + 1

/*
 * JSON-struct for client when sending commands
 */
type CommandStruct struct {
	Command string
	Device  string
	Image   string
	User    string
	Pass    string
}

/*
 * Struct for a NBDDevice
 */
type NBDStruct struct {
	NbdDevice string
	ImageName string
}

/*
 * Struct for a user
 */
type UserStruct struct {
	Username string
	Password string
}

/*
 * Struct used by config
 */
type ConfigStruct struct {
	Mounted []NBDStruct
	User    []UserStruct
}

/*
 * Struct used when printing out JSON-data in the status-handle.
 */
type JSONStruct struct {
	Mounted []NBDStruct
}

/*
 * Handler for HTTP-server to check health-status on a JSON-file.
 */
func HttpCheckHealthHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://reddit.com/r/golang.json") //insert json-object here
	if err != nil {
		fmt.Println("Error: %g", err)
	}
	fmt.Fprintf(w, "<h1>Health Status</h1>\nStatus: %s", resp.Status)
}

/*
 * Handler to list all mounted NBD-devices in a webbrowser.
 */
func HttpStatusHandler(w http.ResponseWriter, r *http.Request) {
	m := JSONStruct{}
	temp := []NBDStruct{}
	if checkJSON := r.URL.Path[lenPath:]; strings.Contains(checkJSON, "json") {
		for key, value := range MountedList {
			temp = append(temp, NBDStruct{key, value})
		}
		m.Mounted = temp
		b, _ := json.Marshal(m)
		fmt.Fprintf(w, string(b))
	} else {
		fmt.Fprintf(w, "Mounted NBD-devices:\n\n")
		for key, value := range MountedList {
			fmt.Fprintln(w, key+"\t"+value+"\n")
		}
	}
}

/*
 * Handler for the HTTP-server, takes commands from netdclient and
 * mounts, unmounts devices and image-files as well as printing out
 * available and mounted NBD-devices.
 */
func HttpRootHandler(w http.ResponseWriter, r *http.Request) {
	cmd := CommandStruct{}

	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	_ = json.Unmarshal(bs, &cmd)
	switch cmd.Command {
	case "check":
		fmt.Fprintln(w, "HTTP-Server is online.")
		break

	case "disc":
		fmt.Fprintln(w, "HTTP-Server shutting down...")
		httpAlive <- 1
		break

	case "lista":
		fmt.Fprintln(w, "List of all available NBD-devices:")
		for key, value := range AvailableList {
			if AvailableList[key] != "" {
				fmt.Fprintln(w, value)
			}
		}
		break

	case "listm":
		fmt.Fprintln(w, "List of all mounted NBD-devices:")
		for key, value := range MountedList {
			fmt.Fprintln(w, key+"\t"+value)
		}
		break

	case "mount":
		if strings.Contains(cmd.Device, "/dev/nbd") {
			for i := 0; i < len(AvailableList); i++ {
				if AvailableList[i] == cmd.Device {

					LinkedLogins[len(LinkedLogins)+1], err = nethandler.SetupConnection(cmd.Image, cmd.User, cmd.Pass, cmd.Device)
					if err != nil {
						fmt.Fprintf(w, "Error: ", err)
						fmt.Fprintf(w, "\n")
						return
					}

					AddToMountedList(cmd.Device, cmd.Image)
					fmt.Fprintf(w, "Successfully mounted "+cmd.Image+" to "+cmd.Device+"\n")
					return
				}
			}
			for _, value := range AvailableList {
				if value != "" {
					AddToMountedList(value, cmd.Image)
					fmt.Fprintf(w, "Device "+cmd.Device+" is already mounted.\n"+cmd.Image+" has been mounted to "+value+" instead.\n")
					return
				}
			}
			fmt.Fprintf(w, "No more devices available!\n")
		} else {
			fmt.Fprintf(w, "Specified device not recognised.\n")
		}
		break

	case "unmount":
		//TODO Real unmounting of NBD-devices
		for key, _ := range AvailableList {
			if AvailableList[key] == "" {
				delete(MountedList, cmd.Device)
				AvailableList[key] = cmd.Device
				fmt.Fprint(w, "Successfully unmounted "+cmd.Device)
				break
			}
		}
		break
	}
}

/*
 * Help-function to add NBD-device to MountedList
 */
func AddToMountedList(nbd, img string) {
	MountedList[nbd] = img
	for key, value := range AvailableList {
		if value == nbd {
			AvailableList[key] = ""
			break
		}
	}
}

func SetupHttp(server string, nrDevices int) chan int {
	fmt.Println("HTTP-Server starting on", server)

	AvailableList = make([]string, nrDevices)
	MountedList = make(map[string]string)
	LinkedLogins = make(map[int]nethandler.IVBSSession)

	for i := 0; i < nrDevices; i++ {
		AvailableList[i] = ("/dev/nbd" + strconv.Itoa(i))
	}

	http.HandleFunc("/", HttpRootHandler)
	http.HandleFunc("/status/", HttpStatusHandler)
	http.HandleFunc("/check-health", HttpCheckHealthHandler)

	go http.ListenAndServe(server, nil)

	fmt.Println("HTTP-Server is up and running!")

	return httpAlive
}
