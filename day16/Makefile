all: container

container: main FORCE
	podman build .

main: FORCE
	zig build-exe -O ReleaseFast main.zig

FORCE:
