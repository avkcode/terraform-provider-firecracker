build:
	@go build -o terraform-provider-firecracker
	@cp terraform-provider-firecracker ~/.terraform.d/plugins/registry.terraform.io/hashicorp/firecracker/0.1.0/linux_amd64/

run: build
	@rm -rf test/.terraform.lock.hcl
	@terraform -chdir=test init
	@terraform -chdir=test apply -auto-approve

test: clean stop-firecracker start-firecracker start-firecracker start-socat
	@echo "Testing the /boot-source endpoint..."; curl -v -X PUT -H "Content-Type: application/json" -d '{"kernel_image_path":"/srv/terraform-provider-firecracker/test/vmlinux","boot_args":"console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw"}' http://localhost:8080

start-socat: clean stop-firecracker start-firecracker start-firecracker start-socat
	@echo "Starting socat to forward traffic from localhost:8080 to /tmp/firecracker.sock..."
	@socat TCP-LISTEN:8080,reuseaddr,fork UNIX-CONNECT:/tmp/firecracker.sock &
	@echo "socat started successfully."

stop-socat:
	@echo "Stopping socat..."
	-@pkill -f "socat TCP-LISTEN:8080"
	@echo "socat stopped."

clean: stop-socat
	@echo "Cleaning up..."
	@rm -f /tmp/firecracker.sock
	@echo "Cleanup complete."

start-firecracker:
	@echo "Starting Firecracker..."
	@firecracker --api-sock /tmp/firecracker.sock &
	@echo "Firecracker started successfully."

stop-firecracker:
	@echo "Stopping Firecracker..."
	-@pkill -f "firecracker --api-sock"
	@echo "Firecracker stopped."

.PHONY: all start-socat stop-socat clean start-firecracker stop-firecracker
