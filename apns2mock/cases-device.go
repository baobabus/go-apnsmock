// Copyright 2017 Aleksey Blinov. All rights reserved.

package apns2mock

// DeviceTokenHandlers deal with device tokens.
//
// - Device tokens starting with '1' are 400, "BadDeviceToken".
// - Device tokens starting with '2' are 410, "Unregistered".
// - Device tokens starting with the same letter/digit as APNS topic
//   are 400, "DeviceTokenNotForTopic".
//
var DeviceTokenHandlers = []func(req *APNSRequest) (int, string){
	func(req *APNSRequest) (int, string) {
		if req.DeviceToken[0] == '1' {
			return 400, "BadDeviceToken"
		}
		if req.DeviceToken[0] == '2' {
			return 410, "Unregistered"
		}
		if t := req.Header.Get("apns-topic"); t != "" && req.DeviceToken[0] == t[0] {
			return 400, "DeviceTokenNotForTopic"
		}
		return 0, ""
	},
}
