package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"

	"github.com/pkg/errors"
)

type Decryption struct {
	PrivateKey *rsa.PrivateKey
	Nonce      [12]byte
}

func NewDecryption(privateKeyPath string) (*Decryption, error) {
	if privateKeyPath == "" {
		return nil, errors.New("no private key provided")
	}
	PrivateKey, err := GetPrivateKey(privateKeyPath)
	if err != nil {
		return nil, err
	}

	return &Decryption{PrivateKey: PrivateKey, Nonce: [12]byte{}}, nil
}

func (d *Decryption) Decrypt(encrypted []byte) ([]byte, error) {
	newKey, err := NewKey()
	if err != nil {
		return encrypted, err
	}
	if len(encrypted) < EncryptedSessionKeySize {
		return encrypted, errors.New("wrong key size")
	}
	err = newKey.Decrypt(d.PrivateKey, encrypted[:EncryptedSessionKeySize])
	if err != nil {
		return encrypted, errors.Wrap(err, "failed to decrypt")
	}
	aesblock, err := aes.NewCipher(newKey)
	if err != nil {
		return encrypted, errors.Wrap(err, "failed to get aesblock")
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return encrypted, errors.Wrap(err, "failed to get aesgcm")
	}

	decrypted, err := aesgcm.Open(nil, d.Nonce[:], encrypted[EncryptedSessionKeySize:], nil)
	if err != nil {
		return decrypted, errors.Wrap(err, "failed to decrypt")
	}

	return decrypted, nil
}
