package provider

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var dialer net.Dialer

func init() {
	schema.DescriptionKind = schema.StringMarkdown

	rand.Seed(time.Now().Unix())
}

// New returns a *schema.Provider.
func New() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{},

		ResourcesMap: map[string]*schema.Resource{
			"wol_mac": wolMacResource,
		},

		DataSourcesMap: map[string]*schema.Resource{},
	}
}

var wolMacResource = &schema.Resource{
	Description: "send wol magic to mac address",

	Schema: map[string]*schema.Schema{
		"mac": {
			Description: "mac address",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"send": {
			Description: "true to send",
			Type:        schema.TypeBool,
			Required:    true,
		},
	},

	CreateContext: onCreateOrUpdate,
	UpdateContext: onCreateOrUpdate,
	ReadContext:   noop,
	DeleteContext: noop,
}

func noop(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}
func onCreateOrUpdate(ctx context.Context, r *schema.ResourceData, _ interface{}) diag.Diagnostics {
	if !r.Get("send").(bool) {
		return nil
	}
	err := sendWol(ctx, r.Get("mac").(string))
	if err != nil {
		return errToDiag(err)
	}
	return nil
}

func sendWol(ctx context.Context, macStr string) error {
	mac, err := net.ParseMAC(macStr)
	if err != nil {
		return err
	}
	conn, err := dialer.DialContext(ctx, "udp", "255.255.255.255:9")
	if err != nil {
		return err
	}
	defer conn.Close()

	var magic [17 * 6]byte
	copy(magic[0:], []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	for i := 1; i <= 16; i++ {
		copy(magic[i*16:], mac)
	}
	_, err = conn.Write(magic[:])
	if err != nil {
		return err
	}

	return nil
}

func errToDiag(errs ...error) (diags diag.Diagnostics) {
	for _, err := range errs {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  err.Error(),
			Detail:   fmt.Sprintf("%+v", err),
		})
	}
	return
}
