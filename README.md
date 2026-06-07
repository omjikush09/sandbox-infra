# Sandboxing Infra

Sandboxing Infra is an experimental infrastructure project for running isolated workloads with Firecracker microVMs. The repository currently combines Go packages for interacting with Firecracker and Terraform configuration for provisioning an AWS EC2 host with nested virtualization enabled.

> This project is still in development. APIs, package layout, infrastructure settings, and setup steps are expected to change.

## Project Goals

- Provision Linux hosts suitable for Firecracker-based sandboxing.
- Start and configure Firecracker microVMs from Go.
- Manage VM networking with TAP devices, IP forwarding, and NAT.
- Build toward orchestration components for sandbox lifecycle management.

## Repository Layout

```text
.
├── main.go                         # Root Go entry point placeholder
├── go.work                         # Go workspace
├── demon/                          # Daemon placeholder package
├── packages/
│   ├── vm/                         # Firecracker client and VM startup code
│   │   ├── client/                 # HTTP client for Firecracker Unix socket API
│   │   └── start/                  # VM startup and network setup helpers
│   ├── envd/                       # Environment daemon placeholder
│   └── orchestator/                # Orchestrator placeholder
└── infra/
    ├── main.tf                     # Terraform AWS provider and EC2 module wiring
    └── modules/ec2/                # EC2 host, security group, and bootstrap scripts
```

## Current Components

### Firecracker VM package

The `packages/vm` module contains early Go code for:

- Creating an HTTP client that talks to the Firecracker API over a Unix socket.
- Sending Firecracker API requests.
- Starting the `firecracker` process with an API socket.
- Setting up TAP networking and host NAT rules for guest connectivity.

### Infrastructure

The Terraform configuration under `infra/` provisions an AWS EC2 instance using an Ubuntu 22.04 AMI. The EC2 module enables nested virtualization, opens SSH access, and uses `init.sh` to install Firecracker and `jailer`.

The `setup.sh` script includes quickstart-style commands for downloading a kernel and root filesystem, creating a TAP device, and running a Firecracker microVM.

## Prerequisites

- Go, matching the versions declared in the modules and workspace.
- Terraform.
- AWS credentials configured for Terraform.
- An AWS key pair named `ec2aws`, or Terraform changes to use a different key pair.
- A Linux host with KVM support for running Firecracker.
- `firecracker` and `jailer` installed on the target host.

## Development Notes

The project is currently a work in progress. Several packages are placeholders, and some names and configuration values still need cleanup before production use. Treat the current code as an early prototype for Firecracker sandbox infrastructure rather than a complete sandbox platform.

## Useful Commands

Run the VM package:

```sh
go run ./packages/vm
```

Initialize and review Terraform infrastructure:

```sh
cd infra
terraform init
terraform plan -var="AWS_REGION=us-east-1"
```

Apply Terraform only after reviewing the generated plan and confirming the AWS cost and security implications.
