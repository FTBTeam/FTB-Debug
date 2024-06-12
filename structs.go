package main

import (
	"os"
	"os/user"
)

type (
	FTBApp struct {
		User            *user.User
		OWLocation      string
		InstallLocation string
		AppVersion      string
		AppBranch       string
		Released        int
		Structure       AppStructure
		Settings        AppSettings
	}

	AppSettings struct {
		EnableAnalytics  string `json:"enableAnalytics"`
		EnableBeta       string `json:"enableBeta"`
		Memory           string `json:"memory"`
		EnablePreview    string `json:"enablePreview"`
		SessionString    string `json:"sessionString"`
		ListMode         string `json:"listMode"`
		ShowAdverts      string `json:"showAdverts"`
		AutoOpenChat     string `json:"autoOpenChat"`
		PackCardSize     string `json:"packCardSize"`
		ThreadLimit      string `json:"threadLimit"`
		MtConnect        string `json:"mtConnect"`
		Height           string `json:"height"`
		Jvmargs          string `json:"jvmargs"`
		CacheLife        string `json:"cacheLife"`
		CloudSaves       string `json:"cloudSaves"`
		InstanceLocation string `json:"instanceLocation"`
		SpeedLimit       string `json:"speedLimit"`
		EnableChat       string `json:"enableChat"`
		LoadInApp        string `json:"loadInApp"`
		Verbose          string `json:"verbose"`
		AutomateMojang   string `json:"automateMojang"`
		KeepLauncherOpen string `json:"keepLauncherOpen"`
		Width            string `json:"width"`
		ExitOverwolf     string `json:"exitOverwolf"`
		BlockedUsers     string `json:"blockedUsers"`
	}

	Instance struct {
		Private         bool     `json:"_private"`
		UUID            string   `json:"uuid"`
		ID              int      `json:"id"`
		Art             string   `json:"art"`
		Path            string   `json:"path"`
		VersionID       int      `json:"versionId"`
		Name            string   `json:"name"`
		MinMemory       int      `json:"minMemory"`
		RecMemory       int      `json:"recMemory"`
		Memory          int      `json:"memory"`
		Version         string   `json:"version"`
		Dir             string   `json:"dir"`
		Authors         []string `json:"authors"`
		McVersion       string   `json:"mcVersion"`
		JvmArgs         string   `json:"jvmArgs"`
		EmbeddedJre     bool     `json:"embeddedJre"`
		URL             string   `json:"url"`
		ArtURL          string   `json:"artUrl"`
		Width           int      `json:"width"`
		Height          int      `json:"height"`
		ModLoader       string   `json:"modLoader"`
		LastPlayed      int      `json:"lastPlayed"`
		IsModified      bool     `json:"isModified"`
		IsImport        bool     `json:"isImport"`
		CloudSaves      bool     `json:"cloudSaves"`
		HasInstMods     bool     `json:"hasInstMods"`
		InstallComplete bool     `json:"installComplete"`
		PackType        int      `json:"packType"`
	}

	AppStructure struct {
		Bin Bin
	}

	Bin struct {
		Exists bool
		//Profile bool
	}

	FilesToUploadStruct struct {
		File os.FileInfo
		Path string
	}

	CheckURLStruct struct {
		Method             string
		ValidateResponse   bool
		ExpectedStatusCode int
		ExpectedReponse    string
	}

	NetworkCheck struct {
		URL     string
		Success bool
		Error   bool
		Status  string
	}

	// Manifest
	Manifest struct {
		Version                 string               `json:"version,omitempty"`
		MetaDetails             MetaDetails          `json:"metaDetails,omitempty"`
		AppDetails              AppDetails           `json:"appDetails,omitempty"`
		AppLogs                 map[string]string    `json:"appLogs,omitempty"`
		ProviderInstanceMapping map[string]Instances `json:"providerInstanceMapping,omitempty"`
		InstanceLogs            []InstanceLogs       `json:"instanceLogs,omitempty"`
		NetworkChecks           []NetworkCheck       `json:"networkChecks,omitempty"`
	}
	MetaDetails struct {
		InstanceCount     int    `json:"instanceCount,omitempty"`
		Today             string `json:"today,omitempty"`
		Time              int64  `json:"time,omitempty"`
		AddedAccounts     int    `json:"addedAccounts,omitempty"`
		HasActiveAccounts bool   `json:"hasActiveAccounts"`
	}
	AppDetails struct {
		App           string `json:"app,omitempty"`
		SharedVersion string `json:"sharedVersion,omitempty"`
	}
	Instances struct {
		Name        string `json:"name,omitempty"`
		PackType    int    `json:"packType"`
		PackId      int    `json:"packId,omitempty"`
		PackVersion int    `json:"packVersion,omitempty"`
	}
	InstanceLogs struct {
		Created   int64             `json:"created,omitempty"`
		Name      string            `json:"name,omitempty"`
		UUID      string            `json:"uuid,omitempty"`
		McVersion string            `json:"mcVersion,omitempty"`
		ModLoader string            `json:"modLoader,omitempty"`
		Logs      map[string]string `json:"logs,omitempty"`
		CrashLogs map[string]string `json:"crashLogs,omitempty"`
	}

	// Pste.me response
	PsteMeResp struct {
		Data PsteMeData `json:"data"`
		Ok   bool       `json:"ok"`
	}
	PsteMeData struct {
		DeleteID string `json:"delete_id"`
		ID       string `json:"id"`
		Message  string `json:"message"`
	}

	// App meta.json
	AppMeta struct {
		AppVersion string         `json:"appVersion"`
		Commit     string         `json:"commit"`
		Branch     string         `json:"branch"`
		Released   int            `json:"released"`
		Runtime    AppMetaRuntime `json:"runtime"`
	}
	AppMetaJvmArgs struct {
		Value string `json:"value"`
	}
	AppMetaRuntime struct {
		Version string           `json:"version"`
		Jar     string           `json:"jar"`
		Env     []interface{}    `json:"env"`
		JvmArgs []AppMetaJvmArgs `json:"jvmArgs"`
	}

	Profiles struct {
		Version  string `json:"version"`
		Profiles []struct {
			UUID      string `json:"uuid"`
			LastLogin int    `json:"lastLogin"`
			Username  string `json:"username"`
		} `json:"profiles"`
		ActiveProfile string `json:"activeProfile"`
	}
)
