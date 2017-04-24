package crypt

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var raw []byte
var (
	ErrNoPayloadEnv = errors.New("PAYLOAD env was not set")
)

func init() {
	var err error
	payload := os.Getenv("PAYLOAD")
	if len(payload) > 0 {
		raw, err = ioutil.ReadFile(payload)
		if err != nil {
			fmt.Printf("Error opening payload: %s\n", err)
		}
	}
}

func TestBasicEncryptDecrypt(t *testing.T) {
	assert := assert.New(t)
	data := []byte("Hello World! 🍣")
	key := []byte("super secret 🍣")
	encrypted, err := SymetricEncrypt(key, nil, data)
	assert.NoError(err)
	assert.NotEmpty(encrypted)

	decrypted, err := SymetricDecrypt(key, nil, encrypted)
	assert.NoError(err)
	assert.NotEmpty(decrypted)
	assert.Equal(data, decrypted)
}

func TestDstEncryptDecrypt(t *testing.T) {
	assert := assert.New(t)
	data := []byte("Hello World! 🍣")
	key := []byte("super secret 🍣")
	for _, dstSize := range []int{1, 10, 100, 1000} {
		dst := make([]byte, dstSize)
		encrypted, err := SymetricEncrypt(key, dst, data)
		assert.NoError(err)
		assert.NotEmpty(encrypted)

		decrypted, err := SymetricDecrypt(key, dst, encrypted)
		assert.NoError(err)
		assert.NotEmpty(decrypted)
		assert.Equal(data, decrypted)
	}
}

func BenchmarkSymetricEncrypt(b *testing.B) {
	if raw == nil {
		b.Fatal(ErrNoPayloadEnv)
	}
	b.SetBytes(int64(len(raw)))
	key := []byte("super secret 🍣")
	for i := 0; i < b.N; i++ {
		SymetricEncrypt(key, nil, raw)
	}
}

func BenchmarkSymetricDecrypt(b *testing.B) {
	if raw == nil {
		b.Fatal(ErrNoPayloadEnv)
	}
	b.SetBytes(int64(len(raw)))
	key := []byte("super secret 🍣")
	encrypted, _ := SymetricEncrypt(key, nil, raw)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SymetricDecrypt(key, nil, encrypted)
	}
}
