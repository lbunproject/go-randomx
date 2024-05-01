//go:build !amd64 || purego

package aes

const HasHardAESImplementation = false

func NewHardAES() AES {
	return nil
}
