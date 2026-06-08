package main

import _ "go.uber.org/mock/gomock"

//go:generate mockgen -package mocksevent -destination internal/mocks/event/bus.go github.com/thomas-marquis/it-happened/event Bus
//go:generate mockgen -package mockruntime -destination internal/mocks/runtime/clock.go github.com/thomas-marquis/it-happened/eventest/internal/runtime Clock
