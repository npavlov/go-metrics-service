package crypto

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"os"

	"github.com/pkg/errors"
)

const (
	RandomSize              int = 2 * aes.BlockSize
	EncryptedSessionKeySize int = 512
)

type Key []byte

func NewKey() (Key, error) {
	key, err := GenerateRandom(RandomSize)
	if err != nil {
		return key, errors.Wrap(err, "failed to generate random")
	}

	return key, nil
}

func (k Key) Encrypt(public *rsa.PublicKey) ([]byte, error) {
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, public, k)
	if err != nil {
		return encrypted, errors.Wrap(err, "failed to encrypt")
	}

	return encrypted, nil
}

func (k Key) Decrypt(private *rsa.PrivateKey, encrypted []byte) error {
	if err := rsa.DecryptPKCS1v15SessionKey(nil, private, encrypted, k); err != nil {
		return errors.Wrap(err, "failed to decrypt")
	}

	return nil
}

func GenerateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return b, errors.Wrap(err, "failed to generate random")
	}

	return b, nil
}

const (
	PrivateKeyTitle = "RSA PRIVATE KEY"
	PublicKeyTitle  = "RSA PUBLIC KEY"
)

func readFile(fname string) ([]byte, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}
	defer func() {
		err = file.Close()
	}()
	bytes, err := io.ReadAll(file)

	return bytes, nil
}

func GetPrivateKey(fname string) (*rsa.PrivateKey, error) {
	op := "GetPrivateKey"
	b, err := readFile(fname)
	if err != nil {
		return nil, errors.Wrapf(err, "%s failed with an error", op)
	}
	block, _ := pem.Decode(b)
	if block == nil || block.Type != PrivateKeyTitle {
		return nil, errors.New(op + " failed to decode PEM block containing private key")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse private key")
	}

	return key, nil
}

func GetPublicKey(publicKeyPath string) (*rsa.PublicKey, error) {
	publicKeyBytes, err := readFile(publicKeyPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed with an error")
	}
	block, _ := pem.Decode(publicKeyBytes)
	if block == nil || block.Type != PublicKeyTitle {
		return nil, errors.New("failed to decode PEM block containing public key")
	}
	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse public key")
	}

	return key, nil
}
