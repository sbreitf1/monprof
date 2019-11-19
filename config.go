package main

type config struct {
	Profiles []profile `json:"profiles"`
}

type profile struct {
	Name       string      `json:"name"`
	Conditions []condition `json:"conditions"`
	Commands   []string    `json:"cmds"`
}

type condition struct {
	Monitor string `json:"mon"`
}
