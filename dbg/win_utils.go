//go:build windows

package dbg

import (
	"fmt"
	wmi "github.com/yusufpapurcu/wmi"
)

type (
	Win32_OperatingSystem struct {
		Caption string
		Version string
	}
)

func getSysInfo() (oSystem string, err error) {
	var dst []Win32_OperatingSystem

	q := wmi.CreateQuery(&dst, "")
	err = wmi.Query(q, &dst)
	if err != nil {
		return "", err
	}
	oSystem = fmt.Sprintf("%s (%s)", dst[0].Caption, dst[0].Version)
	return oSystem, nil
}
