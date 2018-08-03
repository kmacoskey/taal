# Taal

Terraform as a (golang) library.

Terraform is a wonderful pattern and language for interacting with cloud API's. Unfortunately it is by design a standalone utility. The pattern of writing a program that interacts with the terraform standalone utility has been invented many times. This is yet another implentation, but hopefully more generic and re-usable as a golang libary.

## Installation

With a [correctly configured](https://golang.org/doc/install#testing) Go toolchain:

```sh
go get -u github.com/kmacoskey/taal
```

## Examples

Lets create some infrastructure. 

The terraform configuration can be HCL or JSON:

```go
config := []byte(`
  provider "google" {
    project     = "my-gce-project-id"
    region      = "us-central1"
  }

  resource "google_compute_instance" "test" {
    name         = "test"
    machine_type = "n1-standard-1"
    zone         = "us-central1-a"

    network_interface {
      network = "default"
    }

    boot_disk {
      initialize_params {
        image = "debian-cloud/debian-8"
      }
    }
  }`)
```

Credentials must be provided:

```
credentials := []byte(` Your Credentials `)
```

With Credentials and Config, you can terraform apply and destroy:

```go
t := taal.NewInfra()

t.Config(config)
t.Credentials(credentials)

if err := t.Apply(); err != nil {
  panic(fmt.Println("Error applying terraform config"))
}

if err := t.Destroy(); err != nil {
  panic(fmt.Println("Error destorying terraform config"))
}
```

This is all you need to know for basic usage. More advanced options are explained below.

### Terraform State

The terraform state can be retrieved for exporting and then subsequent uses.

```go
if err := t.Apply(); err != nil {
  panic(fmt.Println("Error applying terraform config"))
}

// Export the current state
state := t.State()

...

// Import and use the previous state
t_new := taal.NewInfra()
...
t_new.SetState(state)

if err := t_new.Destroy(); err != nil 
  panic(fmt.Println("Error destorying terraform config"))
}
```

### Terraform Input Variables

Supply [input variables](https://www.terraform.io/docs/configuration/variables.html).

```go
config := []byte(`
  variable "name" { type = "string" }
  variable "instance_type" { type = "string" }
`)
t.Config(config)

...

inputs := {
  name: "foo"
  instance_type: "n1-standard-1"
}

t.Inputs(intputs)

if err := t.Apply(); err != nil {
  panic(fmt.Println("Error applying terraform config"))
}
```

### Terraform Outputs

Access terraform outputs.

```go
config := []byte(`
output "address" {
  value = "${compute_instance.test.network_interface.0.address}"
}`)
t.Config(config)

credentials := []byte(` Your Credentials `)
t.Credentials(credentials)

if err := t.Apply(); err != nil {
  panic(fmt.Println("Error applying terraform config"))
}

outputs := t.Outputs()
fmt.Println("address: %s", outputs["address"])
```

## How to Test

```sh
make test
```

