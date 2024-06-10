package main

import (
	"net/http"
	"os"
	"path/filepath"
)

var (
	checkRequestsURLs = map[string]CheckURLStruct{
		"https://api.feed-the-beast.com/": {
			Method:             "HEAD",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusNotFound,
			ExpectedReponse:    "",
		},
		"https://meta.feed-the-beast.com/v1/health": {
			Method:             "HEAD",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusOK,
			ExpectedReponse:    "",
		},
		"https://api.modpacks.ch/public/api/ping": {
			Method:             "GET",
			ValidateResponse:   true,
			ExpectedStatusCode: http.StatusOK,
			ExpectedReponse:    "{\"status\":\"success\",\"reply\":\"pong\"}",
		},
		"https://maven.fabricmc.net": {
			Method:             "HEAD",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusOK,
			ExpectedReponse:    "",
		},
		"https://maven.neoforged.net/net/neoforged/neoforge/maven-metadata.xml": {
			Method:             "HEAD",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusOK,
			ExpectedReponse:    "",
		},
		"https://maven.minecraftforge.net/net/minecraftforge/forge/maven-metadata.xml": {
			Method:             "HEAD",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusOK,
			ExpectedReponse:    "",
		},
		"https://api.creeper.host/api/health": {
			Method:             "HEAD",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusOK,
			ExpectedReponse:    "",
		},
		"https://maven.creeperhost.net": {
			Method:             "HEAD",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusOK,
			ExpectedReponse:    "",
		},
	}
	windowsAppPath = filepath.Join(os.Getenv("localappdata"), "Programs", "ftb-app")
	linuxAppPath   = filepath.Join(os.Getenv("HOME"), ".ftb-app")
	macAppPath     = filepath.Join("/Applications", "FTB Electron App.app")
)
