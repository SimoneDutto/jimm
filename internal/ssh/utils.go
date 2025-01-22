// Copyright 2025 Canonical.

package ssh

import (
	gossh "golang.org/x/crypto/ssh"
)

// GetFingerprintsFromPrivateKey returns the fingerprints of the public key from the private key.
func GetFingerprintsFromPrivateKey(privateKey []byte) (map[string]string, error) {
	key, err := gossh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)

	m["SHA256"] = gossh.FingerprintSHA256(key.PublicKey())
	m["MD5"] = gossh.FingerprintLegacyMD5(key.PublicKey())
	return m, nil
}
