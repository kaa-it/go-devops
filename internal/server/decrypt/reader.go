package decrypt

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"
)

type AESReader struct {
	buffer *bytes.Buffer
}

func decrypt(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	decrypted := make([]byte, 0, len(data))

	step := privateKey.Size()
	total := len(data)

	for start := 0; start < total; start += step {
		finish := start + step
		if finish > total {
			finish = total
		}

		decoded, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, data[start:finish])
		if err != nil {
			return nil, err
		}

		decrypted = append(decrypted, decoded...)
	}

	return decrypted, nil
}

func NewAESReader(privateKey *rsa.PrivateKey, r *http.Request) (*AESReader, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	data, err := decrypt(privateKey, body)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(data)

	return &AESReader{
		buffer: buf,
	}, nil
}

func (ar *AESReader) Read(p []byte) (n int, err error) {
	return ar.buffer.Read(p)
}

func (ar *AESReader) Close() error {
	return nil
}
