package main

import _ "go.uber.org/mock/gomock"

//go:generate mockgen -package mocks_events -destination internal/mocks/events/notifier.go github.com/thomas-marquis/it-happened/event Notifier
