package main

import (
	"fmt"
	"net/http"
	"net/url"
	"flag"
)

const(
	 httpString = "http://"
	 mountString = "/mount"
)

type Data struct {
	Int int
}

func main(){
	fmt.Println("Netdriver-Client started!")
	var cmd, server, target string

	flag.StringVar(&server, "c", "", "IP-address for HTTP-server")
	flag.Parse()

	for{
		_, err := fmt.Scanln(&cmd)
		
		if err != nil{
			fmt.Println(err)
		}

		switch cmd {
			default:
				fmt.Println("Unknown command! Try type 'help' for available commands.")

			case "mount":
				fmt.Print("Type in your target:")
				_, _ = fmt.Scanln(&target)
				values := make(url.Values)
				values.Set("data", target)
				resp, err := http.PostForm((httpString+server+mountString), values)

			        if err != nil {
			                fmt.Println(err)
			        }
			        defer resp.Body.Close()
			        fmt.Println(httpString+server+mountString)
				break

			case "list":
				fmt.Println("List of mounted NBD-devices:")
				break

			case "disc":
				break

			case "help":
				fmt.Println("\nCommands available:")
				fmt.Println("disc \t Disconnects you from the server.")
				fmt.Println("exit \t Exits the Netdriver-Client.")
				fmt.Println("list \t Lists all available NBD-devices.")
				fmt.Println("mount \t Mounts a NBD-device to specific image.")
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