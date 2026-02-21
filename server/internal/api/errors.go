package api

const (
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeInvalidRequest = "INVALID_REQUEST"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeInternal       = "INTERNAL_ERROR"

	ErrCodePolicyExceedsSingleTxLimit  = "POLICY_EXCEEDS_SINGLE_TX_LIMIT"
	ErrCodePolicyExceedsDailyLimit     = "POLICY_EXCEEDS_DAILY_LIMIT"
	ErrCodePolicyExceedsDailyTxLimit   = "POLICY_EXCEEDS_DAILY_TX_LIMIT"
	ErrCodePolicyAddressNotWhitelisted = "POLICY_ADDRESS_NOT_WHITELISTED"
	ErrCodePolicyOutsideTimeWindow     = "POLICY_OUTSIDE_TIME_WINDOW"
	ErrCodePolicyInvalidAmount         = "POLICY_INVALID_AMOUNT"
)
