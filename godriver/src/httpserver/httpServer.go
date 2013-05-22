package httpserver

import (
	"net/http"
	"fmt"
	"strings"
	"nethandler"
	"strconv"
)

var httpAlive = make(chan int)
var lista map[int]string
var listm map[string]string

/*
 * Handler for HTTP-server to check health-status on a JSON-file.
 */
func HttpCheckHealthHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://reddit.com/r/golang.json") //insert json-object here
	if err != nil{
		fmt.Println("Error: %g", err)
	}
	if resp.StatusCode != http.StatusOK{
		fmt.Println(resp.Status)
	}
	fmt.Fprintf(w, "<h1>Health Status</h1>\nStatus: %s", resp.Status)
}

/*
 * Handler for the HTTP-server, takes commands from netdclient and
 * mounts, unmounts devices and image-files as well as printing out
 * available and mounted NBD-devices.
 */
func HttpRootHandler(w http.ResponseWriter, r *http.Request) {
	
	if r.ContentLength < 0 {
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
				if strings.Contains(targetNBD, "/dev/nbd"){
					for i:=0; i<len(lista); i++{
						if lista[i] == targetNBD{
							
							if _, err := nethandler.SetupConnection(targetImg, "", "", targetNBD); err != nil{
								break
							}
							
							listm[lista[i]] = targetImg
							lista[i] = ""
							fmt.Fprintln(w, "Successfully mounted "+targetImg+" to "+targetNBD+"\n")
							return
						}
					}
					for key, value := range lista{
						if value != ""{
							listm[lista[key]] = targetImg
							lista[key] = ""
							break
						}
					}
				}
			break

		case "check":
			fmt.Fprintln(w, "HTTP-Server is online.")
			break

		case "unmount":
			//TODO Real unmounting of NBD-devices
			targetNBD := r.Form["nbd"][0]
			for key, _ := range lista {
				if lista[key] == ""{
					delete(listm, targetNBD)
					lista[key] = targetNBD
					fmt.Fprintln(w, "Successfully unmounted "+targetNBD)
					break
				}
			}
			break

		case "lista":
			fmt.Fprintln(w, "List of all available NBD-devices:")
			for i:=0; i<len(lista); i++{
				if lista[i] != ""{
					fmt.Fprintln(w, lista[i])
				}
			}
			break

		case "listm":
			fmt.Fprintln(w, "List of all mounted NBD-devices:")
			for key, value := range listm{
				fmt.Fprintln(w, key+"\t"+value)
			}
			break

	}
	return
}

func SetupHttp(server string, nrDevices int) (chan int) {
	fmt.Println("HTTP-Server starting on", server)

	lista = make(map[int]string)
	listm = make(map[string]string)

	for i:=0; i<nrDevices; i++{
		lista[i] = ("/dev/nbd"+strconv.Itoa(i))
	}

	http.HandleFunc("/", HttpRootHandler)
	http.HandleFunc("/check-health", HttpCheckHealthHandler)

	go http.ListenAndServe(server, nil)

	fmt.Println("HTTP-Server is up and running!")

	return httpAlive
}