package main

import (
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha256"
    "bytes"
    "golang.org/x/crypto/ripemd160"
    "log"
)

const version = byte(0x00)
const addressChecksumLen = 4


type Wallet struct {
    PrivateKey ecdsa.PrivateKey
    PublicKey  []byte
}

func NewWallet() *Wallet {
    private, public := newKeyPair()
    wallet := Wallet{private, public}

    return &wallet
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
    curve := elliptic.P256() //ECDSA is based on elliptic curves, so we need one
    private, err := ecdsa.GenerateKey(curve, rand.Reader)
    if err != nil {
        log.Panic(err)
    }
    pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)  //public key is a combination of X, Y coordinates

    return *private, pubKey
}

// GetAddress returns wallet address, look at address-generation-scheme.png
func (w Wallet) GetAddress() []byte { 
    pubKeyHash := HashPubKey(w.PublicKey)

    versionedPayload := append([]byte{version}, pubKeyHash...)
    checksum := checksum(versionedPayload)

    fullPayload := append(versionedPayload, checksum...)
    address := Base58Encode(fullPayload)

    return address
}

// Take the public key and hash it twice with RIPEMD160(SHA256(PubKey)) hashing algorithms.
func HashPubKey(pubKey []byte) []byte {
    publicSHA256 := sha256.Sum256(pubKey)

    RIPEMD160Hasher := ripemd160.New()
    _, err := RIPEMD160Hasher.Write(publicSHA256[:])
    if err != nil {
        log.Panic(err)
    }
    publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

    return publicRIPEMD160
}

// ValidateAddress check if address if valid
func ValidateAddress(address string) bool {
    pubKeyHash := Base58Decode([]byte(address))
    actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
    version := pubKeyHash[0]
    pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
    targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

    return bytes.Compare(actualChecksum, targetChecksum) == 0
}

//The checksum is the first four bytes of the resulted hash.
func checksum(payload []byte) []byte {
    firstSHA := sha256.Sum256(payload)
    secondSHA := sha256.Sum256(firstSHA[:])

    return secondSHA[:addressChecksumLen]
}








