package crypto_test

import (
	"bytes"
	"compress/gzip"
	"crypto/rsa"
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/pkg/crypto"
)

func TestDecryption_Decrypt(t *testing.T) {
	t.Parallel()

	publicFilePath := filepath.Join("testdata", "test_public.key")
	privateFilePath := filepath.Join("testdata", "test_private.key")

	encryption, err := crypto.NewEncryption(publicFilePath)
	require.NoError(t, err)

	decryption, err := crypto.NewDecryption(privateFilePath)
	require.NoError(t, err)

	payload := []byte(`here is my message`)

	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	_, err = writer.Write(payload)
	require.NoError(t, err)

	if err := writer.Close(); err != nil {
		t.Errorf("failed to close writer: %s", err)
	}

	encrypted, err := encryption.Encrypt(buf.Bytes())
	require.NoError(t, err)

	reader, err := gzip.NewReader(&buf)
	require.NoError(t, err)

	decompressed, err := io.ReadAll(reader)
	require.NoError(t, err)

	if !bytes.Equal(payload, decompressed) {
		t.Errorf("failed to compress: %s", err)
	}

	type fields struct {
		PrivateKey *rsa.PrivateKey
		Nonce      [12]byte
	}
	type args struct {
		encrypted []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:   "Test #1 Success",
			fields: fields(*decryption),
			args: args{
				encrypted: encrypted,
			},
			want:    decompressed,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := &crypto.Decryption{
				PrivateKey: tt.fields.PrivateKey,
				Nonce:      tt.fields.Nonce,
			}
			decrypted, err := d.Decrypt(tt.args.encrypted)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decryption.Decrypt() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			reader, err := gzip.NewReader(bytes.NewBuffer(decrypted))
			if err != nil {
				t.Errorf("failed to get reader: %s", err)
			}
			got, err := io.ReadAll(reader)
			if err != nil {
				t.Errorf("failed to decompress: %s", err)
			}

			if !bytes.Equal(got, tt.want) {
				t.Errorf("Decryption.Decrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}
