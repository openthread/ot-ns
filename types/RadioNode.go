package types

type RadioNode interface {
	// IsCcaFailed checks whether the CCA process for this RadioNode failed (true), or not (false).
	IsCcaFailed() bool

	// TxPhase gets the current Tx phase of this RadioNode. 0 = Not started.
	TxPhase() uint16

	// SetCcaFailed sets the CCA-failed status to failed (true), or not (false).
	SetCcaFailed(status bool)

	// SetTxPhase sets the current Tx phase of this RadioNode.
	SetTxPhase(phase uint16)

	// AdvanceTxPhase advances the TxPhase by 1
	AdvanceTxPhase()

	// AddEvent adds a new Event to the event queue of this RadioNode
	AddEvent(evt *Event)
}
