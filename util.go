package mercury

import (
	"crypto"
	"crypto/x509"
	"strings"
)

func splitPath(path string) []string {
	// if we have a single backslash and nothing else, we'll get
	// []string{"", ""} instead of []string{""}, which is what we want for
	// proper path matching.
	if path == "/" {
		return []string{""}
	}
	path = strings.TrimSuffix(path, "/") // if only some of the paths have
	// this and others don't, they won't match when they potentially should
	return strings.Split(path, "/")
}

// FingerprintCertificate computes a SHA1 hash of the raw certificate bytes.
func FingerprintCertificate(cert *x509.Certificate) []byte {
	return FingerprintCertificateWithHash(cert, crypto.SHA1)
}

// FingerprintCertificateWithHash computes a hash of a given type from the raw certificate bytes.
func FingerprintCertificateWithHash(cert *x509.Certificate, hashType crypto.Hash) []byte {
	hf := hashType.New()
	return hf.Sum(cert.Raw)
}
