package ncloud

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	RegisterResource("ncloud_network_interface_attachment", resourceNcloudNetworkInterfaceAttachment())
}

func resourceNcloudNetworkInterfaceAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceNcloudNetworkInterfaceAttachmentCreate,
		Read:   resourceNcloudNetworkInterfaceAttachmentRead,
		Delete: resourceNcloudNetworkInterfaceAttachmentDelete,
		Schema: map[string]*schema.Schema{
			"attachment_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"order": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"server_instance_no": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network_interface_no": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceNcloudNetworkInterfaceAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)

	networkInterfaceNo := d.Get("network_interface_no").(string)
	niInstance, err := getNetworkInterface(config, networkInterfaceNo)
	if err != nil {
		return err
	}

	serverInstanceNo := d.Get("server_instance_no").(string)
	if err := attachVpcNetworkInterface(
		config,
		networkInterfaceNo,
		*niInstance.SubnetNo,
		serverInstanceNo,
	); err != nil {
		return err
	}

	attachmentId, err := newAttachmentId(networkInterfaceNo, serverInstanceNo)
	if err != nil {
		return err
	}

	d.SetId(attachmentId.Id())

	return resourceNcloudNetworkInterfaceAttachmentRead(d, meta)
}

func resourceNcloudNetworkInterfaceAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)

	attachmentId, err := attachmentIdFromString(d.Id())
	if err != nil {
		return err
	}

	niInstance, err := getNetworkInterface(config, attachmentId.NetworkInterfaceNo())
	if err != nil {
		return err
	}

	if niInstance == nil || *niInstance.InstanceNo != attachmentId.ServerInstanceNo() {
		d.SetId("")
		return nil
	}

	order, err := ParseNetworkInterfaceOrder(niInstance)
	if err != nil {
		return err
	}

	d.Set("attachment_id", d.Id())
	d.Set("order", order)
	d.Set("server_instance_no", attachmentId.ServerInstanceNo())
	d.Set("network_interface_no", attachmentId.NetworkInterfaceNo())
	d.Set("status", niInstance.NetworkInterfaceStatus.Code)

	return nil
}

func resourceNcloudNetworkInterfaceAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)

	networkInterfaceNo := strings.Split(d.Id(), "_")[0]
	niInstance, err := getNetworkInterface(config, networkInterfaceNo)
	if err != nil {
		return err
	}

	serverInstanceNo := d.Get("server_instance_no").(string)
	if err := detachVpcNetworkInterface(
		config,
		networkInterfaceNo,
		*niInstance.SubnetNo,
		serverInstanceNo,
	); err != nil {
		return err
	}

	return nil
}

type attachmentId struct {
	id                 string
	networkInterfaceNo string
	serverInstanceNo   string
}

func newAttachmentId(networkInterfaceNo, serverInstanceNo string) (*attachmentId, error) {
	id := fmt.Sprintf("%s_%s", networkInterfaceNo, serverInstanceNo)
	return attachmentIdFromString(id)
}

func attachmentIdFromString(id string) (*attachmentId, error) {
	ids := strings.Split(id, "_")
	if len(ids) != 2 {
		err := fmt.Errorf("invalid id format, required <NetworkInterfaceNo>_<serverInstanceNo> form")
		return nil, err
	}

	return &attachmentId{
		id:                 id,
		networkInterfaceNo: ids[0],
		serverInstanceNo:   ids[1],
	}, nil
}

func (a attachmentId) Id() string {
	return a.id
}

func (a attachmentId) NetworkInterfaceNo() string {
	return a.networkInterfaceNo
}

func (a attachmentId) ServerInstanceNo() string {
	return a.serverInstanceNo
}
