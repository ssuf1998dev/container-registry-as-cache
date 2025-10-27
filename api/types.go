package api

var CRAC_VERSION = 1

type BaseOptions struct {
	Repo     string
	Username string
	Password string
	Insecure bool
}

type Meta struct {
	Version int `json:"created,omitempty"`
}
