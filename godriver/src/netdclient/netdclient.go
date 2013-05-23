package main

import (
	"fmt"
	"net/http"
	"net/url"
	"flag"
	"encoding/json"
	"strings"
	"httpserver"
)

/*
 *	Simple string constants for sending POST
 */
const(
	 httpString = "http://"
	 slash = "/"
)

// Global-variables used within the client
var server, serverAdress, menu, targetImg, targetNBD, userName, passWord string

/*
 * Sends a POST-form to the HTTP-server
 * checking the connection-status.
 *
 * returns boolean for status
 */
func checkConnection() bool{
	values := make(url.Values)
	values.Set("command", "check")

	_, err := http.PostForm(serverAdress, values)

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

	temp := make([]byte, length)
	
	resp.Body.Read(temp)
	fmt.Print(string(temp))
	
	fmt.Println("------------------------------------------------")
}

func sendJSONRequest(cmd string) *http.Response{
	Cmd := httpserver.MountStruct{}
	Cmd.Command = cmd

	if cmd == "mount" || cmd == "unmount" {
		fmt.Println("Type in your target NBD-device")
		_, _ = fmt.Scan(&Cmd.Device)

		if cmd == "mount" {
			fmt.Println("Type in your target image for",targetNBD)
			_, _ = fmt.Scan(&Cmd.Image)
			fmt.Println("Type in your username")
			_, _ = fmt.Scan(&Cmd.User)
			fmt.Println("Type in your password")
			_, _ = fmt.Scan(&Cmd.Pass)
		}
	}

	b, _ := json.Marshal(Cmd)
	x := strings.NewReader(string(b)) 
	resp, _ := http.Post(serverAdress+"nmount", "application/json", x)
	return resp


}

/*
 * Takes a command and wraps it to POST-form
 * then sends it to the HTTP-server.
 * 
 * returns Response from HTTP-server
 */
func sendRequest(cmd string) *http.Response{
	values := make(url.Values)
	values.Set("command", cmd)
	
	if cmd == "unmount" || cmd == "mount" {
		fmt.Println("Type in your target NBD-device")
		_, _ = fmt.Scan(&targetNBD)
		
		if cmd == "mount" {
		fmt.Println("Type in your target image for",targetNBD)
		_, _ = fmt.Scan(&targetImg)
		fmt.Println("Type in your username")
		_, _ = fmt.Scan(&userName)
		fmt.Println("Type in your password")
		_, _ = fmt.Scan(&passWord)
			
		values.Set("target", targetImg)
		values.Set("user", userName)
		values.Set("pass", passWord)
		}
		values.Set("nbd", targetNBD)
	}

	resp, err := http.PostForm(serverAdress, values)

    if err != nil {
            fmt.Println("Error: %g",&err)
    }
    return resp
}

/*
 * Netdriver-Client main function.
 * The mainmenu for netdclient, takes a command
 * and sends the request to the HTTP-server for IVBS-Netdriver.
 */
func main(){
	fmt.Println("Netdriver-Client started!")

	flag.StringVar(&server, "c", "localhost:8080", "IP-address to HTTP-server")
	flag.Parse()
	serverAdress = httpString+server+slash //adds http:// and / to the serverstring

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
		       //resp := sendRequest(menu)
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
				fmt.Println("check \t Checks the status of the HTTP-server")
				fmt.Println("disc \t Disconnects the HTTP-server. (Shutdown)")
				fmt.Println("exit \t Exits the Netdriver-Client.")
				fmt.Println("lista \t Lists all AVAIABLE NBD-devices.")
				fmt.Println("listm \t Lists all MOUNTED NBD-devices.")
				fmt.Println("mount \t Mounts a NBD-device to specific image.")
				fmt.Println("unmount\t Unmounts the NBD-device specified.")
				fmt.Println("------------------------------------------------")
				fmt.Println("-------------------SHORTCUTS--------------------")
				fmt.Println("------------------------------------------------")
				fmt.Println("mount <device> <image>\t mounts device to image directly")
				fmt.Println("unmount <device>\t unmounts device directly")
				fmt.Println("exit y\t\t\t will exit without prompt")
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