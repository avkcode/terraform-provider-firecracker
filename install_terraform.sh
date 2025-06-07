#!/bin/bash

# Script to install Terraform

# Set Terraform version
TERRAFORM_VERSION="1.7.5"

echo "Installing Terraform version ${TERRAFORM_VERSION}..."

# Create a temporary directory
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

# Download Terraform
echo "Downloading Terraform..."
curl -s -O "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip"

# Install unzip if not available
if ! command -v unzip &> /dev/null; then
    echo "Installing unzip..."
    sudo apt-get update
    sudo apt-get install -y unzip
fi

# Extract the zip file
echo "Extracting Terraform..."
unzip "terraform_${TERRAFORM_VERSION}_linux_amd64.zip"

# Move the executable to a directory in the PATH
echo "Installing Terraform binary..."
sudo mv terraform /usr/local/bin/

# Clean up
cd
rm -rf "$TMP_DIR"

# Verify installation
echo "Verifying installation..."
terraform --version

echo "Terraform installation complete!"
