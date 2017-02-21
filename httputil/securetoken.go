package httputil

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"io"
)

const (
	blockSize     = 16
	signatureSize = 32
)

var (
	ErrShortKey = errors.New("Signing/encryption keys must be 32 characters")
	ErrGenIV    = errors.New("Failed to generate initialization vector")
	ErrBadSig   = errors.New("Invalid signature")
	ErrBadData  = errors.New("Invalid input data")
)

type TokenHandler struct {
	signingKey []byte
	cryptKey   []byte

	blockCipher cipher.Block
}

func NewTokenHandler(signingKey, cryptKey []byte) (TokenHandler, error) {
	if len(signingKey) != 32 || len(cryptKey) != 32 {
		return TokenHandler{}, ErrShortKey
	}

	h := TokenHandler{
		signingKey: signingKey,
		cryptKey:   cryptKey,
	}

	b, err := aes.NewCipher(h.cryptKey)
	if err != nil {
		return h, err
	}

	h.blockCipher = b
	return h, nil
}

func (h TokenHandler) Encode(value interface{}) ([]byte, error) {
	// Convert value to byte slice with gob encoder.
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(value); err != nil {
		return nil, err
	}

	encoded := buf.Bytes()

	// Prepend initialization vector to encoded value.
	iv := randBytes(blockSize)
	if iv == nil {
		return nil, ErrGenIV
	}

	// Encrypt the byte slice with AES-CTR using cryptKey.
	// The result here is that `encoded` is encrypted in place.
	stream := cipher.NewCTR(h.blockCipher, iv)
	stream.XORKeyStream(encoded, encoded)
	encoded = append(iv, encoded...)

	// Create signature using hmac with signingKey.
	signature, err := createSignature(h.signingKey, encoded)
	if err != nil {
		return nil, err
	}

	// Combine encrypted value and signature.
	encoded = append(signature, encoded...)

	// Encode as base64.
	encoded64 := make([]byte, base64.URLEncoding.EncodedLen(len(encoded)))
	base64.URLEncoding.Encode(encoded64, encoded)
	return encoded64, nil
}

func (h TokenHandler) Decode(encoded64 []byte, value interface{}) error {
	// Decode from base64.
	encoded := make([]byte, base64.URLEncoding.DecodedLen(len(encoded64)))
	n, err := base64.URLEncoding.Decode(encoded, encoded64)
	if err != nil {
		return err
	}

	encoded = encoded[:n]

	// Split encrtyped value and signature.
	signature := encoded[:signatureSize]
	encoded = encoded[signatureSize:]

	// Check that signature is correct.
	if !checkSignatureMatches(encoded, signature, h.signingKey) {
		return ErrBadSig
	}

	if len(encoded) < blockSize {
		return ErrBadData
	}

	// Split initialization vector from encrypted value.
	iv := encoded[:blockSize]
	encoded = encoded[blockSize:]

	// Decrypt encrypted value.
	// The result should be that `encoded` is no longer encoded. :)
	stream := cipher.NewCTR(h.blockCipher, iv)
	stream.XORKeyStream(encoded, encoded)

	// Decode value into object with gob decoder.
	dec := gob.NewDecoder(bytes.NewBuffer(encoded))
	return dec.Decode(value)
}

func createSignature(signingKey, value []byte) ([]byte, error) {
	h := hmac.New(sha256.New, signingKey)
	_, err := h.Write(value)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func checkSignatureMatches(value, valueMAC, signingKey []byte) bool {
	expected, err := createSignature(signingKey, value)
	if err != nil {
		return false
	}
	return hmac.Equal(valueMAC, expected)
}

func randBytes(N int) []byte {
	k := make([]byte, N)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil
	}
	return k
}
