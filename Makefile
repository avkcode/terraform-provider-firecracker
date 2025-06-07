# Make help the default target by placing it first
help:
	@echo "Firecracker Terraform Provider - Development Tools"
	@echo "=================================================="
	@echo "Available commands:"
	@echo ""
	@echo "Development workflow:"
	@echo "  build              - Build the provider and install it locally"
	@echo "  run                - Build, initialize, and apply the Terraform configuration"
	@echo "  test               - Run a basic API test against Firecracker"
	@echo ""
	@echo "Environment management:"
	@echo "  setup              - Set up the complete environment"
	@echo "  teardown           - Tear down the complete environment"
	@echo "  clean              - Clean up temporary files and stop services"
	@echo ""
	@echo "Individual services:"
	@echo "  start-firecracker  - Start the Firecracker VMM"
	@echo "  stop-firecracker   - Stop the Firecracker VMM"
	@echo "  start-socat        - Start socat to forward traffic to Firecracker socket"
	@echo "  stop-socat         - Stop the socat process"
	@echo ""
	@echo "Verification:"
	@echo "  check-terraform    - Verify Terraform installation"
	@echo "  check-files        - Verify required files exist"
	@echo "  check-deps         - Verify all dependencies are installed"

# Add a check for terraform installation
check-terraform:
	@echo "Checking for Terraform installation..."
	@if command -v terraform &> /dev/null; then \
		echo "✅ Terraform found at: $$(which terraform)"; \
	else \
		echo "❌ Error: terraform command not found"; \
		echo "Please install Terraform first:"; \
		echo "  1. Visit https://developer.hashicorp.com/terraform/downloads"; \
		echo "  2. Download and install the appropriate version for your system"; \
		echo "  3. Ensure terraform is in your PATH"; \
		exit 1; \
	fi

# Add a check for required files
check-files:
	@echo "Checking for required files..."
	@mkdir -p test
	@if [ ! -f "test/vmlinux" ]; then \
		echo "❌ Error: Kernel file 'test/vmlinux' not found"; \
		echo "You need to download a Firecracker-compatible kernel."; \
		echo "You can download a sample kernel from:"; \
		echo "  https://github.com/firecracker-microvm/firecracker/blob/main/docs/getting-started.md#running-firecracker"; \
		echo ""; \
		echo "Example commands:"; \
		echo "  mkdir -p test"; \
		echo "  curl -fsSL -o test/vmlinux https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin"; \
		exit 1; \
	fi
	@if [ ! -f "test/firecracker-rootfs.ext4" ]; then \
		echo "❌ Error: Root filesystem 'test/firecracker-rootfs.ext4' not found"; \
		echo "You need to download a Firecracker-compatible root filesystem."; \
		echo "You can download a sample rootfs from:"; \
		echo "  https://github.com/firecracker-microvm/firecracker/blob/main/docs/getting-started.md#running-firecracker"; \
		echo ""; \
		echo "Example commands:"; \
		echo "  mkdir -p test"; \
		echo "  curl -fsSL -o test/firecracker-rootfs.ext4 https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/rootfs/bionic.rootfs.ext4"; \
		exit 1; \
	fi
	@echo "✅ All required files found."


# Add a check for all dependencies
check-deps: check-terraform
	@echo "Checking for required dependencies..."
	@if ! command -v go &> /dev/null; then \
		echo "❌ Error: go command not found"; \
		echo "Please install Go first:"; \
		echo "  Visit https://golang.org/doc/install"; \
		exit 1; \
	fi
	@echo "✅ Go found at: $$(which go)"
	
	@if ! command -v firecracker &> /dev/null; then \
		echo "❌ Error: firecracker command not found"; \
		echo "Please install Firecracker first:"; \
		echo "  Visit https://github.com/firecracker-microvm/firecracker/blob/main/docs/getting-started.md"; \
		exit 1; \
	fi
	@echo "✅ Firecracker found at: $$(which firecracker)"
	
	@if ! command -v socat &> /dev/null; then \
		echo "❌ Error: socat command not found"; \
		echo "Please install socat first:"; \
		echo "  For Ubuntu/Debian: sudo apt-get install socat"; \
		echo "  For CentOS/RHEL: sudo yum install socat"; \
		echo "  For macOS: brew install socat"; \
		exit 1; \
	fi
	@echo "✅ socat found at: $$(which socat)"
	@echo "All dependencies are installed."

# Fix the build target to create directories if they don't exist
build: check-deps
	@echo "Building provider..."
	@go build -o terraform-provider-firecracker
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/hashicorp/firecracker/0.1.0/linux_amd64/
	@cp terraform-provider-firecracker ~/.terraform.d/plugins/registry.terraform.io/hashicorp/firecracker/0.1.0/linux_amd64/
	@echo "✅ Build complete."

# Add dependency tracking to run target
run: build check-terraform check-files setup
	@echo "Running Terraform..."
	@rm -rf test/.terraform.lock.hcl
	@terraform -chdir=test init
	@terraform -chdir=test apply -auto-approve
	@echo "✅ Terraform apply completed successfully."

# Fix the test target to not start multiple Firecracker instances
test: clean stop-firecracker start-firecracker start-socat check-files
	@echo "Testing the /boot-source endpoint..."
	@echo "Sending request to configure boot source..."
	@curl -s -X PUT -H "Content-Type: application/json" \
		-d '{"kernel_image_path":"./test/vmlinux","boot_args":"console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw"}' \
		-w "\nStatus code: %{http_code}\n" \
		http://localhost:8080/boot-source
	@echo "✅ API test completed."

# Fix the start-socat target to not recursively call itself
start-socat:
	@echo "Starting socat to forward traffic from localhost:8080 to /tmp/firecracker.sock..."
	@if pgrep -f "socat TCP-LISTEN:8080" > /dev/null; then \
		echo "⚠️  socat is already running."; \
	else \
		socat TCP-LISTEN:8080,reuseaddr,fork UNIX-CONNECT:/tmp/firecracker.sock & \
		echo "✅ socat started successfully."; \
	fi

stop-socat:
	@echo "Stopping socat..."
	@if pgrep -f "socat TCP-LISTEN:8080" > /dev/null; then \
		pkill -f "socat TCP-LISTEN:8080"; \
		echo "✅ socat stopped."; \
	else \
		echo "⚠️  No socat process found."; \
	fi

clean: stop-socat
	@echo "Cleaning up..."
	@if [ -S /tmp/firecracker.sock ]; then \
		rm -f /tmp/firecracker.sock; \
		echo "Removed /tmp/firecracker.sock"; \
	fi
	@echo "✅ Cleanup complete."

start-firecracker:
	@echo "Starting Firecracker..."
	@if pgrep -f "firecracker --api-sock" > /dev/null; then \
		echo "⚠️  Firecracker is already running."; \
	else \
		firecracker --api-sock /tmp/firecracker.sock & \
		echo "✅ Firecracker started successfully."; \
		sleep 1; \  # Give Firecracker time to initialize
	fi

stop-firecracker:
	@echo "Stopping Firecracker..."
	@if pgrep -f "firecracker --api-sock" > /dev/null; then \
		pkill -f "firecracker --api-sock"; \
		echo "✅ Firecracker stopped."; \
	else \
		echo "⚠️  No Firecracker process found."; \
	fi

# Add a full setup target
setup: clean stop-firecracker start-firecracker start-socat
	@echo "✅ Environment is ready."

# Add a full teardown target
teardown: stop-socat stop-firecracker clean
	@echo "✅ Environment has been torn down."

# Add a status target to check the current state of services
status:
	@echo "Checking environment status..."
	@echo "Firecracker: $$(if pgrep -f 'firecracker --api-sock' > /dev/null; then echo '✅ Running'; else echo '❌ Not running'; fi)"
	@echo "socat: $$(if pgrep -f 'socat TCP-LISTEN:8080' > /dev/null; then echo '✅ Running'; else echo '❌ Not running'; fi)"
	@echo "Socket: $$(if [ -S /tmp/firecracker.sock ]; then echo '✅ Exists'; else echo '❌ Missing'; fi)"

.PHONY: help build run test start-socat stop-socat clean start-firecracker stop-firecracker setup teardown check-terraform check-files check-deps status
