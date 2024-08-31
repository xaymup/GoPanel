package util

import (
    "github.com/pquerna/otp/totp"
    "github.com/skip2/go-qrcode"
    "fmt"
)

func Generate2FAQRCode() ([]byte, string, error) {
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

func ValidateOTP(otp string) (bool, error) {

    // Read the secret from the file
    secret, err := ReadSecretFromFile()
    if err != nil {
        return false, err
    }

    // Validate the OTP using the secret
    valid := totp.Validate(otp, secret)
    return valid, nil
}