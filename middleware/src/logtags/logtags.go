package logtags

// These LogTags are shared logtags between the Middleware and the Supervisor.
// The supervisor watches these and triggers the corresponding handler.
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
