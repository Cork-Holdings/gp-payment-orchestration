package feecalculator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Cork-Holdings/gp_payment_orchestration/internal/global"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/feeprofiles"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantfeeprofiles"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/merchantpaymentchannels"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/paymentchannels"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/prefixes"
	"github.com/Cork-Holdings/gp_payment_orchestration/internal/modules/transactiontypes"
	"github.com/redis/go-redis/v9"
)

type profileFeeBandResult struct {
	FeeAmount float64
	BandID    string
	BandRange string
	BandType  string
	BandRate  float64
}

func calculateTransactionFeeByChargeType(
	ctx context.Context,
	cache *redis.Client,
	feeProfile feeprofiles.FeeProfile,
	amount float64,
) (*profileFeeBandResult, error) {
	chargeType := feeProfile.ChargeType
	if chargeType == "" {
		chargeType = "percentage"
	}

	result := &profileFeeBandResult{}

	switch chargeType {
	case "fixed":
		result.FeeAmount = math.Round(feeProfile.ChargeAmount*100) / 100
		// Apply minimum fee if configured and the fixed charge is below it
		if feeProfile.MinimumFee > 0 {
			result.FeeAmount = math.Max(result.FeeAmount, math.Round(feeProfile.MinimumFee*100)/100)
		}
		result.BandType = "fixed"
		result.BandRate = feeProfile.ChargeAmount

	case "percentage":
		result.FeeAmount = math.Round(amount*(feeProfile.ChargeAmount/100)*100) / 100
		if feeProfile.MinimumFee > 0 {
			result.FeeAmount = math.Max(result.FeeAmount, math.Round(feeProfile.MinimumFee*100)/100)
		}
		result.BandType = "percentage"
		result.BandRate = feeProfile.ChargeAmount

	case "band":
		bandListKey := fmt.Sprintf("profile_band_list_%s", feeProfile.ID)
		var bands []feeprofiles.ProfileFeeBands

		val, err := cache.Get(ctx, bandListKey).Result()
		if err == nil {
			json.Unmarshal([]byte(val), &bands)
		} else {
			if err := global.GetDB().Where("fee_profile_id = ? AND status = ?", feeProfile.ID, "active").
				Order("min_amount").Find(&bands).Error; err != nil {
				log.Printf("❌ No profile fee bands found for fee profile %s", feeProfile.ID)
				return nil, fmt.Errorf("fee configuration error")
			}
			data, _ := json.Marshal(bands)
			cache.Set(ctx, bandListKey, data, 30*time.Minute)
		}

		var selectedBand *feeprofiles.ProfileFeeBands
		for i := range bands {
			b := &bands[i]
			if b.MinAmount <= amount && (b.MaxAmount == 0 || b.MaxAmount >= amount) {
				selectedBand = b
				break
			}
		}

		if selectedBand == nil {
			log.Printf("❌ No profile fee band found for amount %f on fee profile %s", amount, feeProfile.ID)
			return nil, fmt.Errorf("no applicable fee band for this amount")
		}

		if selectedBand.ChargeType == "fixed" {
			result.FeeAmount = math.Round(selectedBand.ChargeAmount*100) / 100
		} else {
			result.FeeAmount = math.Round(amount*(selectedBand.ChargeAmount/100)*100) / 100
		}

		// Apply minimum fee if configured on the fee profile
		if feeProfile.MinimumFee > 0 {
			result.FeeAmount = math.Max(result.FeeAmount, math.Round(feeProfile.MinimumFee*100)/100)
		}

		result.BandID = selectedBand.ID.String()
		maxAmountStr := "∞"
		if selectedBand.MaxAmount > 0 {
			maxAmountStr = fmt.Sprintf("%.2f", selectedBand.MaxAmount)
		}
		result.BandRange = fmt.Sprintf("%.2f - %s", selectedBand.MinAmount, maxAmountStr)
		result.BandType = selectedBand.ChargeType
		result.BandRate = selectedBand.ChargeAmount

	default:
		result.FeeAmount = math.Round(amount*(feeProfile.ChargeAmount/100)*100) / 100
		result.BandType = "percentage"
		result.BandRate = feeProfile.ChargeAmount
	}

	return result, nil
}

func CalculateFees(merchantId string, phoneNumber string, amount float64, transactionTypeCode TransactionTypeCode) (*FeeCalculationResult, error) {
	ctx := context.Background()
	db := global.GetDB()
	cache := global.GetCache()
	log.Printf("fee calculation started merchant_id=%s amount=%.2f transaction_type=%s", merchantId, amount, transactionTypeCode)

	// 1. Extract prefix
	if len(phoneNumber) < 5 {
		return &FeeCalculationResult{Error: "Invalid phone number", Status: "error"}, nil
	}
	prefix := phoneNumber[:5]
	prefixKey := fmt.Sprintf("prefix_%s", prefix)

	var prefixRecord prefixes.Prefix
	val, err := cache.Get(ctx, prefixKey).Result()
	if err == nil {
		json.Unmarshal([]byte(val), &prefixRecord)
	} else {
		if err := db.Where("prefix = ?", prefix).First(&prefixRecord).Error; err != nil {
			log.Printf("❌ Prefix not found: %s", prefix)
			return &FeeCalculationResult{Error: "Prefix resolution error", Status: "error"}, nil
		}
		data, _ := json.Marshal(prefixRecord)
		cache.Set(ctx, prefixKey, data, time.Hour)
	}

	// 2. Resolve transaction type
	txnTypeKey := fmt.Sprintf("txn_type_%s", transactionTypeCode)
	var transactionType transactiontypes.TransactionType
	val, err = cache.Get(ctx, txnTypeKey).Result()
	if err == nil {
		json.Unmarshal([]byte(val), &transactionType)
	} else {
		if err := db.Where("code = ?", transactionTypeCode).First(&transactionType).Error; err != nil {
			log.Printf("❌ Invalid transaction type code: %s", transactionTypeCode)
			return &FeeCalculationResult{Error: "Processing configuration error. Please contact support.", Status: "error"}, nil
		}
		data, _ := json.Marshal(transactionType)
		cache.Set(ctx, txnTypeKey, data, time.Hour)
	}

	// 3. Resolve payment channel by prefix + txn_type
	channelKey := fmt.Sprintf("channel_%s_%s", prefix, transactionType.ID)
	var paymentChannel paymentchannels.PaymentChannel
	val, err = cache.Get(ctx, channelKey).Result()
	if err == nil {
		json.Unmarshal([]byte(val), &paymentChannel)
	} else {
		var prefixPaymentChannel prefixes.PrefixPaymentChannel
		if err := db.Joins("PaymentChannel").
			Where("prefix_id = ? AND \"PaymentChannel\".transaction_type_id = ? AND \"PaymentChannel\".status = ?", prefixRecord.ID, transactionType.ID, "active").
			First(&prefixPaymentChannel).Error; err != nil {
			log.Printf("❌ No payment channel found for prefix %s and transaction type %s", prefix, transactionTypeCode)
			return &FeeCalculationResult{Error: "Service temporarily unavailable. Please try again later or contact support.", Status: "error"}, nil
		}
		paymentChannel = prefixPaymentChannel.PaymentChannel
		data, _ := json.Marshal(paymentChannel)
		cache.Set(ctx, channelKey, data, time.Hour)
	}

	// 4. Validate merchant-channel relationship
	merchantChannelKey := fmt.Sprintf("merchant_channel_%s_%s", merchantId, paymentChannel.ID)
	var merchantChannel merchantpaymentchannels.MerchantPaymentChannel
	val, err = cache.Get(ctx, merchantChannelKey).Result()
	if err == nil {
		json.Unmarshal([]byte(val), &merchantChannel)
	} else {
		if err := db.Where("merchant_id = ? AND payment_channel_id = ? AND status = ? AND approval_status = ?", merchantId, paymentChannel.ID, "active", "approved").
			First(&merchantChannel).Error; err != nil {
			log.Printf("❌ Merchant %s not approved for channel %s", merchantId, paymentChannel.ID)
			return &FeeCalculationResult{Error: "Merchant not authorized for this channel", Status: "error"}, nil
		}
		data, _ := json.Marshal(merchantChannel)
		cache.Set(ctx, merchantChannelKey, data, 30*time.Minute)
	}

	// 5. Resolve fee profile
	feeProfileKey := fmt.Sprintf("fee_profile_%s_%s_%s", merchantId, paymentChannel.SubscriptionID, transactionType.ID)
	var merchantFeeProfile merchantfeeprofiles.MerchantFeeProfile
	val, err = cache.Get(ctx, feeProfileKey).Result()
	if err == nil {
		json.Unmarshal([]byte(val), &merchantFeeProfile)
	} else {
		if err := db.Preload("FeeProfile").
			Joins("JOIN fee_profiles ON fee_profiles.id = merchant_fee_profiles.fee_profile_id").
			Where("merchant_fee_profiles.merchant_id = ? AND fee_profiles.payment_channel_id = ? AND fee_profiles.transaction_type_id = ? AND fee_profiles.status = ?",
				merchantId, paymentChannel.ID, transactionType.ID, "active").
			First(&merchantFeeProfile).Error; err != nil {
			log.Printf("❌ No fee profile for merchant %s and transaction type %s", merchantId, transactionTypeCode)
			return &FeeCalculationResult{Error: "Unable to process at the moment. Please contact support.", Status: "error"}, nil
		}
		data, _ := json.Marshal(merchantFeeProfile)
		cache.Set(ctx, feeProfileKey, data, 30*time.Minute)
	}

	feeProfile := merchantFeeProfile.FeeProfile
	transactionFeePercent := feeProfile.ChargeAmount
	providerFeePercent, _ := strconv.ParseFloat(paymentChannel.ProviderFee, 64)
	calculationMode := feeProfile.CalculationMode
	if calculationMode == "" {
		calculationMode = "standard"
	}
	isMergedMode := calculationMode == "merged"

	// 6. Calculate provider fee
	providerFeeAmount := 0.00
	var result FeeCalculationResult

	if paymentChannel.FeeType == "band" {
		bandListKey := fmt.Sprintf("channel_band_list_%s", paymentChannel.ID)
		var bands []paymentchannels.ChannelFeeBands
		val, err = cache.Get(ctx, bandListKey).Result()
		if err == nil {
			json.Unmarshal([]byte(val), &bands)
		} else {
			if err := db.Where("payment_channel_id = ? AND status = ?", paymentChannel.ID, "active").
				Order("min_amount").Find(&bands).Error; err != nil {
				log.Printf("❌ No MNO bands found for channel %s", paymentChannel.ID)
				return &FeeCalculationResult{Error: "Unable to process request. Please try again later.", Status: "error"}, nil
			}
			data, _ := json.Marshal(bands)
			cache.Set(ctx, bandListKey, data, 30*time.Minute)
		}

		var selectedBand *paymentchannels.ChannelFeeBands
		for _, b := range bands {
			if b.MinAmount <= amount && (b.MaxAmount == 0 || b.MaxAmount >= amount) {
				selectedBand = &b
				break
			}
		}

		if selectedBand == nil {
			log.Printf("❌ No MNO band found for amount %f on channel %s", amount, paymentChannel.ID)
			return &FeeCalculationResult{Error: "Unable to process request. Please try again later.", Status: "error"}, nil
		}

		if selectedBand.ChargeType == "fixed" {
			providerFeeAmount = selectedBand.ChargeAmount
		} else {
			providerFeeAmount = math.Round(amount*(selectedBand.ChargeAmount/100)*100) / 100
		}

		result.ProviderFeeBandID = selectedBand.ID.String()
		maxAmountStr := "∞"
		if selectedBand.MaxAmount > 0 {
			maxAmountStr = fmt.Sprintf("%.2f", selectedBand.MaxAmount)
		}
		result.ProviderFeeBandRange = fmt.Sprintf("%.2f - %s", selectedBand.MinAmount, maxAmountStr)
		result.ProviderFeeBandType = selectedBand.ChargeType
		result.ProviderFeeBandRate = selectedBand.ChargeAmount
	} else {
		if paymentChannel.FeeType == "percentage" {
			providerFeeAmount = math.Round(amount*(providerFeePercent/100)*100) / 100
		} else {
			providerFeeAmount = providerFeePercent
		}
		result.ProviderFeePercent = providerFeePercent
	}

	// 7. Handle disburse logic
	// Disbursement: fee is added on top of amount (merchant pays amount + fee)
	// Commission = transaction fee - provider fee
	if transactionTypeCode == TransactionTypeDisbursement {
		feeResult, err := calculateTransactionFeeByChargeType(ctx, cache, feeProfile, amount)
		if err != nil {
			return &FeeCalculationResult{Error: "Unable to calculate fees. Please contact support.", Status: "error"}, nil
		}

		transactionFeeAmount := feeResult.FeeAmount

		grossAmount := math.Round((amount+transactionFeeAmount)*100) / 100
		netAmount := math.Round(amount*100) / 100
		commissionFeeAmount := math.Round((transactionFeeAmount-providerFeeAmount)*100) / 100
		commissionFeePercent := 0.0
		if commissionFeeAmount > 0 {
			commissionFeePercent = math.Round((commissionFeeAmount/amount)*100*100) / 100
		}

		result.TransactionFeePercent = transactionFeePercent
		result.TransactionFeeAmount = transactionFeeAmount
		result.ProviderFeeAmount = providerFeeAmount
		result.TotalFeeAmount = transactionFeeAmount
		result.CommissionFeePercent = commissionFeePercent
		result.CommissionFeeAmount = commissionFeeAmount
		result.GrossAmount = grossAmount
		result.NetAmount = netAmount
		result.FeeProfileID = feeProfile.ID.String()
		result.TransactionType = transactionType.ID.String()
		result.PaymentChannelID = paymentChannel.ID.String()

		// Add profile fee band metadata if band-based
		if feeResult.BandID != "" {
			result.ProfileFeeBandID = feeResult.BandID
			result.ProfileFeeBandRange = feeResult.BandRange
			result.ProfileFeeBandType = feeResult.BandType
			result.ProfileFeeBandRate = feeResult.BandRate
		}

		log.Printf("fee calculation completed merchant_id=%s transaction_type=%s calculation_mode=%s amount=%.2f transaction_fee=%.2f provider_fee=%.2f gross_amount=%.2f commission_fee=%.2f net_amount=%.2f fee_profile_id=%s payment_channel_id=%s",
			merchantId, transactionTypeCode, calculationMode, amount, transactionFeeAmount, providerFeeAmount, grossAmount, commissionFeeAmount, netAmount, feeProfile.ID, paymentChannel.ID)

		return &result, nil
	}

	// 8. Handle collect logic
	feeResult, err := calculateTransactionFeeByChargeType(ctx, cache, feeProfile, amount)
	if err != nil {
		return &FeeCalculationResult{Error: "Unable to calculate fees. Please contact support.", Status: "error"}, nil
	}

	if isMergedMode {
		// Merged mode: merchant fee + provider fee = total fee charged to merchant
		// Commission = merchant fee (platform keeps this, provider fee goes to MNO)
		merchantFeeAmount := feeResult.FeeAmount

		totalFeeAmount := math.Round((merchantFeeAmount+providerFeeAmount)*100) / 100
		grossAmount := math.Round(amount*100) / 100 // For collections, gross = amount collected
		netAmount := math.Round((amount-totalFeeAmount)*100) / 100
		commissionFeeAmount := merchantFeeAmount
		commissionFeePercent := math.Round((commissionFeeAmount/amount)*100*100) / 100

		result.TransactionFeePercent = transactionFeePercent
		result.TransactionFeeAmount = totalFeeAmount
		result.TotalFeeAmount = totalFeeAmount
		result.ProviderFeeAmount = providerFeeAmount
		result.CommissionFeePercent = commissionFeePercent
		result.CommissionFeeAmount = commissionFeeAmount
		result.GrossAmount = grossAmount
		result.NetAmount = netAmount
		result.FeeProfileID = feeProfile.ID.String()
		result.TransactionType = transactionType.ID.String()
		result.PaymentChannelID = paymentChannel.ID.String()

		// Add profile fee band metadata if band-based
		if feeResult.BandID != "" {
			result.ProfileFeeBandID = feeResult.BandID
			result.ProfileFeeBandRange = feeResult.BandRange
			result.ProfileFeeBandType = feeResult.BandType
			result.ProfileFeeBandRate = feeResult.BandRate
		}

		log.Printf("fee calculation completed merchant_id=%s transaction_type=%s calculation_mode=merged amount=%.2f transaction_fee=%.2f provider_fee=%.2f gross_amount=%.2f commission_fee=%.2f net_amount=%.2f fee_profile_id=%s payment_channel_id=%s",
			merchantId, transactionTypeCode, amount, totalFeeAmount, providerFeeAmount, grossAmount, commissionFeeAmount, netAmount, feeProfile.ID, paymentChannel.ID)

		return &result, nil
	} else {
		// Standard mode: transaction fee is the total fee charged to merchant
		// Ensure fee covers provider cost; if provider fee exceeds transaction fee, upcharge customer
		transactionFeeAmount := feeResult.FeeAmount

		// Avoid negative commission: total fee must be at least the provider fee
		totalFeeToCharge := transactionFeeAmount
		feeUplift := 0.0 // Track any upcharge needed to cover provider fees

		if providerFeeAmount > transactionFeeAmount {
			totalFeeToCharge = providerFeeAmount
			feeUplift = math.Round((providerFeeAmount-transactionFeeAmount)*100) / 100
		}

		commissionFeeAmount := math.Round((totalFeeToCharge-providerFeeAmount)*100) / 100
		commissionFeePercent := 0.0
		if commissionFeeAmount > 0 {
			commissionFeePercent = math.Round((commissionFeeAmount/amount)*100*100) / 100
		}

		grossAmount := math.Round(amount*100) / 100 // For collections, gross = amount collected
		netAmount := math.Round((amount-totalFeeToCharge)*100) / 100

		result.TransactionFeePercent = transactionFeePercent
		result.TransactionFeeAmount = totalFeeToCharge
		result.TotalFeeAmount = totalFeeToCharge
		result.ProviderFeeAmount = providerFeeAmount
		result.CommissionFeePercent = commissionFeePercent
		result.CommissionFeeAmount = commissionFeeAmount
		result.GrossAmount = grossAmount
		result.NetAmount = netAmount
		result.FeeProfileID = feeProfile.ID.String()
		result.TransactionType = transactionType.ID.String()
		result.PaymentChannelID = paymentChannel.ID.String()

		// Store fee uplift metadata for audit trail
		result.FeeUplift = feeUplift

		// Add profile fee band metadata if band-based
		if feeResult.BandID != "" {
			result.ProfileFeeBandID = feeResult.BandID
			result.ProfileFeeBandRange = feeResult.BandRange
			result.ProfileFeeBandType = feeResult.BandType
			result.ProfileFeeBandRate = feeResult.BandRate
		}

		log.Printf("fee calculation completed merchant_id=%s transaction_type=%s calculation_mode=standard amount=%.2f transaction_fee=%.2f provider_fee=%.2f fee_uplift=%.2f gross_amount=%.2f commission_fee=%.2f net_amount=%.2f fee_profile_id=%s payment_channel_id=%s",
			merchantId, transactionTypeCode, amount, transactionFeeAmount, providerFeeAmount, feeUplift, grossAmount, commissionFeeAmount, netAmount, feeProfile.ID, paymentChannel.ID)

		return &result, nil
	}
}

func CalculateFee(req *feeprofiles.CalculateFeeRequest) (*feeprofiles.CalculateFeeResponse, error) {
	txnType := TransactionTypeDisbursement
	if strings.ToUpper(req.TransactionType) == "MNO_COLLECTION" || strings.ToUpper(req.TransactionType) == "COLLECTION" {
		txnType = TransactionTypeCollection
	}

	result, err := CalculateFees(req.MerchantID, req.PhoneNumber, req.Amount, txnType)
	if err != nil {
		return nil, err
	}
	if result.Status == "error" {
		return nil, fmt.Errorf(result.Error)
	}

	return &feeprofiles.CalculateFeeResponse{
		FeeAmount:        result.TotalFeeAmount,
		FeePercentage:    result.TransactionFeePercent,
		FeeCurrency:      req.Currency,
		PaymentChannelID: result.PaymentChannelID,
		TransactionType:  result.TransactionType,
	}, nil
}
