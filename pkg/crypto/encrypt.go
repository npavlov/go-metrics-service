package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"

	"github.com/pkg/errors"
)

type Encryption struct {
	PublicKey *rsa.PublicKey
	Nonce     [12]byte
}

func NewEncryption(publicKeyPath string) (*Encryption, error) {
	if publicKeyPath == "" {
		return nil, errors.New("no public key provided")
	}
	PublicKey, err := GetPublicKey(publicKeyPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get public key")
	}

	return &Encryption{PublicKey: PublicKey, Nonce: [12]byte{}}, nil
}

func (e *Encryption) Encrypt(payload []byte) ([]byte, error) {
	key, err := NewKey()
	if err != nil {
		return payload, errors.Wrap(err, "failed to get new key")
	}
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return payload, errors.Wrap(err, "failed to get aes block")
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return payload, errors.Wrap(err, "failed to get aesgcm")
	}
	eKey, err := key.Encrypt(e.PublicKey)
	if err != nil {
		return payload, errors.Wrap(err, "failed to encrypt")
	}
	encrypted := make([]byte, 0, len(payload)+EncryptedSessionKeySize+len(key))
	encrypted = append(encrypted, eKey...)
	encrypted = append(encrypted, aesgcm.Seal(nil, e.Nonce[:], payload, nil)...)

	return encrypted, nil
}
