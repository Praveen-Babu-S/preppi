package handler

import (
	"strconv"
	"time"
)

func parseID(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

func uintToStr(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

func unixTime(sec int64) time.Time {
	return time.Unix(sec, 0)
}

func optionalUint(id uint) string {
	if id == 0 {
		return ""
	}
	return uintToStr(id)
}
