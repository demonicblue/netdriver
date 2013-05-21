package main

import (
	"fmt"
	"net/http"
	"net/url"
	"flag"
)

/*
 *	Simple string constants for sending POST
 */
const(
	 httpString = "http://"
	 slash = "/"
)

var server, serverAdress, menu, targetImg, targetNBD string

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

func sendRequest(cmd string) *http.Response{
	values := make(url.Values)
	values.Set("command", cmd)
	
	if cmd == "unmount" || cmd == "mount" {
		fmt.Println("Type in your target NBD-device")
		_, _ = fmt.Scan(&targetNBD)
		
		if cmd == "mount" {
		fmt.Println("Type in your target image for",targetNBD)
		_, err := fmt.Scan(&targetImg)
			if err != nil {
				fmt.Println("%g", err)
			}
		values.Set("target", targetImg)
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
 * Usage: ./netdclient -c ip-address:port
 * Sends PostForms to the HTTP-server to handle the
 * IVBS-Netdriver.
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
		        resp := sendRequest(menu)
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