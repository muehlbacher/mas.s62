package main

import (
	"fmt"
	"log"
	"pset02/block"
	"pset02/client"
)

func main() {

	fmt.Printf("NameChain Miner v0.1\n")

	first_block, err := client.GetTipFromServer()
	if err != nil {
		log.Fatalf("lol Server Error: %s \n", err)
	}
	block1 := block.Block{
		PrevHash: first_block.Hash(), Name: "Dominik", Nonce: "lol",
	}
	// Your code here!
	block1.Mine(4)
	client.SendBlockToServer(block1)

	// Basic idea:
	// Get tip from server, mine a block pointing to that tip,
	// then submit to server.
	// To reduce stales, poll the server every so often and update the
	// tip you're mining off of if it has changed.

	return
}
