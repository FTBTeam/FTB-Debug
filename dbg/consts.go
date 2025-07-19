package dbg

import (
	"net/http"
	"os"
	"path/filepath"
)

var (
	checkRequestsURLs = map[string]CheckURLStruct{
		"https://api.feed-the-beast.com/v1": {
			Method:             "HEAD",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusNotFound,
			ExpectedReponse:    "",
		},
		"https://meta.feed-the-beast.com/v1/health": {
			Method:             "GET",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusOK,
			ExpectedReponse:    "",
		},
		"https://piston-meta.mojang.com/mc/game/version_manifest_v2.json": {
			Method:             "GET",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusOK,
			ExpectedReponse:    "",
		},
		"https://launchermeta.mojang.com/mc/game/version_manifest_v2.json": {
			Method:             "GET",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusOK,
			ExpectedReponse:    "",
		},
		"https://api.adoptium.net/v3/assets/latest/21/hotspot?architecture=x64&image_type=jre": {
			Method:             "GET",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusOK,
			ExpectedReponse:    "",
		},
		"https://github.com/adoptium/temurin21-binaries/releases/download/jdk-21.0.4%2B7/OpenJDK21U-jre_x64_windows_hotspot_21.0.4_7.zip": {
			Method:             "HEAD",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusOK,
			ExpectedReponse:    "",
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
		"https://maven.creeperhost.net": {
			Method:             "HEAD",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusOK,
			ExpectedReponse:    "",
		},
		"https://login.microsoftonline.com/consumers/oauth2/v2.0/devicecode": {
			Method:             "GET",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedReponse:    "",
		},
		"https://login.microsoftonline.com/consumers/oauth2/v2.0/token": {
			Method:             "POST",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedReponse:    "",
		},
		"https://api.minecraftservices.com/authentication/login_with_xbox": {
			Method:             "POST",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedReponse:    "",
		},
		"https://user.auth.xboxlive.com/user/authenticate": {
			Method:             "POST",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedReponse:    "",
		},
		"https://xsts.auth.xboxlive.com/xsts/authorize": {
			Method:             "POST",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedReponse:    "",
		},
		"https://api.minecraftservices.com/entitlements/license?requestId=RANDOM_UUID": {
			Method:             "GET",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusUnauthorized,
			ExpectedReponse:    "",
		},
		"https://api.minecraftservices.com/entitlements/mcstore": {
			Method:             "GET",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusUnauthorized,
			ExpectedReponse:    "",
		},
		"https://api.minecraftservices.com/minecraft/profile": {
			Method:             "GET",
			ValidateResponse:   false,
			ExpectedStatusCode: http.StatusUnauthorized,
			ExpectedReponse:    "",
		},
	}

	windowsAppPath  = filepath.Join(os.Getenv("localappdata"), "Programs", "ftb-app")
	overwolfAppPath = filepath.Join(os.Getenv("localappdata"), "Overwolf", "Extensions", owUID)
	overwolfAppLogs = filepath.Join(os.Getenv("localappdata"), "Overwolf", "Log", "Apps", "FTB App")

	//linuxAppPath = filepath.Join(os.Getenv("HOME"), ".ftb-app")

	macAppPath = filepath.Join("/Applications", "FTB Electron App.app")
)
