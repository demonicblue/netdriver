package main

import (
	"fmt"
	"net/http"
	"net/url"
	"net"
	"flag"
)

/*
 *	Simple string constants for sending POST
 */
const(
	 httpString = "http://"
	 slash = "/"
)

var server string

func checkConnection() bool{
	if _, err := net.Dial("tcp", server); err != nil{
		fmt.Println("HTTP-server is offline.")
		return false
	}
	return true
}

/*
 * Netdriver-Client main function.
 * Usage: ./netdclient -c ip-address:port
 * Sends PostForms to the HTTP-server to handle the
 * IVBS-Netdriver.
 */
func main(){
	fmt.Println("Netdriver-Client started!")
	var cmd, targetImg, targetNBD string

	flag.StringVar(&server, "c", "localhost:12345", "IP-address to HTTP-server")
	flag.Parse()
	serverAdress := httpString+server+slash //adds http:// and / to the serverstring

	for{
		_, err := fmt.Scanln(&cmd)
		
		if err != nil{
			fmt.Println(err)
		}

		switch cmd {
			default:
				fmt.Println("Unknown command! Try type 'help' for available commands.")

			case "mount":
				if checkConnection() != true {
					break
				}
				fmt.Print("Type in your target NBD-device to mount:")
				_, _ = fmt.Scanln(&targetNBD)
				fmt.Print("Type in your target image for ",targetNBD,":")
				_, _ = fmt.Scanln(&targetImg)
				values := make(url.Values)
				values.Set("command", "mount")
				values.Set("target", targetImg)
				values.Set("nbd", targetNBD)
				resp, err := http.PostForm(serverAdress, values)

		        if err != nil {
		                fmt.Println("Error: %g",&err)
		        }
		        defer resp.Body.Close()
				
				break

			case "unmount":
				fmt.Print("Type in your target NBD-device to unmount:")
				_, _ = fmt.Scanln(&targetNBD)
				values := make(url.Values)
				values.Set("command", "unmount")
				values.Set("nbd", targetNBD)
				resp, err := http.PostForm(serverAdress, values)
		        
		        if err != nil {
		                fmt.Println("Error: %g",err)
		        }
		        defer resp.Body.Close()
				
				break

			case "listm":
				if checkConnection() != true {
					break
				}
				fmt.Println("List of all mounted NBD-devices:")
				values := make(url.Values)
				values.Set("command", "listm")
				resp, _ := http.PostForm(serverAdress, values)
				temp := make([]byte, 1024)
				for {
					if _, err := resp.Body.Read(temp); err == nil {
						fmt.Println(string(temp))
						break
					}
					if _, err := resp.Body.Read(temp); err != nil {
						break
					}
				}
				break

			case "lista":
				if checkConnection() != true {
					break
				}
				fmt.Println("List of all available NBD-devices:")
				values := make(url.Values)
				values.Set("command", "lista")
				resp, _ := http.PostForm(serverAdress, values)
				temp := make([]byte, 1024)
				for {
					if _, err := resp.Body.Read(temp); err == nil {
						fmt.Println(string(temp))
						break
					}
				}
				break

			case "disc":
				if checkConnection() != true {
					break
				}
				values := make(url.Values)
				values.Set("command", "exit")
				_, _ = http.PostForm(serverAdress, values)
				break

			case "check":
				if checkConnection(){
					fmt.Println("HTTP-server is online.")
				}
				break

			case "help":
				fmt.Println("\nCommands available:")
				fmt.Println("check \t Checks the status of the HTTP-server")
				fmt.Println("disc \t Disconnects the HTTP-server. (Shuts it down)")
				fmt.Println("exit \t Exits the Netdriver-Client.")
				fmt.Println("lista \t Lists all AVAIABLE NBD-devices.")
				fmt.Println("listm \t Lists all MOUNTED NBD-devices.")
				fmt.Println("mount \t Mounts a NBD-device to specific image.")
				fmt.Println("unmount \t Unmounts the NBD-device specified.")
				fmt.Println("------------------------------------------------")
				break

			case "exit":
				for{
				fmt.Println("Are you sure you want to exit? Y/N")
				_, _ = fmt.Scanln(&cmd)
					if cmd == "Y" || cmd == "y"{
						return
					}
					if cmd == "N" || cmd == "n"{
						break
					}
				}
			}
	}
}