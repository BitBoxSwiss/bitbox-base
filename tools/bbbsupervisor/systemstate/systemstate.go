// Package systemstate defines the Base system states and peripherals of those.
package systemstate

import "github.com/digitalbitbox/bitbox02-api-go/api/firmware/messages"

// MapDescriptionCodePriority defines the priority of the DescriptionCode.
var MapDescriptionCodePriority = map[messages.BitBoxBaseHeartbeatRequest_DescriptionCode]int{
	messages.BitBoxBaseHeartbeatRequest_EMPTY:                 0,
	messages.BitBoxBaseHeartbeatRequest_INITIAL_BLOCK_SYNC:    1200,
	messages.BitBoxBaseHeartbeatRequest_DOWNLOAD_UPDATE:       1400,
	messages.BitBoxBaseHeartbeatRequest_UPDATE_FAILED:         2300,
	messages.BitBoxBaseHeartbeatRequest_NO_NETWORK_CONNECTION: 2400,
	messages.BitBoxBaseHeartbeatRequest_REBOOT:                2500,
	messages.BitBoxBaseHeartbeatRequest_SHUTDOWN:              2510,
	messages.BitBoxBaseHeartbeatRequest_OUT_OF_DISK_SPACE:     3400,
	messages.BitBoxBaseHeartbeatRequest_REDIS_ERROR:           3900,
}

// MapDescriptionCodeStateCode maps a DescriptionCode to the corresponding StateCode.
var MapDescriptionCodeStateCode = map[messages.BitBoxBaseHeartbeatRequest_DescriptionCode]messages.BitBoxBaseHeartbeatRequest_StateCode{
	messages.BitBoxBaseHeartbeatRequest_EMPTY:                 messages.BitBoxBaseHeartbeatRequest_IDLE,
	messages.BitBoxBaseHeartbeatRequest_INITIAL_BLOCK_SYNC:    messages.BitBoxBaseHeartbeatRequest_WORKING,
	messages.BitBoxBaseHeartbeatRequest_DOWNLOAD_UPDATE:       messages.BitBoxBaseHeartbeatRequest_WORKING,
	messages.BitBoxBaseHeartbeatRequest_REBOOT:                messages.BitBoxBaseHeartbeatRequest_WARNING,
	messages.BitBoxBaseHeartbeatRequest_SHUTDOWN:              messages.BitBoxBaseHeartbeatRequest_WARNING,
	messages.BitBoxBaseHeartbeatRequest_UPDATE_FAILED:         messages.BitBoxBaseHeartbeatRequest_WARNING,
	messages.BitBoxBaseHeartbeatRequest_NO_NETWORK_CONNECTION: messages.BitBoxBaseHeartbeatRequest_WARNING,
	messages.BitBoxBaseHeartbeatRequest_OUT_OF_DISK_SPACE:     messages.BitBoxBaseHeartbeatRequest_ERROR,
	messages.BitBoxBaseHeartbeatRequest_REDIS_ERROR:           messages.BitBoxBaseHeartbeatRequest_ERROR,
}
