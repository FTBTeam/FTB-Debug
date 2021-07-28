package main

import "os/user"

type (
	FTBApp struct {
		User            *user.User
		OWLocation      string
		InstallLocation string
		AppVersion      string
		JarVersion      string
		WebVersion      string
		AppBranch       string
		Structure       AppStructure
	}

	AppStructure struct {
		MCBin MinecraftBin
	}

	MinecraftBin struct {
		Exists  bool
		Profile bool
	}

	VersionJson struct {
		JarVersion string `json:"jarVersion"`
		WebVersion string `json:"webVersion"`
		Branch     string `json:"branch"`
	}
)
