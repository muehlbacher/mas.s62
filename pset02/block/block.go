package block

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
)

// A hash is a sha256 hash, as in pset01
type Hash [32]byte

// ToString gives you a hex string of the hash
func (hash Hash) ToString() string {
	return fmt.Sprintf("%x", hash)
}

// Blocks are what make the chain in this pset; different than just a 32 byte array
// from last time.  Has a previous block hash, a name and a nonce.
type Block struct {
	PrevHash Hash
	Name     string
	Nonce    string
}

// ToString turns a block into an ascii string which can be sent over the
// network or printed to the screen.
func (block Block) ToString() string {
	return fmt.Sprintf("%x %s %s", block.PrevHash, block.Name, block.Nonce)
}

// Hash returns the sha256 hash of the block.  Hopefully starts with zeros!
func (block Block) Hash() Hash {
	return sha256.Sum256([]byte(block.ToString()))
}

// BlockFromString takes in a string and converts it to a block, if possible
func BlockFromString(s string) (Block, error) {
	var bl Block

	// check string length
	if len(s) < 66 || len(s) > 100 {
		return bl, fmt.Errorf("invalid string length %d, expect 66 to 100", len(s))
	}
	// split into 3 substrings via spaces
	subStrings := strings.Split(s, " ")

	if len(subStrings) != 3 {
		return bl, fmt.Errorf("got %d elements, expect 3", len(subStrings))
	}

	hashbytes, err := hex.DecodeString(subStrings[0])
	if err != nil {
		return bl, err
	}
	if len(hashbytes) != 32 {
		return bl, fmt.Errorf("got %d byte hash, expect 32", len(hashbytes))
	}

	copy(bl.PrevHash[:], hashbytes)

	bl.Name = subStrings[1]

	// remove trailing newline if there; the blocks don't include newlines, but
	// when transmitted over TCP there's a newline to signal end of block
	bl.Nonce = strings.TrimSpace(subStrings[2])

	// TODO add more checks on name/nonce ...?

	return bl, nil
}

func FindBlock(targetBits uint8, firstNonce int, fitted_block chan<- Block) {
	var block Block
	var hash Hash
	target := strings.Repeat("0", int(targetBits))
	nonce := 0
	first := fmt.Sprintf("%d", firstNonce)
	for {
		block.Nonce = first + fmt.Sprintf("%d", nonce)

		hash = block.Hash()

		if strings.HasPrefix(hash.ToString(), target) {
			fitted_block <- block
			break
		}
		nonce++
		if nonce%1000000 == 0 {
			fmt.Println("Tries: ", block.Nonce, first)
		}
	}
}

func (self *Block) Mine(targetBits uint8) {

	var wg sync.WaitGroup
	fitted_block := make(chan Block)

	for i := 1; i <= 16; i++ {
		wg.Add(1)
		go func(firstNonce int) {
			defer wg.Done()
			FindBlock(targetBits, firstNonce, fitted_block)
		}(i)
	}
	select {
	case result := <-fitted_block:
		fmt.Println("Found block:", result)
		return
	case <-func() chan struct{} {
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()
		return done
	}():
	}

	// Wait for the block to be mined
	go func() {
		wg.Wait()
		close(fitted_block) // Close the channel when done
	}()

	// Receive and handle the mined block
	for block := range fitted_block {
		fmt.Println("Found block:", block)
	}
}
