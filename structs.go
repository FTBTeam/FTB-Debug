package main

import "os/user"

type (
	FTBApp struct {
		User *user.User
		InstallLocation string
		AppVersion string
	}
)
