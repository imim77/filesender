package main

import (
	"time"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	writeWait  = 10 * time.Second
)

type ErrorMessage struct {
	Type string `json:"type"`
	Code int    `json:"code"`
}

type Client struct {
}
