package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

var (
	// ErrCorruptedMessage is returned when the encrypted input is incorrect
	ErrCorruptedMessage = errors.New("encrypted message seems corrupted")
	encoding            = binary.LittleEndian
)

// SymetricEncrypt encrypts the plain byte with the give key. It produces random IV that
// gets written to the output
func SymetricEncrypt(key, dst, src []byte) ([]byte, error) {
	aesKey := pbkdf2.Key(key, nil, 3, 32, sha256.New)
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	// Create the nonces
	nonceSize := aead.NonceSize()
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	encrypted := aead.Seal(dst[:0], nonce, src, nil)
	// Format is nonce size / nonce / encrypted
	result := make([]byte, 4+len(nonce)+len(encrypted))
	encoding.PutUint32(result[0:4], uint32(nonceSize))
	copy(result[4:4+nonceSize], nonce)
	copy(result[4+nonceSize:], encrypted)
	return result, nil
}

// SymetricDecrypt decrypts the encrypted key with the key
func SymetricDecrypt(key, dst, src []byte) ([]byte, error) {
	if len(src) < 4 { // We can't even get nonce size
		return nil, ErrCorruptedMessage
	}
	nonceSize := int(encoding.Uint32(src[0:4]))
	if nonceSize+4 > len(src) { // We don't even have all the nonce
		return nil, ErrCorruptedMessage
	}
	nonce := src[4 : 4+nonceSize]
	aesKey := pbkdf2.Key(key, nil, 3, 32, sha256.New)
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCMWithNonceSize(block, nonceSize)
	if err != nil {
		return nil, err
	}
	return aead.Open(dst[:0], nonce, src[4+nonceSize:], nil)
}
