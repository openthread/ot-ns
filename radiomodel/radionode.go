package radiomodel

// RadioNode is the status of a single radio node of the radio model, used by all radio models.
type RadioNode struct {
	// IsCcaFailed tracks whether the last CCA process failed (true), or not (false).
	IsCcaFailed bool

	// TxPhase tracks the current Tx phase. 0 = Not started. >0 is started (exact value depends on radio model)
	TxPhase int
}
