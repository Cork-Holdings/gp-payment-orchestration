package merchantapis

import (
	"errors"
	"os"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantapikeys"
	"github.com/Cork-Holdings/gp_payment_orchestration/utils"
)

func FindMerchantByClientID(
	app *global.App,
	clientID string,
) (*merchantapikeys.MerchantAPIKey, error) {

	encryptionKey := []byte(os.Getenv("ENCRYPTION_KEY"))

	var merchants []merchantapikeys.MerchantAPIKey

	if err := app.DB.Find(&merchants).Error; err != nil {
		return nil, err
	}

	for _, merchant := range merchants {

		decryptedID, err := utils.Decrypt(
			merchant.ClientID,
			encryptionKey,
		)

		if err != nil {
			continue
		}

		if decryptedID == clientID {
			return &merchant, nil
		}
	}

	return nil, errors.New("merchant not found")
}

func GetProviderFromPhoneNumber(phoneNumber string) string {
	if len(phoneNumber) < 5 {
		return "Invalid phone number"
	}

	prefix := phoneNumber[:5]

	switch prefix {
	case "26096", "26076":
		return "mtn"

	case "26097", "26077", "26057":
		return "airtel"

	case "26095", "26075":
		return "zamtel"

	default:
		return "Unsupported Provider for " + phoneNumber
	}
}
