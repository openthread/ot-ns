package radiomodel

// RadioNode is the status of a single radio node of the radio model, used by all radio models.
type RadioNode struct {
	// IsCcaFailed tracks whether the last CCA process failed (true), or not (false).
	IsCcaFailed bool

	// IsTxFailed tracks whether the current/last Tx attempt failed (true), ot not (false).
	IsTxFailed bool

	// TxPhase tracks the current Tx phase. 0 = Not started. >0 is started (exact value depends on radio model)
	TxPhase int

	// TimeLastTxEnded is the timestamp (us) when the last Tx or Tx-attempt by this RadioNode ended.
	TimeLastTxEnded uint64

	// IsLastTxLong indicates whether the RadioNode's last Tx was a long frame (LIFS applies) or short (SIFS applies)
	IsLastTxLong bool
}
