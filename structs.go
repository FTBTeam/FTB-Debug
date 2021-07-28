package main

import "os/user"

type VersionJson struct {
	JarVersion string `json:"jarVersion"`
	WebVersion string `json:"webVersion"`
	Branch     string `json:"branch"`
}

type (
	FTBApp struct {
		User            *user.User
		OWLocation      string
		InstallLocation string
		AppVersion      string
		JarVersion      string
		WebVersion      string
		AppBranch       string
	}
)
