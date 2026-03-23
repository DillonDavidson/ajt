package main

import (
	"errors"
)

func shiftTime(timestamp string, deltaSeconds float32, subType SubtitleType) (string, error) {
	switch subType {
	case SRT:
		return shiftSRTTime(timestamp, deltaSeconds)
	case ASS:
		return shiftASSTime(timestamp, deltaSeconds)
	default:
		return "", errors.New("no idea what subtitle type")
	}
}
