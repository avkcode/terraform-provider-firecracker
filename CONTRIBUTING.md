# Contributing to the Terraform Provider for Firecracker

Thank you for your interest in contributing to the Terraform Provider for Firecracker! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please be respectful and considerate of others when contributing to this project.

## How to Contribute

### Reporting Bugs

If you find a bug, please create an issue on GitHub with the following information:

- A clear, descriptive title
- A detailed description of the issue
- Steps to reproduce the problem
- Expected behavior
- Actual behavior
- Any relevant logs or error messages
- Your environment (Terraform version, Go version, OS, etc.)

### Suggesting Enhancements

We welcome suggestions for enhancements! Please create an issue on GitHub with:

- A clear, descriptive title
- A detailed description of the proposed enhancement
- Any relevant examples or use cases

### Pull Requests

1. Fork the repository
2. Create a new branch for your changes
3. Make your changes
4. Run tests to ensure your changes don't break existing functionality
5. Submit a pull request

#### Pull Request Guidelines

- Follow the existing code style
- Include tests for new functionality
- Update documentation as needed
- Keep pull requests focused on a single change
- Write clear, descriptive commit messages

## Development Environment

### Prerequisites

- Go 1.22 or later
- Terraform 1.0 or later
- Firecracker for testing

### Building the Provider

```bash
make build
```

### Running Tests

```bash
make test
```

## Documentation

If you're adding new features or changing existing ones, please update the documentation accordingly.

## License

By contributing to this project, you agree that your contributions will be licensed under the project's [Apache License 2.0](LICENSE).
