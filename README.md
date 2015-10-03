# Terraform provider for SoftLayer

This is a terraform provider that lets you provision
servers on SoftLayer via [Terraform](https://terraform.io/).

## Installing

For now, there are no published artifacts/binaries for this provider.
In order to use it, you must compile the project from source, and then
put the `terraform-provider-softlayer` binary somewhere in your system.

[Copied from the Terraform documentation](https://www.terraform.io/docs/plugins/basics.html):
> To install a plugin, put the binary somewhere on your filesystem, then configure Terraform to be able to find it. The configuration where plugins are defined is ~/.terraformrc for Unix-like systems and %APPDATA%/terraform.rc for Windows.

This is what you would need to add to your .terraformrc, if you have built the project from source:

```hcl
providers {
  softlayer = "$GOPATH/bin/terraform-provider-softlayer"
}
```

If the binary is placed somewhere else, update the path accordingly.

## Using the provider

Example for setting up a virtual server with an SSH key (create this as sl.tf and run terraform commands from this directory):

```hcl
provider "softlayer" {
    username = ""
    api_key = ""
}

resource "softlayer_ssh_key" "my_key" {
    name = "my_key"
    public_key = "${file(\"~/.ssh/id_rsa.pub\")}"
}

resource "softlayer_virtualserver" "my_server" {
    name = "my_server"
    domain = "example.com"
    ssh_keys = ["${softlayer_ssh_key.my_key.id}"]
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
credentials in your configuration file, you can leave them
out:

```
provider "softlayer" {}
```

...and instead set these environment variables:

- **SOFTLAYER_USERNAME**: Your SoftLayer username
- **SOFTLAYER_API_KEY**: Your API key

## Building

1.  [Install Go](https://golang.org/doc/install) on your machine
2.  [Set up Gopath](https://golang.org/doc/code.html)
3.  `git clone` this repository into `$GOPATH/src/github.com/finn-no/terraform-provider-softlayer`
4.  Run `go get` to get dependencies
5.  Run `go install` to build the binary. You will now find the
    binary at `$GOPATH/bin/terraform-provider-softlayer`.

## Running
0.  You must create a new key not already added to softlayer (ssh-keygen).  We will assume that is id_rsa.
1.  create the example file sl.tf in your working directory
2.  terraform plan
3.  terraform apply
4.  look up the public ip in the softlayer dashboard
5.  ssh -i ~/.ssh/id_rsa.pub root@<public-ip>
