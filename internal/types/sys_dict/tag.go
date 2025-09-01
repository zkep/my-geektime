package sys_dict

type Tag struct {
	Option
	Options []Option `json:"options"`
}

type Option struct {
	Label string `json:"label"`
	Value int32  `json:"value"`
}

type TagData struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	Data   []Tag  `json:"data,omitempty"`
}

var (
	ProductTypes = map[int32]Option{
		1: {Label: "体系课", Value: 1},
		4: {Label: "公开课", Value: 4},
	}

	ProductForms = []Option{
		{Label: "图文+音频", Value: 1},
		{Label: "视频", Value: 2},
	}
)
