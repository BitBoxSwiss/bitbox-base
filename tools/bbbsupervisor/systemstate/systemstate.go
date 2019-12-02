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

// MapDescriptionCodeStateCode defines the corresponding the StateCode for a DescriptionCode.
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

// These LogTags are shared strings between subsystems on the Base and the
// Supervisor. The supervisor watches the log of the subsystems on the Base and
// triggers a state change.
const (
	// LogTagMWUpdateStart is logged by the middleware when the update progress
	// is started. This triggers the Supervisor to set the
	// `BitBoxBaseHeartbeatRequest_DOWNLOAD_UPDATE` descriptionCode to active and
	// sets the `BitBoxBaseHeartbeatRequest_UPDATE_FAILED` descriptionCode to
	// inactive.
	LogTagMWUpdateStart string = "LogTag:Middleware:Base_Image_Update_Start"

	// LogTagMWUpdateSuccess is logged by the middleware when the update progress
	// ends with the success case. This triggers the supervisor to set the
	// `BitBoxBaseHeartbeatRequest_DOWNLOAD_UPDATE` descriptionCode to inactive.
	LogTagMWUpdateSuccess string = "LogTag:Middleware:Base_Image_Update_Success"

	// LogTagMWUpdateFailure is logged by the middleware when the update progress
	// ends with the success case. This triggers the supervisor to set the
	// `BitBoxBaseHeartbeatRequest_DOWNLOAD_UPDATE` descriptionCode to inactive
	// and sets the `BitBoxBaseHeartbeatRequest_UPDATE_FAILED` descriptionCode to
	// active.
	LogTagMWUpdateFailure string = "LogTag:Middleware:Base_Image_Update_Failure"

	// LogTagMWReboot is logged by the middleware when a Base reboot is started
	// via RPC. This triggers the supervisor to set the descriptionCode
	// `BitBoxBaseHeartbeatRequest_REBOOT` to active. This descriptionCode is
	// reset on every start of the Supervisor.
	LogTagMWReboot string = "LogTag:Middleware:Base_Reboot"

	// LogTagMWShutdown is logged by the middleware when a Base shutdown is
	// started via RPC. This triggers the supervisor to set the descriptionCode
	// `BitBoxBaseHeartbeatRequest_SHUTDOWN` to active. This descriptionCode is
	// reset on every start of the Supervisor.
	LogTagMWShutdown string = "LogTag:Middleware:Base_Shutdown"
)
