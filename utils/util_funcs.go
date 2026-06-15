package utils

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func CapitalizeError(msg string) error {
	return errors.New(strings.ToUpper(msg[:1]) + msg[1:])
}

// GetNetworkProvider returns the network provider based on the phone number prefix
func GetNetworkProvider(phonenumber string) (string, error) {

	if strings.HasPrefix(phonenumber, "+26") {
		phonenumber = strings.TrimPrefix(phonenumber, "+26")
	} else if strings.HasPrefix(phonenumber, "26") {
		phonenumber = strings.TrimPrefix(phonenumber, "26")
	}
	// Ensure the number has at least 9 digits (after removing country code)
	if len(phonenumber) < 9 {
		return "", errors.New("invalid phone number format")
	}

	// Extract the prefix (first three digits)
	prefix := phonenumber[:3]

	// Map of Zambian mobile network prefixes
	providerMap := map[string]string{
		"096": "mtn",
		"076": "mtn",
		"056": "mtn",
		"097": "airtel",
		"077": "airtel",
		"057": "airtel",
		"095": "zamtel",
		"075": "zamtel",
		"055": "zamtel",
	}

	// Get the provider based on the prefix
	if provider, exists := providerMap[prefix]; exists {
		return provider, nil
	}

	return "", errors.New("unknown network provider")
}

func GenerateTenDigitCode() string {
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(9000000000) + 100000000
	return fmt.Sprintf("%d", code)
}

func GenerateSixDigitCode() string {
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(900000) + 100000
	return fmt.Sprintf("%d", code)
}

func GetIPAddress(ctx context.Context) string {
	ip, ok := ctx.Value("ip").(string)

	if !ok {
		return "unkwown"
	}

	return ip

}
