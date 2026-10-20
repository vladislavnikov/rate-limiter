package types

import "time"

type Client struct {
	requestsCount int
	windowEnd     time.Time
}
