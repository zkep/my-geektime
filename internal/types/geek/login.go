package geek

type LoginRequest struct {
	Platform  int     `json:"platform"`
	Appid     int     `json:"appid"`
	Remember  int     `json:"remember"`
	Data      string  `json:"data"`
	Source    string  `json:"source"`
	Ucode     string  `json:"ucode"`
	Sc        LoginSc `json:"sc"`
	Cellphone string  `json:"cellphone"`
	Password  string  `json:"password"`
}

type LoginSc struct {
	UID          string `json:"uid"`
	ReportSource string `json:"report_source"`
	UserUniqueID string `json:"user_unique_id"`
	Refer        string `json:"refer"`
}

type LoginResponse struct {
	Code  int        `json:"code,omitempty"`
	Data  LoginData  `json:"data,omitempty"`
	Error LoginError `json:"error,omitempty"`
	Extra Extra      `json:"extra,omitempty"`
}

type LoginError struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

type Actives struct {
	Athena bool `json:"athena"`
}

type LoginData struct {
	UID               int     `json:"uid"`
	Ucode             string  `json:"ucode"`
	UIDStr            string  `json:"uid_str"`
	Type              int     `json:"type"`
	Cellphone         string  `json:"cellphone"`
	Country           string  `json:"country"`
	Nickname          string  `json:"nickname"`
	Avatar            any     `json:"avatar"`
	Gender            string  `json:"gender"`
	Birthday          string  `json:"birthday"`
	Graduation        string  `json:"graduation"`
	Profession        string  `json:"profession"`
	Industry          string  `json:"industry"`
	Description       string  `json:"description"`
	Overdue           int     `json:"overdue"`
	Cert              int     `json:"cert"`
	Province          string  `json:"province"`
	City              string  `json:"city"`
	Mail              string  `json:"mail"`
	Wechat            string  `json:"wechat"`
	GithubName        string  `json:"github_name"`
	GithubEmail       string  `json:"github_email"`
	Company           string  `json:"company"`
	Post              string  `json:"post"`
	ExpirenceYears    string  `json:"expirence_years"`
	MyPosition        string  `json:"my_position"`
	WorkYear          string  `json:"work_year"`
	DirectionInterest string  `json:"direction_interest"`
	LearnGoal         string  `json:"learn_goal"`
	DepartmentName    string  `json:"department_name"`
	TeamSize          string  `json:"team_size"`
	InternalTraining  string  `json:"internal_training"`
	School            string  `json:"school"`
	RealName          string  `json:"real_name"`
	Openid            string  `json:"openid"`
	Euid              string  `json:"euid"`
	Subtype           int     `json:"subtype"`
	Role              int     `json:"role"`
	Name              string  `json:"name"`
	Address           string  `json:"address"`
	Mobile            string  `json:"mobile"`
	Contact           string  `json:"contact"`
	Position          string  `json:"position"`
	Passworded        bool    `json:"passworded"`
	CreateTime        int     `json:"create_time"`
	JoinInfoq         string  `json:"join_infoq"`
	Actives           Actives `json:"actives"`
	IsStudent         int     `json:"is_student"`
	StudentExpireTime int     `json:"student_expire_time"`
	DeviceID          string  `json:"device_id"`
	Ticket            string  `json:"ticket"`
	TTL               int     `json:"ttl"`
	OssToken          string  `json:"oss_token"`
	IsWhite           int     `json:"is_white"`
}
