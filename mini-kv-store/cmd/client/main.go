package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/exc-git/mini-kv-store/internal/transport/rpc"
)

func main() {
	serverAddr := flag.String("server", "localhost:8080", "Server address")
	flag.Parse()

	client, err := rpc.NewClient(*serverAddr)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Mini KV Store Client")
	fmt.Println("Commands: get <key>, put <key> <value>, delete <key>, exit")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "get":
			if len(parts) != 2 {
				fmt.Println("Usage: get <key>")
				continue
			}
			value, found, err := client.Get(parts[1])
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else if found {
				fmt.Printf("%s\n", value)
			} else {
				fmt.Println("Key not found")
			}

		case "put":
			if len(parts) != 3 {
				fmt.Println("Usage: put <key> <value>")
				continue
			}
			if err := client.Put(parts[1], parts[2]); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		case "delete":
			if len(parts) != 2 {
				fmt.Println("Usage: delete <key>")
				continue
			}
			if err := client.Delete(parts[1]); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		case "exit":
			return

		default:
			fmt.Println("Unknown command. Available: get, put, delete, exit")
		}
	}
}
