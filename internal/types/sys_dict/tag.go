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

const (
	IsColumnCore  = 1
	IsOpencourse  = 4
	IsColumn      = 5
	IsMentor      = 6
	IsDailylesson = 19
	IsQconp       = 20
)

var (
	ProductTypes = map[string]Option{
		"1": {Label: "体系课", Value: IsColumnCore},
		"4": {Label: "公开课", Value: IsOpencourse},
		"5": {Label: "线下大会", Value: IsColumn},
		"6": {Label: "社区课", Value: IsMentor},
		"d": {Label: "每日一课", Value: IsDailylesson},
		"q": {Label: "大厂案例", Value: IsQconp},
	}

	OriginTypes = map[int32]string{
		IsColumnCore:  "1",
		IsOpencourse:  "4",
		IsColumn:      "5",
		IsMentor:      "6",
		IsDailylesson: "d",
		IsQconp:       "q",
	}

	ProductForms = []Option{
		{Label: "图文+音频", Value: 1},
		{Label: "视频", Value: 2},
	}
)
