package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/exc-git/mini-kv-store/internal/transport/rpc"
)

func main() {
	serverAddr := flag.String("server", "localhost:9090", "Server RPC address")
	flag.Parse()

	client, err := rpc.NewClient(*serverAddr)
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer client.Close()

	fmt.Println("Mini KV Store Client")
	fmt.Println("Commands:")
	fmt.Println("  get <key>")
	fmt.Println("  set <key> <value>")
	fmt.Println("  set-ttl <key> <value> <seconds>")
	fmt.Println("  delete <key>")
	fmt.Println("  stats")
	fmt.Println("  exit")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		args := strings.Fields(line)
		cmd := args[0]
		args = args[1:]

		ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Second)
		defer cancel()

		switch cmd {
		case "get":
			if len(args) != 1 {
				fmt.Println("Usage: get <key>")
				continue
			}
			val, err := client.Get(ctx, args[0])
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println(val)
			}

		case "set":
			if len(args) != 2 {
				fmt.Println("Usage: set <key> <value>")
				continue
			}
			err := client.Set(ctx, args[0], args[1])
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		case "set-ttl":
			if len(args) != 3 {
				fmt.Println("Usage: set-ttl <key> <value> <seconds>")
				continue
			}
			ttl, err := time.ParseDuration(args[2] + "s")
			if err != nil {
				fmt.Printf("Invalid TTL: %v\n", err)
				continue
			}
			err = client.SetWithTTL(ctx, args[0], args[1], ttl)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		case "delete":
			if len(args) != 1 {
				fmt.Println("Usage: delete <key>")
				continue
			}
			err := client.Delete(ctx, args[0])
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		case "stats":
			stats, err := client.Stats(ctx)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Cluster Stats:")
				fmt.Printf("Leader: %s\n", stats.Leader)
				fmt.Println("Nodes:")
				for _, node := range stats.Nodes {
					fmt.Printf("- %s\n", node)
				}
				fmt.Printf("Commit Index: %d\n", stats.CommitIndex)
				fmt.Printf("Last Applied: %d\n", stats.LastApplied)
			}

		case "exit":
			return

		default:
			fmt.Println("Unknown command:", cmd)
		}
	}
}
