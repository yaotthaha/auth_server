package lib

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

func RSAGen(bits uint64) (privateKey, publicKey []byte, err error) {
	privateKeyGen, err := rsa.GenerateKey(rand.Reader, int(bits))
	if err != nil {
		return []byte(""), []byte(""), err
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKeyGen)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	privateKey = pem.EncodeToMemory(block)
	publicKeyGen := &privateKeyGen.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKeyGen)
	if err != nil {
		return []byte(""), []byte(""), err
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	publicKey = pem.EncodeToMemory(block)
	return
}

func RSADecrypt(data_encrypt []byte, privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return []byte(""), errors.New("Block GEN Fail")
	}
	privatekeyInside, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return []byte(""), err
	}
	data, err := rsa.DecryptPKCS1v15(rand.Reader, privatekeyInside, data_encrypt)
	if err != nil {
		return []byte(""), err
	}
	return data, nil
}
