package models

type SignUpRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
