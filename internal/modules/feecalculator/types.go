package feecalculator

type TransactionTypeCode string

const (
	TransactionTypeCollection   TransactionTypeCode = "MNO_COLLECTION"
	TransactionTypeDisbursement TransactionTypeCode = "MNO_DISBURSEMENT"
)

type FeeCalculationResult struct {
	TransactionFeePercent float64 `json:"transaction_fee_percent"`
	TransactionFeeAmount  float64 `json:"transaction_fee_amount"`
	ProviderFeeAmount     float64 `json:"provider_fee_amount"`
	TotalFeeAmount        float64 `json:"total_fee_amount"`
	CommissionFeePercent  float64 `json:"commission_fee_percent"`
	CommissionFeeAmount   float64 `json:"commission_fee_amount"`
	GrossAmount           float64 `json:"gross_amount"`
	NetAmount             float64 `json:"net_amount"`
	FeeProfileID          string  `json:"fee_profile_id"`
	TransactionType       string  `json:"transaction_type"`
	PaymentChannelID      string  `json:"payment_channel_id"`

	// Provider fee band metadata (MNO/channel bands)
	ProviderFeeBandID    string  `json:"provider_fee_band_id,omitempty"`
	ProviderFeeBandRange string  `json:"provider_fee_band_range,omitempty"`
	ProviderFeeBandType  string  `json:"provider_fee_band_type,omitempty"`
	ProviderFeeBandRate  float64 `json:"provider_fee_band_rate,omitempty"`
	ProviderFeePercent   float64 `json:"provider_fee_percent,omitempty"`

	// Profile fee band metadata (merchant's tiered fee structure)
	ProfileFeeBandID    string  `json:"profile_fee_band_id,omitempty"`
	ProfileFeeBandRange string  `json:"profile_fee_band_range,omitempty"`
	ProfileFeeBandType  string  `json:"profile_fee_band_type,omitempty"`
	ProfileFeeBandRate  float64 `json:"profile_fee_band_rate,omitempty"`

	Error  string `json:"error,omitempty"`
	Status string `json:"status,omitempty"`
}
