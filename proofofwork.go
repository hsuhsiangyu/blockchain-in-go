package main

import (
    "fmt"
    "crypto/sha256"
    "math"
    "math/big"
    "bytes"
)

const targetBits = 8 

var (
    maxNonce = math.MaxInt64
)

//holds a pointer to a block and a pointer to a target
type ProofOfWork struct {
    block  *Block
    target *big.Int
}

// initialize a big.Int with the value of 1 and shift it left by 256 - targetBits bits.
func NewProofOfWork(b *Block) *ProofOfWork {
    target := big.NewInt(1)
    target.Lsh(target, uint(256-targetBits))

    pow := &ProofOfWork{b, target}

    return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
    data := bytes.Join(
        [][]byte{
            pow.block.PrevBlockHash,
            pow.block.Data,
            IntToHex(pow.block.Timestamp),
            IntToHex(int64(targetBits)),
            IntToHex(int64(nonce)),
        },
        []byte{},
    )                                                            
    return data
}

// Run performs a proof-of-work
func (pow *ProofOfWork) Run() (int, []byte) {
    var hashInt big.Int
    var hash [32]byte
    nonce := 0

    fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
    for nonce < maxNonce {
            data := pow.prepareData(nonce)
            hash = sha256.Sum256(data)
            fmt.Printf("\r%x", hash)
            hashInt.SetBytes(hash[:])  //Convert the hash to a big integer.
                    
            if hashInt.Cmp(pow.target) == -1 {
                break
            } else {
                nonce++
            }
        }
        fmt.Print("\n\n")
                                                        
    return nonce, hash[:]
}

// Validate validates block's PoW
func (pow *ProofOfWork) Validate() bool {
    var hashInt big.Int

    data := pow.prepareData(pow.block.Nonce)
    hash := sha256.Sum256(data)
    hashInt.SetBytes(hash[:])

    isValid := hashInt.Cmp(pow.target) == -1

    return isValid
}




