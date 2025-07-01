package setting

type SettingUpdate struct {
	StorageHost   string   `json:"storageHost,omitempty"`
	SiteProxyURL  string   `json:"siteProxyUrl,omitempty"`
	SiteDownload  bool     `json:"siteDownload,omitempty"`
	SiteProxyUrls []string `json:"siteProxyUrls,omitempty"`
	SitePlayUrls  []string `json:"sitePlayUrls,omitempty"`
}
