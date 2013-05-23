package httpserver

import (
	"net/http"
	"fmt"
	"strings"
	"nethandler"
	"strconv"
	"encoding/json"
)

var httpAlive = make(chan int)
var Lista map[int]string
var Listm map[string]string
var LinkedLogins map[int]nethandler.IVBSSession

const lenPath = len("/status")+1

type NBDStruct struct {
	NbdDevice string
	ImageName string
}

type UserStruct struct{
	Username string
	Password string
}

type ConfigStruct struct{
	Mounted []NBDStruct
	User 	[]UserStruct
}

type JSONStruct struct {
	Mounted []NBDStruct
}

/*
 * Handler for HTTP-server to check health-status on a JSON-file.
 */
func HttpCheckHealthHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://reddit.com/r/golang.json") //insert json-object here
	if err != nil{
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
 		for key, value := range Listm {
 			temp = append(temp, NBDStruct{key, value})
		}	
			m.Mounted = temp
 			b, _ := json.Marshal(m)
			fmt.Fprintf(w, string(b))
 	} else {
 		fmt.Fprintf(w, "Mounted NBD-devices:\n\n")
		for key, value := range Listm {
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
	
	if r.ContentLength <= 0 {
		return
	}
	
	err := r.ParseForm()
	if err != nil {
		fmt.Println("Error:%g", err)
	}

	cmd := r.Form["command"][0]

	switch cmd{

		case "disc":
			fmt.Fprintln(w, "HTTP-Server shutting down...")
			httpAlive <- 1
			break

		case "mount":
			//TODO Real mounting to NBD-devices with real images
				targetNBD := r.Form["nbd"][0]
				targetImg := r.Form["target"][0]
				userName := r.Form["user"][0]
				passWord := r.Form["pass"][0]
				if strings.Contains(targetNBD, "/dev/nbd"){
					for i:=0; i<len(Lista); i++{
						if Lista[i] == targetNBD{
							
							LinkedLogins[len(LinkedLogins)+1], err = nethandler.SetupConnection(targetImg, userName, passWord, targetNBD);
							if err != nil{
								fmt.Println("Error: ",err)
								break
							}

							AddToMountList(targetNBD, targetImg)
							fmt.Fprintf(w, "Successfully mounted "+targetImg+" to "+targetNBD+"\n"+userName+" "+passWord+"\n")
							return
						}
					}
					for _, value := range Lista{
						if value != ""{
							AddToMountList(value, targetImg)
							fmt.Fprintf(w, "Device "+targetNBD+" is already mounted.\n"+targetImg+" has been mounted to "+value+" instead.\n")
							break
						}
					}
					fmt.Fprintf(w, "No more devices available!\n")
				}
				fmt.Fprintf(w, "Specified device not recognised.")
			break

		case "check":
			fmt.Fprintln(w, "HTTP-Server is online.")
			break

		case "unmount":
			//TODO Real unmounting of NBD-devices
			targetNBD := r.Form["nbd"][0]
			for key, _ := range Lista {
				if Lista[key] == ""{
					delete(Listm, targetNBD)
					Lista[key] = targetNBD
					fmt.Fprint(w, "Successfully unmounted "+targetNBD)
					break
				}
			}
			break

		case "lista":
			fmt.Fprintln(w, "List of all available NBD-devices:")
			for _, value := range Lista{
				if value != ""{
					fmt.Fprintln(w, value)
				}
			}
			break

		case "listm":
			fmt.Fprintln(w, "List of all mounted NBD-devices:")
			for key, value := range Listm{
				fmt.Fprintln(w, key+"\t"+value)
			}
			break

	}
	return
}

func AddToMountList(nbd, img string){
	Listm[nbd] = img
	for key, value := range Lista{
		if value == nbd{
			Lista[key] = ""
			break
		}
	}
}

func SetupHttp(server string, nrDevices int) (chan int) {
	fmt.Println("HTTP-Server starting on", server)

	Lista = make(map[int]string)
	Listm = make(map[string]string)
	LinkedLogins = make(map[int]nethandler.IVBSSession)

	for i:=0; i<nrDevices; i++{
		Lista[i] = ("/dev/nbd"+strconv.Itoa(i))
	}

	http.HandleFunc("/", HttpRootHandler)
	http.HandleFunc("/status/", HttpStatusHandler)
	http.HandleFunc("/check-health", HttpCheckHealthHandler)

	go http.ListenAndServe(server, nil)

	fmt.Println("HTTP-Server is up and running!")

	return httpAlive
}