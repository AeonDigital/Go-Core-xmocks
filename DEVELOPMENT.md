Development Operations Guide
================================

> [Aeon Digital](https://aeondigital.com.br)  
> rianna@aeondigital.com.br

&nbsp;

> Centralized engineering workflow, local automation tooling pipelines, quality gates, and cross-platform environment constraints.




&nbsp;
________________________________________________________________________________

## 1. CLONING & INTERNAL CORE DEVELOPMENT

This repository utilizes the centralized governance engine (`Go-Core-Template-Dev`)
mounted as a native Git Submodule inside the `.dev/` directory to manage local development
automations, verification scripts, and architectural standards.



&nbsp;
---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- 

### 1.1 Workspace Initialization

To initialize the development environment, provision the global governance files,
map the local Git Hooks, and enforce cross-platform compatibility rules, execute
the single automation target at the root level of your project:

```bash
make init-dev
```



&nbsp;
---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- 

### 1.2 Windows Environment Constraints (Strict LF Enforcement)

This ecosystem strictly enforces **Line Feed (LF)** canonical line terminations across
all Go modules to comply with `gofumpt` structural layouts.

Windows environments naturally utilize **Carriage Return Line Feed (CRLF)**. The
`make init-dev` command automatically injects `core.autocrlf false` locally to prevent
file conversion breakages.

Failing to run the initialization target will cause local `pre-commit` validation
gates to block Git commits due to non-compliant CRLF line endings.




&nbsp;
________________________________________________________________________________

## 2. AUTOMATION & MAKEFILE TOOLING

This repository provisions a centralized **Makefile** to encapsulate complex CLI
operations, standardize local validation workflows, and accelerate day-to-day development
cycles.



&nbsp;
---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- 

### 2.1 Discovering Available Targets

To inspect the complete matrix of available automation commands, along with their
down-stream technical descriptions and arguments, execute the help target at the
root level of your project:

```bash
make help
```



&nbsp;
---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- 

### 2.2 Core Automation Categories

The automation targets are divided into three core governance pillars:

- **Workspace Lifecycle (`make init-dev` / `make dev-sync`):** Configures local repository
  environments, maps submodules, forces canonical LF line-ending layouts, and ensures
  compliance with cross-platform formatting rules.
- **Quality Gates & QA (`make lint` / `make fmt-fix`):** Executes deep semantic static
  analysis checks and forces code-base re-writes to align files with rigid vertical
  formatting structures automatically.
- **Development Shortcuts (`make pre-commit` / `make amend`):** Speeds up rapid local
  micro-commit iterations by chaining automatic staging (`git add .`), toolchain
  pruning (`go mod tidy`), commit creations, and forced pushes securely.