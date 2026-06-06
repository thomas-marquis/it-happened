package main

import _ "go.uber.org/mock/gomock"

//go:generate mockgen -package mocksevent -destination mocks/event/bus.go github.com/thomas-marquis/it-happened/event Bus
