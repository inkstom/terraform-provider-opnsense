package interfaces_test

import "regexp"

var (
	regexpLoopback = regexp.MustCompile(`^lo[0-9]+$`)
	regexpGif      = regexp.MustCompile(`^gif[0-9]+$`)
	regexpGre      = regexp.MustCompile(`^gre[0-9]+$`)
	regexpVxlan    = regexp.MustCompile(`^vxlan[0-9]+$`)
	regexpBridge   = regexp.MustCompile(`^bridge[0-9]+$`)
)
