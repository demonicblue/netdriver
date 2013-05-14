package main

import (
	"fmt"
)

func main(){
	fmt.Println("Netdriver-Client started!")
	var cmd string

	for{
		_, err := fmt.Scanln(&cmd)
		if err != nil{
			fmt.Println(err)
		}

		switch cmd {
			case "mount":
				fmt.Println("Mounting...")
				break

			case "quit":
				for{
				fmt.Println("Are you sure you want to quit? Y/N")
				_, _ = fmt.Scanln(&cmd)
				if cmd == "Y" || cmd == "y"{
					return
				}
				if cmd == "N" || cmd == "n"{
					break
				}
			}

			case "list":
				fmt.Println("List of mounted NBD-devices:")
				break
		}
	}
}