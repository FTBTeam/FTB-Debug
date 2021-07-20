package main

import "os/user"

type (
	FTBApp struct {
		User *user.User
		OWLocation string
		InstallLocation string
		AppVersion string
	}
)
