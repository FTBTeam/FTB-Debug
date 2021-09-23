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
		Settings		AppSettings
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
