package main

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "fmt"
)

func EncryptByGCM(key []byte, plainText string) ([]byte, error) {
    block, err := aes.NewCipher(key); if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block); if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())// Unique nonce is required(NonceSize 12byte)
    _, err = rand.Read(nonce); if err != nil {
        return nil, err
    }
    fmt.Printf("nonce: %012x\n", string(nonce))

    cipherText := gcm.Seal(nil, nonce, []byte(plainText), nil)
    cipherText = append(nonce, cipherText...)

    return cipherText, nil
}

func DecryptByGCM(key []byte, cipherText []byte) (string, error) {
    block, err := aes.NewCipher(key); if err != nil {
        return "", err
    }
    gcm, err := cipher.NewGCM(block); if err != nil {
        return "", err
    }

    nonce := cipherText[:gcm.NonceSize()]
    plainByte, err := gcm.Open(nil, nonce, cipherText[gcm.NonceSize():], nil); if err != nil {
        return "", err
    }
    fmt.Printf("nonce: %012x\n", string(nonce))

    return string(plainByte), nil
}

func main() {
    key := []byte("keyIn32Bytes01234567890123456789")
    cipherText, _ := EncryptByGCM(key, "12345")
    decryptedText, _ := DecryptByGCM(key, cipherText)
    fmt.Printf("Decrypted Text: %v\n", decryptedText)
}
