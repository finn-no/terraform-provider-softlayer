# Terraform provider for SoftLayer

This is a terraform provider that lets you provision
servers on SoftLayer via [Terraform](https://terraform.io/).

## Installing

For now, there are no published artifacts/binaries for this provider.
In order to use it, you must compile the project from source, and then
put the `terraform-provider-softlayer` binary somewhere in your PATH.

## Using the provider

Example for setting up a virtual server with an SSH key:

```hcl
provider "softlayer" {
    username = ""
    api_key = ""
}

resource "softlayer_ssh_key" "my_key" {
    name = "my_key"
    public_key = "~/.ssh/id_rsa.pub"
}

resource "softlayer_virtualserver" "my_server" {
    name = "my_server"
    domain = "example.com"
    ssh_keys = ["${softlayer_ssh_key.my_key.keypair_id}"]
    image = "DEBIAN_7_64"
    region = "ams01"
    public_network_speed = 10
    cpu = 1
    ram = 1024
    disks = [25, 10, 20]
    user_data = "{\"fox\":[45]}"
}
```

You'll need to provide your SoftLayer username and API key,
so that Terraform can connect. If you don't want to put
credentials in your onfiguration file, you can leave them
out:

```
provider "softlayer" {}
```

...and instead set these environment variables:

- **SOFTLAYER_USERNAME**: Your SoftLayer username
- **SOFTLAYER_API_KEY**: Your API key

## Building

1) [Install Go](https://golang.org/doc/install) on your machine
2) [Set up Gopath](https://golang.org/doc/code.html)
3) `git clone` this repository into `$GOPATH/src/github.com/finn-no/terraform-provider-softlayer`
4) Run `go get` to get dependencies
5) Run `go install` to build the binary. You will now find the
   binary at `$GOPATH/bin/terraform-provider-softlayer`.