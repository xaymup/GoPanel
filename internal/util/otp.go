package util

import (
    "github.com/pquerna/otp/totp"
    "github.com/skip2/go-qrcode"
    "fmt"
    "log"
)

func Generate2FAQRCode() ([]byte, string, error) {

    log.Println("Generating OTP key")

    // Generate a new OTP key
	ip, _ := GetServerIP() 
    key, err := totp.Generate(totp.GenerateOpts{
        Issuer:      "GoPanel",
        AccountName: fmt.Sprintf("admin@%s",ip),
    })
    if err != nil {
        return nil, "", err
    }

    // Generate a URL to encode in the QR code
    url := key.URL()

    // Generate the QR code image
    qrCode, err := qrcode.Encode(url, qrcode.Medium, 256)
    if err != nil {
        return nil, "", err
    }

    return qrCode, key.Secret(), nil
}

func ValidateOTP(otp, Secret string) (bool, error) {
    fileSecret, err := ReadSecretFromFile()
    if err == nil && fileSecret != "" {
        // If reading the file was successful and we got a non-empty secret
        Secret = fileSecret
    } else if err != nil {
        // Handle the error (log it or print it)
        fmt.Println("Failed to read secret from file:", err)
    }
    valid := totp.Validate(otp, Secret)
    return valid, nil
}

    // Read the secret from the file
    // secret, err := ReadSecretFromFile()
    // if err != nil {
    //     return false, err
    // }
    // Validate the OTP using the secret