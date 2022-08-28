package types

/**
* OT_ERROR_* error codes from OpenThread that are used by OT-NS.
* (See OpenThread error.h for details)
 */

// TODO currently these are stored in Param1 (uint8) so limited to 0...127.
const (
	OT_ERROR_NONE                   = 0
	OT_ERROR_ABORT                  = 11
	OT_ERROR_CHANNEL_ACCESS_FAILURE = 15
	OT_ERROR_FCS                    = 17
)
