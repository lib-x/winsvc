package winsvc

import (
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

type ServiceOption func(*mgr.Config)

func DisplayName(displayName string) ServiceOption {
	return func(config *mgr.Config) {
		config.DisplayName = displayName
	}
}

func Description(description string) ServiceOption {
	return func(config *mgr.Config) {
		config.Description = description
	}
}

func OnBootStart() ServiceOption {
	return func(config *mgr.Config) {
		config.StartType = windows.SERVICE_BOOT_START
	}
}

func OnSystemStart() ServiceOption {
	return func(config *mgr.Config) {
		config.StartType = windows.SERVICE_SYSTEM_START
	}
}

func AutoStart() ServiceOption {
	return func(config *mgr.Config) {
		config.StartType = windows.SERVICE_AUTO_START
	}
}

func AutoDelayStart() ServiceOption {
	return func(config *mgr.Config) {
		config.StartType = windows.SERVICE_AUTO_START
		config.DelayedAutoStart = true
	}
}

func OnDemandStart() ServiceOption {
	return func(config *mgr.Config) {
		config.StartType = windows.SERVICE_DEMAND_START
	}
}

func DisabledStart() ServiceOption {
	return func(config *mgr.Config) {
		config.StartType = windows.SERVICE_DISABLED
	}
}

func Dependencies(serviceName ...string) ServiceOption {
	return func(config *mgr.Config) {
		for _, svcName := range serviceName {
			config.Dependencies = append(config.Dependencies, svcName)
		}
	}
}
