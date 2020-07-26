package gmlserver

import (
          cryptorandom "crypto/rand"

          "encoding/base64"
)
//Base64URLEncode takes in []byte and encodes it to base64URL with NO padding
func Base64URLEncode(data []byte) string {
        var result = base64.RawURLEncoding.EncodeToString(data)
        return result
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(s int) (string, error) {
        b, err := GenerateRandomBytes(s)
        return Base64URLEncode(b), err
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
        b := make([]byte, n)
        _, err := cryptorandom.Read(b)
        // Note that err == nil only if we read len(b) bytes.
        if err != nil {
                myLogger.Printf("cryptorandom read error: %v", err)
                return nil, err
        }

        return b, nil
}
