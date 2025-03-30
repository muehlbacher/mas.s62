package block

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"runtime"
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

func FindBlockWorker(id int, targetBits uint8, nonce <-chan int32, fitted_block chan<- Block, name string, phash Hash) {
	b1 := Block{
		PrevHash: phash,
		Name:     name,
	}
	var hash Hash
	target := strings.Repeat("0", int(targetBits))
	for n := range nonce {
		for i := n - 10000000; i < n; i++ {
			b1.Nonce = fmt.Sprintf("%d", i)

			hash = b1.Hash()

			if strings.HasPrefix(hash.ToString(), target) {
				fitted_block <- b1
				break
			}
			if i%1000000 == 0 {
				fmt.Println("Tries: ", b1.Nonce, "ID: ", id)
			}
		}
	}
}

func (self *Block) Mine(targetBits uint8) Block {
	const numbWorkers = 10
	const buffersize = 5
	nonce := make(chan int32, buffersize)
	block := make(chan Block)

	runtime.GOMAXPROCS(numbWorkers) // Allow Go to use multiple CPU cores

	for w := 1; w <= numbWorkers; w++ {
		go FindBlockWorker(w, targetBits, nonce, block, self.Name, self.PrevHash)
	}

	go func() {
		var n int32 = 1
		for {
			nonce <- n
			n = n + 10000000
		}
	}()

	select {
	case b := <-block:
		close(block)
		fmt.Println(b.ToString())
		return b
	}

}

func (self *Block) Mine_old(targetBits uint8) {

	var wg sync.WaitGroup
	fitted_block := make(chan Block)

	for i := 1; i <= 16; i++ {
		wg.Add(1)
		go func(firstNonce int) {
			defer wg.Done()
			//FindBlock(targetBits, firstNonce, fitted_block)
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
