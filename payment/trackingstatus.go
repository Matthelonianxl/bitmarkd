// Copyright (c) 2014-2016 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package payment

import ()

// result of track payment/try proof
type TrackingStatus int

// possible status values
const (
	TrackingNotFound  TrackingStatus = iota
	TrackingAccepted  TrackingStatus = iota
	TrackingProcessed TrackingStatus = iota
	TrackingInvalid   TrackingStatus = iota
)

// convert the tracking value for printf
func (ts TrackingStatus) String() string {
	switch ts {
	case TrackingNotFound:
		return "NotFound"
	case TrackingAccepted:
		return "Accepted"
	case TrackingProcessed:
		return "Processed"
	case TrackingInvalid:
		return "Invalid"
	default:
		return "*Unknown*"
	}
}

// convert the tracking value for JSON
func (ts TrackingStatus) MarshalText() ([]byte, error) {
	buffer := []byte(ts.String())
	return buffer, nil
}
