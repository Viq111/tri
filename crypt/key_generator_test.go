package crypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerationKey(t *testing.T) {
	if testing.Short() { // Generation is slow, skip is we want fast
		t.Skip()
	}
	assert := assert.New(t)

	keyLength := 256 // Use a smaller key for tests
	testName := "testing"
	key, err := GenerateNewKey(testName, keyLength)
	assert.NoError(err)
	assert.NotEmpty(key.AsymPrivKey)
	assert.NotEmpty(key.PubKey.AsymPubKey)
	assert.Equal(testName, string(key.PubKey.Name))
}

func TestGenerationWeakKey(t *testing.T) {
	if testing.Short() { // Generation is slow, skip is we want fast
		t.Skip()
	}
	assert := assert.New(t)
	keyLength := 256 // Use a smaller key for tests
	testName := "testing"
	key, err := GenerateNewKey(testName, keyLength)
	assert.NoError(err)
	weak := key.PubKey.GetWeakKey()
	assert.Equal(keyLength, len(weak))
	assert.Equal(key.PubKey.GetWeakKey(), weak) // Regeneration is consistent
}

func TestPublicMarshaling(t *testing.T) {
	if testing.Short() { // Generation is slow, skip is we want fast
		t.Skip()
	}
	assert := assert.New(t)
	keyLength := 256 // Use a smaller key for tests
	testName := "testing"
	password := "super secret"
	key, err := GenerateNewKey(testName, keyLength)
	public := key.PubKey
	marshalled, err := public.Marshal(password)
	assert.NoError(err)
	assert.NotEmpty(marshalled)
	t.Log(string(marshalled))
	unmarshalled, err := ParsePublicKey(password, marshalled)
	assert.NoError(err)
	assert.Equal(public.AsymPubKey, unmarshalled.AsymPubKey)
	assert.Equal(public.KeyLength, unmarshalled.KeyLength)
	assert.Equal(public.Name, unmarshalled.Name)
}
