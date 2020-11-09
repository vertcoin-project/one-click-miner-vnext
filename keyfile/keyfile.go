package keyfile

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/networks"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/scrypt"
)

// KeyFileValid returns true if there is a valid initialized keyfile
// available
func KeyFileValid() bool {
	return len(loadPublicKey()) == 33
}

func CreateKeyFile(pass string) error {
	filename := keyFile()

	// Create random key
	priv32 := new([32]byte)
	rand.Read(priv32[:])

	// Derive pubkey
	_, pub := btcec.PrivKeyFromBytes(btcec.S256(), priv32[:])

	salt := new([24]byte) // salt for scrypt / nonce for secretbox
	dk32 := new([32]byte) // derived key from scrypt

	//get 24 random bytes for scrypt salt (and secretbox nonce)
	_, err := rand.Read(salt[:])
	if err != nil {
		return err
	}
	// next use the pass and salt to make a 32-byte derived key
	dk, err := scrypt.Key([]byte(pass), salt[:], 16384, 8, 1, 32)
	if err != nil {
		return err
	}
	copy(dk32[:], dk[:])

	enckey := append(salt[:], secretbox.Seal(nil, priv32[:], salt, dk32)...)
	return ioutil.WriteFile(filename, append(pub.SerializeCompressed(), enckey...), 0600)
}

func keyFile() string {
	return filepath.Join(util.DataDirectory(), "keyfile.hex")
}

func loadPublicKey() []byte {
	filename := keyFile()
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		logging.Infof("Error reading keyfile: %s", err.Error())
		return []byte{}
	}
	if len(b) != 105 {
		logging.Infof("Keyfile had wrong length. Expected 129, got %d", len(b))
		return []byte{}
	}
	ret := make([]byte, 33)
	copy(ret, b[:33])
	b = nil
	return ret
}

func GetAddress() string {
	pub := loadPublicKey()
	return base58.CheckEncode(btcutil.Hash160(pub), networks.Active.Base58P2PKHVersion)
}

func LoadPrivateKey(password string) ([]byte, error) {
	filename := keyFile()
	keyfile, err := ioutil.ReadFile(filename)
	if err != nil {
		return []byte{}, err
	}
	if len(keyfile) != 105 {
		return []byte{}, fmt.Errorf("Key length error for %s ", filename)
	}

	enckey := keyfile[33:]
	// enckey is actually encrypted, get derived key from pass and salt
	// first extract salt
	salt := new([24]byte)      // salt (also nonce for secretbox)
	dk32 := new([32]byte)      // derived key array
	copy(salt[:], enckey[:24]) // first 24 bytes are scrypt salt/box nonce

	dk, err := scrypt.Key([]byte(password), salt[:], 16384, 8, 1, 32) // derive key
	if err != nil {
		return []byte{}, err
	}
	copy(dk32[:], dk[:]) // copy into fixed size array

	// nonce for secretbox is the same as scrypt salt.  Seems fine.  Really.
	priv, worked := secretbox.Open(nil, enckey[24:], salt, dk32)
	if worked != true {
		return []byte{}, fmt.Errorf("Decryption failed for %s ", filename)
	}

	return priv, nil
}

func TestPassword(password string) bool {
	priv, err := LoadPrivateKey(password)
	if err != nil {
		return false
	}
	_, pub := btcec.PrivKeyFromBytes(btcec.S256(), priv)
	priv = nil
	return bytes.Equal(loadPublicKey(), pub.SerializeCompressed())
}
