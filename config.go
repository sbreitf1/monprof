package main

type Config struct {
	Profiles []Profile `json:"profiles"`
}

type Profile struct {
	Name       string      `json:"name"`
	Conditions []Condition `json:"conditions"`
	Commands   []string    `json:"cmds"`
}

type Condition struct {
	Monitor string `json:"mon"`
}
