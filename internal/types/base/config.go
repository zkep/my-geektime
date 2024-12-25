package base

type Config struct {
	RegisterType string `json:"register_type"`
	LoginType    string `json:"login_type"`
	LoginGuest   Guest  `json:"login_guest"`
}

type Guest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}
