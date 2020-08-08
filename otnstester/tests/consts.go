package main

import "github.com/openthread/ot-ns/simulation"

const (
	RoleLeader = "leader"
	RoleRouter = "router"
	RoleChild  = "child"
)

var (
	DefaultRadioRange = simulation.DefaultNodeConfig().RadioRange
)
