# Add a check for terraform installation
check-terraform:
	@if command -v terraform &> /dev/null; then \
		echo "Terraform found at: $$(which terraform)"; \
	else \
		echo "Error: terraform command not found"; \
		echo "Please install Terraform first:"; \
		echo "  1. Visit https://developer.hashicorp.com/terraform/downloads"; \
		echo "  2. Download and install the appropriate version for your system"; \
		echo "  3. Ensure terraform is in your PATH"; \
		exit 1; \
	fi

# Add a default target that shows available commands
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  build              - Build the provider and install it locally"
	@echo "  run                - Build, initialize, and apply the Terraform configuration"
	@echo "  test               - Run a basic API test against Firecracker"
	@echo "  start-socat        - Start socat to forward traffic to Firecracker socket"
	@echo "  stop-socat         - Stop the socat process"
	@echo "  start-firecracker  - Start the Firecracker VMM"
	@echo "  stop-firecracker   - Stop the Firecracker VMM"
	@echo "  clean              - Clean up temporary files and stop services"
	@echo "  setup              - Set up the complete environment"
	@echo "  teardown           - Tear down the complete environment"

# Fix the build target to create directories if they don't exist
build:
	@echo "Building provider..."
	@go build -o terraform-provider-firecracker
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/hashicorp/firecracker/0.1.0/linux_amd64/
	@cp terraform-provider-firecracker ~/.terraform.d/plugins/registry.terraform.io/hashicorp/firecracker/0.1.0/linux_amd64/
	@echo "Build complete."

# Add dependency tracking to run target
run: build check-terraform
	@echo "Running Terraform..."
	@rm -rf test/.terraform.lock.hcl
	@terraform -chdir=test init
	@terraform -chdir=test apply -auto-approve

# Fix the test target to not start multiple Firecracker instances
test: clean stop-firecracker start-firecracker start-socat
	@echo "Testing the /boot-source endpoint..."
	@curl -v -X PUT -H "Content-Type: application/json" \
		-d '{"kernel_image_path":"/srv/terraform-provider-firecracker/test/vmlinux","boot_args":"console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw"}' \
		http://localhost:8080

# Fix the start-socat target to not recursively call itself
start-socat:
	@echo "Starting socat to forward traffic from localhost:8080 to /tmp/firecracker.sock..."
	@socat TCP-LISTEN:8080,reuseaddr,fork UNIX-CONNECT:/tmp/firecracker.sock &
	@echo "socat started successfully."

stop-socat:
	@echo "Stopping socat..."
	-@pkill -f "socat TCP-LISTEN:8080" || true
	@echo "socat stopped."

clean: stop-socat
	@echo "Cleaning up..."
	@rm -f /tmp/firecracker.sock
	@echo "Cleanup complete."

start-firecracker:
	@echo "Starting Firecracker..."
	@firecracker --api-sock /tmp/firecracker.sock &
	@echo "Firecracker started successfully."
	@sleep 1  # Give Firecracker time to initialize

stop-firecracker:
	@echo "Stopping Firecracker..."
	-@pkill -f "firecracker --api-sock" || true
	@echo "Firecracker stopped."

# Add a full setup target
setup: clean stop-firecracker start-firecracker start-socat
	@echo "Environment is ready."

# Add a full teardown target
teardown: stop-socat stop-firecracker clean
	@echo "Environment has been torn down."

.PHONY: all help build run test start-socat stop-socat clean start-firecracker stop-firecracker setup teardown check-terraform
