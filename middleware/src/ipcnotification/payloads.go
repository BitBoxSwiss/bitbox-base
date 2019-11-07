package ipcnotification

/* This file holds the payload definitions for different notification topics */

// ParseMenderUpdatePayload parses payloads from notifications with the topic
// `mender-update`. The payload should have the following JSON structure:
//  {"success": true}
func ParseMenderUpdatePayload(payload interface{}) (menderUpdateSuccess bool, ok bool) {
	if payloadMap, ok := payload.(map[string]interface{}); ok {
		if val, ok := payloadMap["success"]; ok {
			if success, ok := val.(bool); ok {
				return success, true
			}
		}
	}
	return false, false
}
