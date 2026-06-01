alias b := build
alias r := run
alias c := check

bin_name := "bin/radFS"

# Shows this help menu
default:
    @echo "Usage: just <recipe> [arguments...]"
    @echo ""
    @just --list

# Build the Go binary in bin
build:
    mkdir -p bin
    go build -o {{bin_name}} ./cmd/radFS

# Run the filesystem (Usage: just run <folder_name>)
run mnt dbg="":
    #!/usr/bin/env bash
    mkdir -p {{mnt}}

    if [[ -f {{bin_name}} ]]; then
        ./{{bin_name}} {{dbg}} {{mnt}}
    else
        go run ./cmd/radFS {{dbg}} {{mnt}}
    fi

# Clean up bin
clean:
    rm -rf bin
    go clean

# Force unmount if the app crashes (Very helpful for FUSE)
unmount mnt:
    fusermount -u {{mnt}} || umount {{mnt}}

# Format and check Go code logic
check:
    go fmt ./...
    go vet ./...

# Run all tests
test:
    go test -v ./internal/fs/...

# tidy dependencies
tidy:
    go mod tidy
