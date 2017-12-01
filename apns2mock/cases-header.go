// Copyright 2017 Aleksey Blinov. All rights reserved.

package apns2mock

import (
	"strconv"

	"github.com/satori/go.uuid"
)

// HeaderHandlers deal with request headers.
//
// - Invalid UUID format in APNS Id is 400, "BadMessageId".
// - Priority that is not empty, 5 or 10 is 400, "BadPriority".
// - Empty topic is 400, "MissingTopic".
// - Topics starting with 'd' are 400, "TopicDisallowed".
// - Collapse Ids longer than 64 are 400, "BadCollapseId".
// - Expiration date that cannot be parsed is 400, "BadExpirationDate".
//
var HeaderHandlers = []func(req *APNSRequest) (int, string){
	func(req *APNSRequest) (int, string) {
		if h := req.Header.Get("apns-id"); h != "" {
			if _, err := uuid.FromString(h); err != nil {
				return 400, "BadMessageId"
			}
		}
		switch h := req.Header.Get("apns-priority"); h {
		case "", "5", "10":
		default:
			return 400, "BadPriority"
		}
		if h := req.Header.Get("apns-topic"); h == "" {
			return 400, "MissingTopic"
		}
		if h := req.Header.Get("apns-topic"); h != "" && h[0] == 'd' {
			return 400, "TopicDisallowed"
		}
		if h := req.Header.Get("apns-collapse-id"); len(h) > 64 {
			return 400, "BadCollapseId"
		}
		if h := req.Header.Get("apns-expiration"); h != "" {
			if _, err := strconv.ParseInt(h, 10, 64); err != nil {
				return 400, "BadExpirationDate"
			}
		}
		return 0, ""
	},
}
