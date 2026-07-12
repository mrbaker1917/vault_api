package crypto

import (
	"fmt"
	
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type TOTPSetup struct {
	Secret string
	OTPAuthURL    string
}

func GenerateTOTPSecret(email string) (TOTPSetup, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Vault API",
		AccountName: email,
		SecretSize:  20,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
	})
	if err != nil {
		return TOTPSetup{}, fmt.Errorf("generate totp secret: %w", err)
	}
	return TOTPSetup{
		Secret: key.Secret(),
		OTPAuthURL: key.URL(),
	}, nil
}

func ValidateTOTPCode(secret, code string) bool {
    return totp.Validate(code, secret)
}