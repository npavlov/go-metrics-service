package crypto_test

import (
	"crypto/rsa"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/npavlov/go-metrics-service/pkg/crypto"
)

func TestEncryption_Encrypt(t *testing.T) {
	t.Parallel()

	publicFilePath := filepath.Join("testdata", "test_public.key")
	privateFilePath := filepath.Join("testdata", "test_private.key")

	encryption, err := crypto.NewEncryption(publicFilePath)
	require.NoError(t, err)

	decryption, err := crypto.NewDecryption(privateFilePath)
	require.NoError(t, err)

	type fields struct {
		PublicKey *rsa.PublicKey
		Nonce     [12]byte
	}
	type args struct {
		payload []byte
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
			fields: fields(*encryption),
			args: args{
				payload: []byte("some test message"),
			},
			want:    []byte("some test message"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := &crypto.Encryption{
				PublicKey: tt.fields.PublicKey,
				Nonce:     tt.fields.Nonce,
			}
			encrypted, err := e.Encrypt(tt.args.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encryption.Encrypt() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			got, err := decryption.Decrypt(encrypted)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encryption.Encrypt() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Encryption.Encrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}
