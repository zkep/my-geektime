package sys_dict

import "fmt"

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

func ProductURLWithType(productType string, productID int) string {
	redirect := ""
	switch productType {
	case "c1", "x49", "x50":
		redirect = fmt.Sprintf("https://time.geekbang.org/column/intro/%d", productID)
	case "c3", "c6":
		redirect = fmt.Sprintf("https://time.geekbang.org/course/intro/%d", productID)
	case "p29":
		redirect = fmt.Sprintf("https://time.geekbang.org/opencourse/intro/%d", productID)
	case "p30", "p35":
		redirect = fmt.Sprintf("https://time.geekbang.org/opencourse/videointro/%d", productID)
	case "d":
		redirect = fmt.Sprintf("https://time.geekbang.org/dailylesson/detail/%d", productID)
	case "q":
		redirect = fmt.Sprintf("https://time.geekbang.org/qconplus/detail/%d", productID)
	default:
	}
	return redirect
}

func ProductDetailURLWithType(productType string, productID, articleID int) string {
	redirect := fmt.Sprintf("https://time.geekbang.org/course/detail/%d-%d", productID, articleID)
	switch productType {
	case "d":
		redirect = fmt.Sprintf("https://time.geekbang.org/dailylesson/detail/%d", productID)
	case "q":
		redirect = fmt.Sprintf("https://time.geekbang.org/qconplus/detail/%d", productID)
	default:
	}
	return redirect
}
