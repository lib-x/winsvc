// Package winsvc provides utilities for creating and managing Windows services.
package winsvc

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

// GetAppPath returns the absolute path of the current executable.
func GetAppPath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		return "", fmt.Errorf("GetAppPath: %s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			return "", fmt.Errorf("GetAppPath: %s is directory", p)
		}
	}
	return "", err
}

// InServiceMode returns true if the current process is running as a Windows service.
func InServiceMode() bool {
	isService, err := svc.IsWindowsService()
	if err != nil {
		return false
	}
	return isService
}

// IsAnInteractiveSession returns true if the current process is running in an interactive session.
func IsAnInteractiveSession() bool {
	return !InServiceMode()
}

// InstallService installs a Windows service with the given parameters.
// It takes the application path, service name, display name, description, and optional parameters.
func InstallService(appPath, name, displayName, desc string, params ...string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", name)
	}

	s, err = m.CreateService(name, appPath, mgr.Config{
		DisplayName: displayName,
		Description: desc,
		StartType:   windows.SERVICE_AUTO_START,
	}, params...)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer s.Close()

	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("failed to install event logger: %w", err)
	}

	return nil
}

// InstallServiceWithOption installs a Windows service with custom options.
// It takes the application path, service name, a ServiceArgsOption function, and variadic ServiceOption functions.
func InstallServiceWithOption(appPath, name string, serviceArgs []string, options ...ServiceOption) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", name)
	}

	config := mgr.Config{
		StartType: mgr.StartAutomatic,
	}

	// Apply all provided options
	for _, option := range options {
		option(&config)
	}

	s, err = m.CreateService(name, appPath, config, serviceArgs...)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer s.Close()

	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("failed to install event logger: %w", err)
	}

	return nil
}

// RemoveService removes a Windows service with the given name.
func RemoveService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", name)
	}
	defer s.Close()

	err = s.Delete()
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	err = eventlog.Remove(name)
	if err != nil {
		return fmt.Errorf("failed to remove event logger: %w", err)
	}

	return nil
}

// StartService starts a Windows service with the given name.
func StartService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not access service: %w", err)
	}
	defer s.Close()

	err = s.Start("is", "manual-started")
	if err != nil {
		return fmt.Errorf("could not start service: %w", err)
	}

	return nil
}

// StopService stops a Windows service with the given name.
func StopService(name string) error {
	return controlService(name, svc.Stop, svc.Stopped)
}

// QueryService returns the current status of a Windows service.
func QueryService(name string) (string, error) {
	m, err := mgr.Connect()
	if err != nil {
		return "", fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return "", fmt.Errorf("could not access service: %w", err)
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		return "", fmt.Errorf("could not query service status: %w", err)
	}

	switch status.State {
	case svc.Stopped:
		return "Stopped", nil
	case svc.StartPending:
		return "StartPending", nil
	case svc.StopPending:
		return "StopPending", nil
	case svc.Running:
		return "Running", nil
	case svc.ContinuePending:
		return "ContinuePending", nil
	case svc.PausePending:
		return "PausePending", nil
	case svc.Paused:
		return "Paused", nil
	default:
		return "", fmt.Errorf("unknown service state")
	}
}

func controlService(name string, c svc.Cmd, to svc.State) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not access service: %w", err)
	}
	defer s.Close()

	status, err := s.Control(c)
	if err != nil {
		return fmt.Errorf("could not send control=%d: %w", c, err)
	}

	timeout := time.Now().Add(10 * time.Second)
	for status.State != to {
		if timeout.Before(time.Now()) {
			return fmt.Errorf("timeout waiting for service to go to state=%d", to)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("could not retrieve service status: %w", err)
		}
	}

	return nil
}

var elog debug.Log

// RunAsService runs the provided start and stop functions as a Windows service.
// It takes the service name, start function, stop function, and a debug flag.
func RunAsService(name string, start, stop func(), isDebug bool) error {
	var err error
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return fmt.Errorf("failed to open event log: %w", err)
		}
	}
	defer elog.Close()

	run := svc.Run
	if isDebug {
		run = debug.Run
	}

	elog.Info(1, fmt.Sprintf("starting %s service", name))
	err = run(name, &winService{start: start, stop: stop})
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return fmt.Errorf("service run failed: %w", err)
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", name))
	return nil
}

type winService struct {
	start func()
	stop  func()
}

func (s *winService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	go s.start()

	for c := range r {
		switch c.Cmd {
		case svc.Interrogate:
			changes <- c.CurrentStatus
			time.Sleep(100 * time.Millisecond)
			changes <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			changes <- svc.Status{State: svc.StopPending}
			s.stop()
			return false, 0
		case svc.Pause:
			changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
		case svc.Continue:
			changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
		default:
			elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
		}
	}

	return false, 0
}
