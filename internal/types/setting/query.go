package setting

import "github.com/zkep/my-geektime/internal/config"

type QueryResponse struct {
	Storage config.Storage `json:"storage" yaml:"storage"`
	Site    config.Site    `json:"site" yaml:"site"`
}
