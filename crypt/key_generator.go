package crypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// DefaultBitsSize is the default bits size of the key
	DefaultBitsSize = 8192
	// KeyVersion is the versioning of the Marshaling/Unmarshaling of the key
	KeyVersion = 1
)

var (
	// ErrUnsuportedVersion is returned when there is a key version we can't handle
	ErrUnsuportedVersion = errors.New("unsupported version")
)

// PublicKey defines a tri PublicKey. name is an identifier of the key which can
// be arbitrary, it is used as a label to the RSA OAEP encryption
// (so when decrypting you now it was encrypted with the same key); it is optional
// asyncPubKey is a RSA public key
type PublicKey struct {
	AsymPubKey []byte
	KeyLength  int
	Name       []byte
}

// PrivateKey is the private counter-part to PublicKey. It is a RSA private key
type PrivateKey struct {
	AsymPrivKey []byte
	PubKey      PublicKey
}

// GetWeakKey returns the "weak" (tri-terminology) symetric key used to encode/decode
func (pub *PublicKey) GetWeakKey() []byte {
	return pbkdf2.Key(pub.AsymPubKey, pub.Name, 3, pub.KeyLength, sha256.New)
}

type jsonPublicKey struct {
	KeyLength int `json:"key_length"`
	Version   int `json:"version"`
	// Everything after is encrypted with password
	AsymPubKey       []byte `json:"pub_key"`
	EncryptedVersion []byte `json:"encrypted_version"`
	Name             []byte `json:"name"`
}

// Marshal returns a bytes blob representing a public key
func (pub *PublicKey) Marshal(key string) ([]byte, error) {
	asymK, err := SymetricEncrypt([]byte(key), nil, pub.AsymPubKey)
	if err != nil {
		return nil, err
	}
	encryptedVersion, err := SymetricEncrypt([]byte(key), nil, []byte(fmt.Sprintf("%v", KeyVersion)))
	if err != nil {
		return nil, err
	}
	name, err := SymetricEncrypt([]byte(key), nil, pub.Name)
	if err != nil {
		return nil, err
	}
	j := jsonPublicKey{
		KeyLength:        pub.KeyLength,
		Version:          KeyVersion,
		AsymPubKey:       asymK,
		EncryptedVersion: encryptedVersion,
		Name:             name,
	}
	return json.Marshal(j)
}

// ParsePublicKey parses a json decription of your public key
func ParsePublicKey(key string, src []byte) (PublicKey, error) {
	var j jsonPublicKey
	err := json.Unmarshal(src, &j)
	if err != nil {
		return PublicKey{}, err
	}

	// Check that the message is authentic first
	if j.Version != 1 {
		return PublicKey{}, ErrUnsuportedVersion
	}
	decryptedVersion, err := SymetricDecrypt([]byte(key), nil, j.EncryptedVersion)
	if err != nil {
		return PublicKey{}, err
	}
	dVersion, err := strconv.Atoi(string(decryptedVersion))
	if err != nil {
		return PublicKey{}, err
	}
	if j.Version != dVersion {
		return PublicKey{}, ErrCorruptedMessage
	}
	// Decrypt the rest
	pubKey, err := SymetricDecrypt([]byte(key), nil, j.AsymPubKey)
	if err != nil {
		return PublicKey{}, err
	}
	name, err := SymetricDecrypt([]byte(key), nil, j.Name)
	if err != nil {
		return PublicKey{}, err
	}
	return PublicKey{
		AsymPubKey: pubKey,
		KeyLength:  j.KeyLength,
		Name:       name,
	}, nil
}

// GenerateNewKey generates a new set of PrivateKey and PublicKey
// name is an identifier of the key which can
// be arbitrary, it is used as a label to the RSA OAEP encryption
// (so when decrypting you now it was encrypted with the same key); it is optional
func GenerateNewKey(name string, bits int) (PrivateKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return PrivateKey{}, err
	}
	pubBytes, err := x509.MarshalPKIXPublicKey(priv.Public())
	if err != nil {
		return PrivateKey{}, err
	}
	privBytes := x509.MarshalPKCS1PrivateKey(priv)

	pubKey := PublicKey{
		AsymPubKey: pubBytes,
		KeyLength:  bits,
		Name:       []byte(name),
	}
	privKey := PrivateKey{
		AsymPrivKey: privBytes,
		PubKey:      pubKey,
	}
	return privKey, nil
}
