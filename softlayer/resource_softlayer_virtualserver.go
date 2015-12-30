package softlayer

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	datatypes "github.com/maximilien/softlayer-go/data_types"
)

func resourceSoftLayerVirtualserver() *schema.Resource {
	return &schema.Resource{
		Create: resourceSoftLayerVirtualserverCreate,
		Read:   resourceSoftLayerVirtualserverRead,
		Update: resourceSoftLayerVirtualserverUpdate,
		Delete: resourceSoftLayerVirtualserverDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"domain": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"image": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"image_type": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "os_code",
				ValidateFunc: validateImageType,
			},

			"tags": &schema.Schema{
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},

			"region": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"cpu": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},

			"ram": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},

			"disks": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},

			"public_network_speed": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1000,
			},

			"ipv4_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"ipv4_address_private": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"ssh_keys": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},

			"user_data": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

// this fellow validates image_type
func validateImageType(value interface{}, key string) ([]string, []error) {
	// image_type can be only either 'os_code' or 'template_id'
	if value != "os_code" && value != "template_id" {
		return nil, []error{
			errors.New(fmt.Sprintf("%s must be either 'os_code' or 'template_id'", key)),
		}
	}
	return nil, nil
}

func getNameForBlockDevice(i int) string {
	// skip 1, which is reserved for the swap disk.
	// so we get 0, 2, 3, 4, 5 ...
	if i == 0 {
		return "0"
	} else {
		return strconv.Itoa(i + 1)
	}
}

func getBlockDevices(d *schema.ResourceData) []datatypes.BlockDevice {
	numBlocks := d.Get("disks.#").(int)
	if numBlocks == 0 {
		return nil
	} else {
		blocks := make([]datatypes.BlockDevice, 0, numBlocks)
		for i := 0; i < numBlocks; i++ {
			blockRef := fmt.Sprintf("disks.%d", i)
			name := getNameForBlockDevice(i)
			capacity := d.Get(blockRef).(int)
			block := datatypes.BlockDevice{
				Device: name,
				DiskImage: datatypes.DiskImage{
					Capacity: capacity,
				},
			}
			blocks = append(blocks, block)
		}
		return blocks
	}
}

func resourceSoftLayerVirtualserverCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).virtualGuestService
	if client == nil {
		return fmt.Errorf("The client was nil.")
	}

	dc := datatypes.Datacenter{
		Name: d.Get("region").(string),
	}

	networkComponent := datatypes.NetworkComponents{
		MaxSpeed: d.Get("public_network_speed").(int),
	}

	// Options to be passed to Softlayer API
	opts := datatypes.SoftLayer_Virtual_Guest_Template{
		Hostname:          d.Get("name").(string),
		Domain:            d.Get("domain").(string),
		HourlyBillingFlag: true,
		Datacenter:        dc,
		StartCpus:         d.Get("cpu").(int),
		MaxMemory:         d.Get("ram").(int),
		NetworkComponents: []datatypes.NetworkComponents{networkComponent},
		BlockDevices:      getBlockDevices(d),
	}

	// get image type before doing anything
	image_type := d.Get("image_type").(string)

	// get the image ID
	image := d.Get("image").(string)

	// 'image' value is interpreted depending on the
	// value set on 'image_type'
	if image_type == "template_id" {
		// Set ID of the base image template to be used (if any)
		base_image := datatypes.BlockDeviceTemplateGroup{image}
		opts.BlockDeviceTemplateGroup = &base_image
	} else {
		// Use a stock OS instead for this resource
		opts.OperatingSystemReferenceCode = image
	}

	userData := d.Get("user_data").(string)
	if userData != "" {
		opts.UserData = []datatypes.UserData{
			datatypes.UserData{
				Value: userData,
			},
		}
	}

	// Get configured ssh_keys
	ssh_keys := d.Get("ssh_keys.#").(int)
	if ssh_keys > 0 {
		opts.SshKeys = make([]datatypes.SshKey, 0, ssh_keys)
		for i := 0; i < ssh_keys; i++ {
			key := fmt.Sprintf("ssh_keys.%d", i)
			id := d.Get(key).(int)
			sshKey := datatypes.SshKey{
				Id: id,
			}
			opts.SshKeys = append(opts.SshKeys, sshKey)
		}
	}

	log.Printf("[INFO] Creating virtual machine")

	guest, err := client.CreateObject(opts)

	if err != nil {
		return fmt.Errorf("Error creating virtual server: %s", err)
	}

	d.SetId(fmt.Sprintf("%d", guest.Id))

	log.Printf("[INFO] Virtual Machine ID: %s", d.Id())

	// wait for machine availability
	_, err = WaitForNoActiveTransactions(d, meta)

	if err != nil {
		return fmt.Errorf(
			"Error waiting for virtual machine (%s) to become ready: %s", d.Id(), err)
	}

	_, err = WaitForPublicIpAvailable(d, meta)
	if err != nil {
		return fmt.Errorf(
			"Error waiting for virtual machine (%s) to become ready: %s", d.Id(), err)
	}

	// insert tags on target virtual guest
	num_tags := d.Get("tags.#").(int)
	if num_tags > 0 {
		log.Printf("[INFO] settings tags on virtual server")
		// extract each tag first of all
		tags := make([]string, 0, num_tags)
		for i := 0; i < num_tags; i++ {
			key := fmt.Sprintf("tags.%d", i)
			tag := d.Get(key).(string)
			tags = append(tags, tag)
		}
		// set the actual tags on the instance
		client.SetTags(guest.Id, tags)
	}

	return resourceSoftLayerVirtualserverRead(d, meta)
}

func resourceSoftLayerVirtualserverRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).virtualGuestService
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}
	result, err := client.GetObject(id)
	if err != nil {
		return fmt.Errorf("Error retrieving virtual server: %s", err)
	}

	d.Set("name", result.Hostname)
	d.Set("domain", result.Domain)
	if result.Datacenter != nil {
		d.Set("region", result.Datacenter.Name)
	}
	d.Set("public_network_speed", result.NetworkComponents[0].MaxSpeed)
	d.Set("cpu", result.StartCpus)
	d.Set("ram", result.MaxMemory)
	d.Set("has_public_ip", result.PrimaryIpAddress != "")
	d.Set("ipv4_address", result.PrimaryIpAddress)
	d.Set("ipv4_address_private", result.PrimaryBackendIpAddress)

	connIpAddress := ""
	if result.PrimaryIpAddress != "" {
		connIpAddress = result.PrimaryIpAddress
	} else {
		connIpAddress = result.PrimaryBackendIpAddress
	}

	log.Printf("[INFO] Setting ConnInfo IP: %s", connIpAddress)
	d.SetConnInfo(map[string]string{
		"type": "ssh",
		"host": connIpAddress,
	})

	return nil
}

func resourceSoftLayerVirtualserverUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).virtualGuestService
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}
	result, err := client.GetObject(id)
	if err != nil {
		return fmt.Errorf("Error retrieving virtual server: %s", err)
	}

	result.Hostname = d.Get("name").(string)
	result.Domain = d.Get("domain").(string)
	result.StartCpus = d.Get("cpu").(int)
	result.MaxMemory = d.Get("ram").(int)
	result.NetworkComponents[0].MaxSpeed = d.Get("public_network_speed").(int)

	userData := d.Get("user_data").(string)
	if userData != "" {
		result.UserData = []datatypes.UserData{
			datatypes.UserData{
				Value: userData,
			},
		}
	}

	_, err = client.EditObject(id, result)

	if err != nil {
		return fmt.Errorf("Couldn't update virtual server: %s", err)
	}

	return nil
}

func resourceSoftLayerVirtualserverDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).virtualGuestService
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	_, err = WaitForNoActiveTransactions(d, meta)

	if err != nil {
		return fmt.Errorf("Error deleting virtual server, couldn't wait for zero active transactions: %s", err)
	}

	_, err = client.DeleteObject(id)

	if err != nil {
		return fmt.Errorf("Error deleting virtual server: %s", err)
	}

	return nil
}

func WaitForPublicIpAvailable(d *schema.ResourceData, meta interface{}) (interface{}, error) {
	log.Printf("Waiting for server (%s) to get a public IP", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending: []string{"", "unavailable"},
		Target:  "available",
		Refresh: func() (interface{}, string, error) {
			fmt.Println("Refreshing server state...")
			client := meta.(*Client).virtualGuestService
			id, err := strconv.Atoi(d.Id())
			if err != nil {
				return nil, "", fmt.Errorf("Not a valid ID, must be an integer: %s", err)
			}
			result, err := client.GetObject(id)
			if err != nil {
				return nil, "", fmt.Errorf("Error retrieving virtual server: %s", err)
			}
			if result.PrimaryIpAddress == "" {
				return result, "unavailable", nil
			} else {
				return result, "available", nil
			}
		},
		Timeout:    30 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	return stateConf.WaitForState()
}

func WaitForNoActiveTransactions(d *schema.ResourceData, meta interface{}) (interface{}, error) {
	log.Printf("Waiting for server (%s) to have zero active transactions", d.Id())
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return nil, fmt.Errorf("The instance ID %s must be numeric", d.Id())
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{"", "active"},
		Target:  "idle",
		Refresh: func() (interface{}, string, error) {
			client := meta.(*Client).virtualGuestService
			transactions, err := client.GetActiveTransactions(id)
			if err != nil {
				return nil, "", fmt.Errorf("Couldn't get active transactions: %s", err)
			}
			if len(transactions) == 0 {
				return transactions, "idle", nil
			} else {
				return transactions, "active", nil
			}
		},
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	return stateConf.WaitForState()
}
