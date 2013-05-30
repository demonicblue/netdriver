package main

import (
	"fmt"
	"net/http"
	"flag"
	"encoding/json"
	"strings"
	"httpserver"
)

// Global-variables used within the client
var server, serverAdress, menu string

/*
 * Sends a POST-form to the HTTP-server
 * checking the connection-status.
 *
 * returns boolean for status
 */
func checkConnection() bool{
	Cmd := httpserver.CommandStruct{}
	Cmd.Command = "check"

	bytes, _ := json.Marshal(Cmd)
	JSONData := strings.NewReader(string(bytes)) 
	_, err := http.Post(serverAdress, "application/json", JSONData)

    if err != nil {
            fmt.Println("HTTP-server is offline.")
            return false
    }
	return true
}

/*
 * Takes a Response from the HTTP-server
 * and prints it out.
 */
func readResponse(resp *http.Response){
	length := resp.ContentLength
	
	if length == -1{
		length = 1024
	}

	slice := make([]byte, length)
	
	resp.Body.Read(slice)
	fmt.Print(string(slice))
	
	fmt.Println("------------------------------------------------")
}
/*
 * Sends a JSON-request to the HTTP-server with commands
 * for the server to perform.
 */
func sendJSONRequest(cmd string) *http.Response{
	Cmd := httpserver.CommandStruct{}
	Cmd.Command = cmd

	if cmd == "mount" || cmd == "unmount" {
		fmt.Println("Type in your target NBD-device")
		_, _ = fmt.Scan(&Cmd.Device)

		if cmd == "mount" {
			fmt.Println("Type in your target image for",Cmd.Device)
			_, _ = fmt.Scan(&Cmd.Image)
			fmt.Println("Type in your username")
			_, _ = fmt.Scan(&Cmd.User)
			fmt.Println("Type in your password")
			_, _ = fmt.Scan(&Cmd.Pass)
		}
	}

	bytes, _ := json.Marshal(Cmd)
	JSONData := strings.NewReader(string(bytes)) 
	resp, _ := http.Post(serverAdress, "application/json", JSONData)
	return resp
}

/*
 * Netdriver-Client main function.
 * The mainmenu for netdclient, takes a command
 * and sends the request to the HTTP-server for IVBS-Netdriver.
 */
func main(){
	fmt.Println("Netdriver-Client started!\nNeed any help? Type help for available commands.")

	flag.StringVar(&server, "c", "localhost:8080", "IP-address to HTTP-server")
	flag.Parse()
	serverAdress = "http://"+server+"/"

	for{
		_, err := fmt.Scan(&menu)
		
		if err != nil{
			fmt.Println(err)
		}

		switch menu {
			default:
				fmt.Println("Unknown command! Try type 'help' for available commands.")

			case "mount", "unmount","lista", "listm", "disc":
				if checkConnection() != true {
					break
				}
				resp := sendJSONRequest(menu)
		       	readResponse(resp)
				break

			case "check":
				if checkConnection(){
					fmt.Println("HTTP-server is online.")
					fmt.Println("------------------------------------------------")
				}
				break

			case "help":
				fmt.Println("\nCommands available:")
				fmt.Println("check\t\tChecks the status of the HTTP-server")
				fmt.Println("disc\t\tDisconnects the HTTP-server. (Shutdown)")
				fmt.Println("exit\t\tExits the Netdriver-Client.")
				fmt.Println("lista\t\tLists all AVAIABLE NBD-devices.")
				fmt.Println("listm\t\tLists all MOUNTED NBD-devices.")
				fmt.Println("mount\t\tMounts a NBD-device to specific image.")
				fmt.Println("unmount\t\tUnmounts the NBD-device specified.")
				fmt.Println("------------------------------------------------")
				fmt.Println("-------------------SHORTCUTS--------------------")
				fmt.Println("------------------------------------------------")
				fmt.Println("mount <device> <image> <username> <password>")
				fmt.Println("unmount <device>")
				fmt.Println("exit y")
				fmt.Println("------------------------------------------------")
				break

			case "exit":
				for{
				fmt.Println("Are you sure you want to exit? Y/N")
				_, _ = fmt.Scanln(&menu)
					if menu == "Y" || menu == "y"{
						return
					}
					if menu == "N" || menu == "n"{
						break
					}
				}
			}
	}
}