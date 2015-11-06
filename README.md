# Terraform provider for SoftLayer

This is a terraform provider that lets you provision
servers on SoftLayer via [Terraform](https://terraform.io/).

## Installing

Binaries are published on Bintray: [ ![Download](https://api.bintray.com/packages/finn-no/terraform-provider-softlayer/terraform-provider-softlayer/images/download.svg) ](https://bintray.com/finn-no/terraform-provider-softlayer/terraform-provider-softlayer/_latestVersion)

[Copied from the Terraform documentation](https://www.terraform.io/docs/plugins/basics.html):
> To install a plugin, put the binary somewhere on your filesystem, then configure Terraform to be able to find it. The configuration where plugins are defined is ~/.terraformrc for Unix-like systems and %APPDATA%/terraform.rc for Windows.

You should update your .terraformrc and refer to the binary:

```hcl
providers {
  softlayer = "/path/to/terraform-provider-softlayer"
}
```

## A note about SSH keys

SoftLayer provisions SSH keys to new virtual servers **only once during the creation of the virtual server**. You can provision new SSH keys and assign them to the virtual server during the creation process, assign existing SSH keys by ID, or assign a combination of existing and newly provisioned SSH keys. Changing SSH keys assigned to a virtual server is not possible after it has already been created. It will need to be re-created. If you attempt to create a new SSH key using the ```softlayer_ssh_key``` resource type, and that key is already in the SoftLayer system, you will get an error stating that the key already exists. If this happens, use the Id of the existing SSH key as in the example below.

*To get a list of existing SSH key Id's from SoftLayer, I recommend using the [SoftLayer API Python Client]. Once the client is setup, you can run the following command and get the existing SSH key Id's from the output:*
```
[user@example.com ~]# slcli sshkey list
............
:........:...........:.................................................:..........................:
:   id   :   label   :                   fingerprint                   :          notes           :
:........:...........:.................................................:..........................:
: 123456 : test1-dev : xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx : Test Key 1               :
: 789101 : test2-dev : xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx:xx : Test Key 2               :
:........:...........:.................................................:..........................:
```

[SoftLayer API Python CLient]: http://softlayer-api-python-client.readthedocs.org/en/latest/install/
## Using the provider

Here is an example that will setup the following:
+ An SSH key resource.
+ A virtual server resource that uses an existing SSH key.
+ A virtual server resource using an existing SSH key and a Terraform managed SSH key (created as "test_key_1" in the example below).

(create this as sl.tf and run terraform commands from this directory):
```hcl
provider "softlayer" {
    username = ""
    api_key = ""
}

# This will create a new SSH key that will show up under the \
# Devices>Manage>SSH Keys in the SoftLayer console.
resource "softlayer_ssh_key" "test_key_1" {
    name = "test_key_1"
    public_key = "${file(\"~/.ssh/id_rsa_test_key_1.pub\")}"
    # Windows Example:
    # public_key = "${file(\"C:\ssh\keys\path\id_rsa_test_key_1.pub\")}"
}

# Virtual Server created with existing SSH Key already in SoftLayer \
# inventory and not created using this Terraform template.
resource "softlayer_virtualserver" "my_server_1" {
    name = "my_server_1"
    domain = "example.com"
    ssh_keys = ["123456"]
    image = "DEBIAN_7_64"
    region = "ams01"
    public_network_speed = 10
    cpu = 1
    ram = 1024
}

# Virtual Server created with a mix of previously existing and \
# Terraform created/managed resources.
resource "softlayer_virtualserver" "my_server_2" {
    name = "my_server_2"
    domain = "example.com"
    ssh_keys = ["123456", "${softlayer_ssh_key.test_key_1.id}"]
    image = "CENTOS_6_64"
    region = "ams01"
    public_network_speed = 10
    cpu = 1
    ram = 1024
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

## Building from source

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
