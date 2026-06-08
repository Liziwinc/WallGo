package base62

import "math/rand/v2"


const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateHash() string{
	b := make([]byte, 6)
	for i := range b{
		b[i] = alphabet[rand.N(len(alphabet))]
	}
	return string(b)
}