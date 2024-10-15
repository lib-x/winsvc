# Windows Service Library for Go

[![GoDoc](https://godoc.org/github.com/lib-x/winsvc?status.svg)](https://godoc.org/github.com/lib-x/winsvc)
[![Go Report Card](https://goreportcard.com/badge/github.com/lib-x/winsvc)](https://goreportcard.com/report/github.com/lib-x/winsvc)

A Go library for creating and managing Windows services with ease.

## Features

- Install, uninstall, start, and stop Windows services
- Run your Go application as a Windows service
- Flexible configuration options for service installation
- Support for both standard and advanced installation methods

## Installation

```bash
go get github.com/lib-x/winsvc
```

## Usage

### Basic Example

Here's a simple example of how to use the library:

```go
package main

import (
	"fmt"
	"log"

	"github.com/lib-x/winsvc"
)

func main() {
	if winsvc.InServiceMode() {
		err := winsvc.RunAsService("MyService", startServer, stopServer, false)
		if err != nil {
			log.Fatalf("Failed to run as service: %v", err)
		}
		return
	}

	// Run as a normal application
	startServer()
}

func startServer() {
	fmt.Println("Server starting...")
	// Your server logic here
}

func stopServer() {
	fmt.Println("Server stopping...")
	// Your cleanup logic here
}
```

### Advanced Installation with Options

You can use the optional installation method for more control over service configuration:

```go
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/lib-x/winsvc"
)

var (
	flagServiceName        = flag.String("name", "MyService", "Service name")
	flagServiceDisplayName = flag.String("display", "My Service Display Name", "Service display name")
	flagServiceDesc        = flag.String("desc", "My service description", "Service description")
	flagServiceInstall     = flag.Bool("install", false, "Install the service")
)

func main() {
	flag.Parse()

	if *flagServiceInstall {
		if err := installService(); err != nil {
			log.Fatalf("Failed to install service: %v", err)
		}
		fmt.Println("Service installed successfully")
		return
	}

	// Other service operations or normal application logic
}

func installService() error {
	exePath, err := winsvc.GetAppPath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	options := []winsvc.ServiceOption{
		winsvc.DisplayName(*flagServiceDisplayName),
		winsvc.Description(*flagServiceDesc),
		winsvc.AutoStart(),
		winsvc.Dependencies("dependency1", "dependency2"),
	}

	return winsvc.InstallServiceWithOption(exePath, *flagServiceName, nil, options...)
}
```

This example demonstrates how to:

- Use command-line flags to configure service properties
- Get the current executable path
- Use various service options like display name, description, start type, and dependencies
- Install the service with custom options

To install the service with this configuration, you would run:

```
yourprogram.exe -install -name "MyCustomService" -display "My Custom Service" -desc "This is a custom Windows service"
```

## API Reference

For detailed API documentation, please refer to the [GoDoc](https://godoc.org/github.com/lib-x/winsvc).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

This project was originally forked from [chai2010/winsvc](https://github.com/chai2010/winsvc). We appreciate their initial work on this library.

## Reporting Issues

If you encounter any bugs or have feature requests, please file an issue on the [GitHub issue tracker](https://github.com/lib-x/winsvc/issues).

## Contact

For any questions or support, please contact <czyt.go@gmail.com>.
