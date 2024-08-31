package util

import (
    "crypto/rand"
    "encoding/base64"
)

func GenerateSecretKey(length int) (string, error) {
    // Create a byte slice with the desired length
    bytes := make([]byte, length)

    // Fill the byte slice with random data
    _, err := rand.Read(bytes)
    if err != nil {
        return "", err
    }

    // Encode the byte slice to a base64 string
    // Using base64 encoding to avoid issues with non-printable characters
    randomString := base64.URLEncoding.EncodeToString(bytes)

    // Truncate the string to the desired length
    if length < len(randomString) {
        randomString = randomString[:length]
    }

    return randomString, nil
}