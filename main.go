package main

import (
	"github.com/finn-no/terraform-provider-softlayer/softlayer"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: softlayer.Provider,
	})
}
