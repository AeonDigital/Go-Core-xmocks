Go-Core-xmocks
================================

![Go Test Coverage](https://raw.githubusercontent.com/AeonDigital/Go-Core-xmocks/badges/.badges/main/coverage.svg)

> [Aeon Digital](http://aeondigital.com.br)  
> rianna@aeondigital.com.br

&nbsp;

> A type-safe CLI mock generator for Go that parses interfaces via AST to streamline structural dependency decoupling.

`xmocks` is a generic command-line tool designed to inspect Go interfaces and automatically
generate type-safe mock implementations. It is purpose-built to support the architectural
requirements and strict decoupling patterns throughout the `Go-Core` ecosystem.



### How It Works: The Power of AST

Unlike primitive mock generators that rely on text scanning or complex runtime reflection,
`xmocks` operates directly on your source code using an **Abstract Syntax Tree (AST)**.

An AST is a highly structured, tree-like representation of source code architecture
parsed natively by the Go compiler toolchain (`go/ast`). By leveraging AST parsing,
`xmocks` naturally "understands" Go syntax semantics rather than just viewing your
files as raw text. This guarantees absolute precision when resolving custom types,
local identifiers, and method signatures, ensuring that the emitted mock files are
perfectly formed and compile flawlessly every single time.



### Key Architectural Benefits

- **Dynamic Package Resolution**: Automatically detects the parent source package
  and target folder names to dynamically stamps the correct headers into your generated
  artifacts (e.g., `package mocks`).
- **Automatic Package Prefixing**: Locally declared types and custom structs automatically
  receive their parent package qualifiers (e.g., `xbridge.IFileBridge`), preventing
  import fragmentation and type errors.
- **Variadic Parameter Support**: Offers native handling for variadic parameters
  (`...string`), ensuring that the underlying slice expansion fields (`args...`)
  propagate seamlessly into mock function signatures.
- **Collision-Resistant Pointer Receivers**: All generated mock methods consistently
  utilize `oMock` as their internal pointer receiver name, eliminating naming collisions
  with incoming method parameters (e.g., `m *runtime.MemStats`).




&nbsp;
________________________________________________________________________________

## 1. INSTALLATION

This section describes how to fetch, build, and use this module within your local
workspace or run it across downstream development environments.



### 1.1. Prerequisites

Before installing, ensure that your system meets the following software requirements:

- **Go Environment**: Go 1.21 or higher installed and configured in your system path.
- **Go Modules**: Your target application project must be initialized with a valid
  `go.mod` file.



### 1.2. Global CLI Installation

To install `xmocks` as a globally available binary executable on your development
machine, execute the standard `go install` command pointing directly to the command
subdirectory:

```bash
go install github.com/AeonDigital/Go-Core-xmocks/cmd/xmocks@latest
```

Make sure your `$GOPATH/bin` directory is added to your system's `PATH` environment
variable to execute `xmocks` from anywhere.



### 1.3. Local Workspace Execution

If you prefer to run the generator dynamically without a global installation, or
if you are modifying the generator's source code, you can invoke it directly from
the root of this repository:

```bash
go run ./cmd/xmocks --file=<path/to/source.go> --interface=<InterfaceName>
```




&nbsp;
________________________________________________________________________________

## 2. ARCHITECTURAL CONVENTIONS & CONFIGURATION

This section details the CLI configuration parameters and the architectural constraints
enforced by the `Go-Core` ecocystem that this generator expects.



### 2.1. CLI Flags

The `xmocks` CLI engine is strictly guided by command flags. It does not require
external configuration files or environment variables.


#### Required Flags

- `--file` — The relative or absolute path to the target Go source file containing
  the interface definition.
- `--interface` — The exact case-sensitive name of the target interface to mock (e.g.,
  `IOSBridge`).


#### Optional Flags

- `--alias` — A custom naming alias used as a prefix suffix for the generated structures.
  If omitted, it defaults to the interface name.
- `--output` — Target output directory or direct file path. If pointing to a directory,
  the file will be dynamically named based on the lowercased alias (e.g., `<alias>.go`).
  If omitted, it defaults to `internal/pkgxmock/<alias>.go`.



### 2.2. Architectural Conventions (The Bridge Pattern)

To achieve strict platform isolation and guarantee 100% unit-testability across infrastructure
layers (such as `os`, `io`, and `runtime`), the `Go-Core` layout enforces the **Bridge
Pattern**:

1. **Unified Naming Design**: Public interfaces must be prefixed with a capital `I`
   (e.g., `IOSBridge`). Concrete implementations must be unexported, private structs
   prefixed with a lowercase `s` (e.g., `sOSBridge`). Standard production access
   is funneled through a single public global pointer variable named after the package
   (e.g., `var OSBridge IOSBridge = sOSBridge{}`).

2. **Visibilities & Boundaries**: Consumer code outside the core package communicates
   exclusively with the global variable gateway. The backing private structs remain
   fully hidden.

3. **Compile-Time Safeguards**: To ensure strict interface compliance, a compile-time
   assertion check must exist right below the private struct declaration:


   ```go
   var _ IOSBridge = sOSBridge{}
   ```


4. **Dynamic Lifecycles**: Heavy infrastructure operations remain anchored behind
   global singletons. Volatile resources (like an open file pointer `IFileBridge`)
   **do not possess global entry points**; they are manufactured dynamically by main
   services and injected during test cycles.



### 2.3. Testing Contract & Runtime Protection

The generated mock artifacts follow a strict structural contract to ensure type safety
and explicit expectations:

- **`TestCase<Alias>` Structure**: Dynamically emits a standard test case descriptor
  mapping out inputs, expected results (`Want`, `WantErr`), and an inline environment
  configuration runner (`MockFn`).

- **`OnCall` & `SetReturn` Mechanics**: Generates explicit hook interfaces that allow
  tests to override behaviors safely using type-safe method signatures.

- **Runtime Panic Protection (`panicIfNotConfigured`)**: If your production code
  interacts with a mock method whose behavior has not been explicitly declared or
  stubbed in the active test runner, `xmocks` intercepts the unconfigured call and
  triggers a diagnostic `panic()`, indicating exactly which method breached the test
  configuration contract.




&nbsp;
________________________________________________________________________________

## 3. BASIC USAGE

This section provides actionable command blueprints, architecture batch execution
workflows, and a reference structural example of how to implement the generated mocks
inside your test battery.



### 3.1. Single Code Generation Example

To generate a type-safe mock implementation for a single targeted interface, run
the executable command specifying the input path and destination:

```bash
xmocks --file=xbridge/go/bos/os.go --interface=IOSBridge --alias=OS --output=xbridge/go/bos/mocks/
```



### 3.2. Architecture Batch Generation (Go-Core Pattern)

To batch-process and rebuild the complete core `xbridge` mocking architecture layer
inside your workspace layout, execute the following commands in sequence:

```bash
# 1. Operating System Layer (bos)
go run ./cmd/xmocks --file=xbridge/go/bos/os.go --interface=IOSBridge --alias=OS --output=xbridge/go/bos/mocks/

# 2. Path and File Layout Layer (bfilepath)
go run ./cmd/xmocks --file=xbridge/go/bpath/bfilepath/filepath.go --interface=IFilepathBridge --alias=Filepath --output=xbridge/go/bpath/bfilepath/mocks/

# 3. Environment and Diagnostics Layer (bruntime)
go run ./cmd/xmocks --file=xbridge/go/bruntime/runtime.go --interface=IRuntimeBridge --alias=Runtime --output=xbridge/go/bruntime/mocks/

# 4. Data Streams Layer (bio)
go run ./cmd/xmocks --file=xbridge/go/bio/io.go --interface=IIOBridge --alias=IO --output=xbridge/go/bio/mocks/

# 5. Physical File System Processing Layers (bio/bfs)
go run ./cmd/xmocks --file=xbridge/go/bio/bfs/file.go --interface=IFileBridge --alias=File --output=xbridge/go/bio/bfs/mocks/
go run ./cmd/xmocks --file=xbridge/go/bio/bfs/fileinfo.go --interface=IFileInfoBridge --alias=FileInfo --output=xbridge/go/bio/bfs/mocks/
```



This populates a clean, fully independent subpackage layout under your target boundaries
stamped with `package mocks`, ready for consumption.



### 3.3. How to Use the Generated Mocks in Tests

When `xmocks` runs, it outputs a complete testing infrastructure layout comprising
a `Mock<Alias>` controller, alongside dedicated `OnCall`, `SetReturn`, and `TestCase<Alias>`
types. Below is an example blueprint demonstrating how to leverage these generated
contracts inside standard Go tests:

```go
package mypackage_test

import (
	"testing"
	
	// Import your generated package path layout here
	// "github com/AeonDigital/Go-Core-xmocks/xbridge/go/bos/mocks"
)

func TestServiceOperation(t *testing.T) {
	// 1. Define the test case matrix utilizing the emitted TestCase structure
	tests := []mocks.TestCaseOS{
		{
			Name: "Success when reading environment settings",
			MockFn: func(m *mocks.MockOS) {
				// Use type-safe SetReturn to configure explicit values
				m.SetReturn.LookupEnv("APP_ENV", "production", true)
			},
			Want:    "production",
			WantErr: false,
		},
		{
			Name: "Failure when fallback logic is intercepted",
			MockFn: func(m *mocks.MockOS) {
				// Use OnCall to hook highly detailed custom execution logic
				m.OnCall.LookupEnv(func(key string) (string, bool) {
					if key != "REQUIRED_KEY" {
						t.Errorf("unexpected lookup key: %s", key)
					}
					return "", false
				})
			},
			Want:    "",
			WantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// 2. Initialize the type-safe mock controller instance
			mockOS := mocks.NewMockOS()

			// 3. Bind the test case environment parameters
			if tt.MockFn != nil {
				tt.MockFn(mockOS)
			}

			// [Execute your service under test passing mockOS as dependency]
			// [Assert outcomes against tt.Want or tt.WantErr boundaries]
		})
	}
}
```




### 3.4. Running Project Internal Tests

To execute the generator's internal test engine battery and review overall codebase
statements coverage metrics, issue the following standard verification script command:

```bash
go test -cover -v ./cmd/xmocks
```




&nbsp;
________________________________________________________________________________

## 4. LICENCE

This project is offered under the [MIT license](LICENSE.md).