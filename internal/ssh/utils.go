// Copyright 2025 Canonical.

package ssh

import (
	gossh "golang.org/x/crypto/ssh"
)

func GetFingerprintFromPrivateKey(privateKey []byte) (string, error) {
	key, err := gossh.ParsePrivateKey(privateKey)
	if err != nil {
		return "", err
	}
	return gossh.FingerprintSHA256(key.PublicKey()), nil
}
