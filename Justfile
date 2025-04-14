# -*- Justfile -*-

set dotenv-filename := "config/.env.dev"

app_name := "mouser"
coverage_file := "coverage.out"

# List the available justfile recipes.
[group('general')]
@default:
  just --list --unsorted

# List the lines of code in the project.
[group('general')]
loc:
  scc --remap-unknown "-*- Justfile -*-":"justfile"

# List the outdated direct dependencies (can be slow).
[group('dependencies')]
outdated:
  # Requires https://github.com/psampaz/go-mod-outdated
  go list -u -m -json all | go-mod-outdated -update -direct

# Run go mod tidy and verify.
[group('dependencies')]
tidy:
  go mod tidy
  go mod verify

# Format and vet Go code. Runs before tests.
[group('test')]
check:
	go fmt ./...
	go vet ./...

# Lint code using staticcheck.
[group('test')]
lint: check
	staticcheck -f stylish ./...

# Run the unit tests.
[group('test')]
unit *FLAGS: check
  go test ./... -cover -vet=off -race {{FLAGS}} -short

# Run the integration tests.
[group('test')]
int *FLAGS: check
  go test ./... -cover -vet=off -race {{FLAGS}} -run Integration

# HTML report for unit (default), int, e2e, or all tests.
[group('test')]
cover test='unit': check
  go test ./... -vet=off -coverprofile={{coverage_file}} \
  {{ if test == 'all' { '' } \
    else if test == 'int' { '-run Integration' } \
    else if test == 'e2e' { '-run E2E' } \
    else { '-short' } }}
  go tool cover -html={{coverage_file}}

# Build and run example Mouser CLI application.
[group('build')]
example:
  cd ./cmd/example/ && env go build -o ../../dist/{{app_name}}
  dist/{{app_name}}
