package handler

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
)

// MD5String - return a file md5 hash string
func MD5String(src []byte) (string, error) {
	hash := md5.New()

	if _, err := io.Copy(hash, bytes.NewBuffer(src)); err != nil {
		return "", err
	}

	hashInBytes := hash.Sum(nil)[:16]
	return hex.EncodeToString(hashInBytes), nil
}
