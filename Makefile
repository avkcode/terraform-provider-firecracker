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
	@echo "Advanced testing:"
	@echo "  test-remote-exec   - Create a test configuration using remote-exec provisioner"
	@echo "  setup-network      - Set up networking for remote-exec tests"
	@echo "  prepare-ssh-image  - Instructions for preparing a VM image with SSH enabled"
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
	@echo "  status             - Check the status of all services"

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
	@if command -v go &> /dev/null; then \
		echo "✅ Go found at: $$(which go)"; \
	elif [ -x /usr/local/go/bin/go ]; then \
		echo "✅ Go found at: /usr/local/go/bin/go"; \
		echo "⚠️  Warning: Go is not in your PATH. Consider adding /usr/local/go/bin to your PATH."; \
		export PATH=$$PATH:/usr/local/go/bin; \
	else \
		echo "❌ Error: go command not found"; \
		echo "Please install Go first:"; \
		echo "  Visit https://golang.org/doc/install"; \
		exit 1; \
	fi
	
	@if command -v firecracker &> /dev/null; then \
		echo "✅ Firecracker found at: $$(which firecracker)"; \
	elif [ -x /usr/local/bin/firecracker ]; then \
		echo "✅ Firecracker found at: /usr/local/bin/firecracker"; \
		echo "⚠️  Warning: Firecracker is not in your PATH. Consider adding /usr/local/bin to your PATH."; \
		export PATH=$$PATH:/usr/local/bin; \
	else \
		echo "❌ Error: firecracker command not found"; \
		echo "Please install Firecracker first:"; \
		echo "  Visit https://github.com/firecracker-microvm/firecracker/blob/main/docs/getting-started.md"; \
		exit 1; \
	fi
	
	@if command -v socat &> /dev/null; then \
		echo "✅ socat found at: $$(which socat)"; \
	elif [ -x /usr/local/bin/socat ]; then \
		echo "✅ socat found at: /usr/local/bin/socat"; \
		echo "⚠️  Warning: socat is not in your PATH. Consider adding /usr/local/bin to your PATH."; \
		export PATH=$$PATH:/usr/local/bin; \
	else \
		echo "❌ Error: socat command not found"; \
		echo "Please install socat first:"; \
		echo "  For Ubuntu/Debian: sudo apt-get install socat"; \
		echo "  For CentOS/RHEL: sudo yum install socat"; \
		echo "  For macOS: brew install socat"; \
		exit 1; \
	fi
	@echo "✅ All dependencies are installed."

# Fix the build target to create directories if they don't exist
build: check-deps
	@echo "Building provider..."
	@if command -v go &> /dev/null; then \
		go build -o terraform-provider-firecracker; \
	else \
		/usr/local/go/bin/go build -o terraform-provider-firecracker; \
	fi
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/hashicorp/firecracker/0.1.0/linux_amd64/
	@cp terraform-provider-firecracker ~/.terraform.d/plugins/registry.terraform.io/hashicorp/firecracker/0.1.0/linux_amd64/
	@echo "✅ Build complete."

# Add dependency tracking to run target
run: build check-terraform check-files setup
	@echo "Running Terraform..."
	@rm -rf test/.terraform.lock.hcl
	@rm -f test/remote-exec-template.tf  # Remove this file if it exists
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
		echo "Found socat process, stopping it..."; \
		pkill -TERM -f "socat TCP-LISTEN:8080" || true; \
		sleep 1; \
		if pgrep -f "socat TCP-LISTEN:8080" > /dev/null; then \
			echo "Socat still running, forcing kill..."; \
			pkill -KILL -f "socat TCP-LISTEN:8080" || true; \
		fi; \
		echo "✅ socat stopped."; \
	else \
		echo "⚠️  No socat process found."; \
	fi

clean:
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
		sleep 1; \
	fi

stop-firecracker:
	@echo "Stopping Firecracker..."
	@if pgrep -f "firecracker --api-sock" > /dev/null; then \
		pkill -f "firecracker --api-sock" || true; \
		echo "✅ Firecracker stopped."; \
	else \
		echo "⚠️  No Firecracker process found."; \
	fi

# Add a full setup target
setup:
	@echo "Setting up environment..."
	@$(MAKE) stop-socat || true
	@$(MAKE) clean
	@$(MAKE) stop-firecracker || true
	@$(MAKE) start-firecracker
	@$(MAKE) start-socat
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

# Add a remote-exec test target
test-remote-exec: build check-terraform check-files setup
	@echo "Testing remote-exec provisioner to install Docker..."
	@mkdir -p test/remote-exec
	@mkdir -p templates
	@cp templates/remote-exec-template.tf test/remote-exec/main.tf
	@echo "✅ Created remote-exec test configuration"
	@echo "Generating SSH key for remote access..."
	@mkdir -p test
	@if [ ! -f "test/id_rsa" ]; then \
		ssh-keygen -t rsa -b 2048 -f test/id_rsa -N ""; \
		echo "✅ SSH key generated"; \
	else \
		echo "⚠️ SSH key already exists"; \
	fi
	@echo "To run this test, you need to:"
	@echo "1. Ensure your Firecracker VM has SSH enabled"
	@echo "2. Configure networking with the tap0 interface"
	@echo "3. Add the public key to the VM's authorized_keys (use prepare-ssh-image target)"
	@echo "4. Run: terraform -chdir=test/remote-exec init && terraform -chdir=test/remote-exec apply"
	@echo ""
	@echo "Note: This test requires additional setup and is not fully automated."
	@echo "For more information, see the documentation on setting up networking and SSH access."

# Add a setup-network target to configure tap interfaces
setup-network:
	@echo "Setting up network for remote-exec test..."
	@if ! ip link show tap0 &> /dev/null; then \
		echo "Creating tap0 interface..."; \
		ip tuntap add dev tap0 mode tap; \
		ip addr add 172.16.0.1/24 dev tap0; \
		ip link set tap0 up; \
		echo "Enabling IP forwarding..."; \
		echo 1 > /proc/sys/net/ipv4/ip_forward; \
		echo "Setting up NAT..."; \
		iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE; \
		iptables -A FORWARD -i eth0 -o tap0 -m state --state RELATED,ESTABLISHED -j ACCEPT; \
		iptables -A FORWARD -i tap0 -o eth0 -j ACCEPT; \
		echo "✅ Network setup complete"; \
	else \
		echo "⚠️ tap0 interface already exists"; \
	fi

# Add a prepare-ssh-image target to create a VM image with SSH enabled
prepare-ssh-image: check-files
	@echo "Preparing VM image with SSH enabled..."
	@if [ ! -f "test/id_rsa" ]; then \
		echo "Generating SSH key..."; \
		ssh-keygen -t rsa -b 2048 -f test/id_rsa -N ""; \
	fi
	@echo "This step requires manual intervention to:"
	@echo "1. Mount the rootfs image"
	@echo "2. Install SSH server"
	@echo "3. Configure SSH to allow root login"
	@echo "4. Add the public key to authorized_keys"
	@echo ""
	@echo "Example commands:"
	@echo "  mkdir -p /mnt/rootfs"
	@echo "  mount -o loop test/firecracker-rootfs.ext4 /mnt/rootfs"
	@echo "  chroot /mnt/rootfs apt-get update"
	@echo "  chroot /mnt/rootfs apt-get install -y openssh-server"
	@echo "  mkdir -p /mnt/rootfs/root/.ssh"
	@echo "  cat test/id_rsa.pub > /mnt/rootfs/root/.ssh/authorized_keys"
	@echo "  chown -R root:root /mnt/rootfs/root/.ssh"
	@echo "  chmod 600 /mnt/rootfs/root/.ssh/authorized_keys"
	@echo "  echo 'PermitRootLogin yes' >> /mnt/rootfs/etc/ssh/sshd_config"
	@echo "  umount /mnt/rootfs"
	@echo ""
	@echo "⚠️ Note: This is a manual process and requires root privileges."

.PHONY: help build run test start-socat stop-socat clean start-firecracker stop-firecracker setup teardown check-terraform check-files check-deps status test-remote-exec setup-network prepare-ssh-image
