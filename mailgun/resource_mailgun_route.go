package mailgun

import (
	"context"
	"fmt"
	"log"
	"time"

	mailgun "github.com/mailgun/mailgun-go/v3"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceMailgunRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceMailgunRouteCreate,
		Read:   resourceMailgunRouteRead,
		Update: resourceMailgunRouteUpdate,
		Delete: resourceMailgunRouteDelete,

		Schema: map[string]*schema.Schema{
			"priority": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: false,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},

			"expression": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},

			"actions": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceMailgunRouteCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*mailgun.MailgunImpl)

	opts := mailgun.Route{}

	opts.Priority = d.Get("priority").(int)
	opts.Description = d.Get("description").(string)
	opts.Expression = d.Get("expression").(string)
	actions := d.Get("actions").([]interface{})
	actionArray := []string{}
	for _, i := range actions {
		action := i.(string)
		actionArray = append(actionArray, action)
	}
	opts.Actions = actionArray
	log.Printf("[DEBUG] Route create configuration: %v", opts)

	route, err := client.CreateRoute(context.Background(), opts)

	if err != nil {
		return err
	}

	d.SetId(route.Id)

	log.Printf("[INFO] Route ID: %s", d.Id())

	// Retrieve and update state of route
	_, err = resourceMailgunRouteRetrieve(d.Id(), client, d)

	if err != nil {
		return err
	}

	return nil
}

func resourceMailgunRouteUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*mailgun.MailgunImpl)

	opts := mailgun.Route{}

	opts.Priority = d.Get("priority").(int)
	opts.Description = d.Get("description").(string)
	opts.Expression = d.Get("expression").(string)
	actions := d.Get("actions").([]interface{})
	actionArray := []string{}

	for _, i := range actions {
		action := i.(string)
		actionArray = append(actionArray, action)
	}
	opts.Actions = actionArray

	log.Printf("[DEBUG] Route update configuration: %v", opts)

	route, err := client.UpdateRoute(context.Background(), d.Id(), opts)

	if err != nil {
		return err
	}

	d.SetId(route.Id)

	log.Printf("[INFO] Route ID: %s", d.Id())

	// Retrieve and update state of route
	_, err = resourceMailgunRouteRetrieve(d.Id(), client, d)

	if err != nil {
		return err
	}

	return nil
}

func resourceMailgunRouteDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*mailgun.MailgunImpl)

	log.Printf("[INFO] Deleting Route: %s", d.Id())

	// Destroy the route
	err := client.DeleteRoute(context.Background(), d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting route: %s", err)
	}

	// Give the destroy a chance to take effect
	return resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err = client.GetRoute(context.Background(), d.Id())
		if err == nil {
			log.Printf("[INFO] Retrying until route disappears...")
			return resource.RetryableError(
				fmt.Errorf("route seems to still exist; will check again"))
		}
		log.Printf("[INFO] Got error looking for route, seems gone: %s", err)
		return nil
	})
}

func resourceMailgunRouteRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*mailgun.MailgunImpl)

	_, err := resourceMailgunRouteRetrieve(d.Id(), client, d)

	if err != nil {
		return err
	}

	return nil
}

func resourceMailgunRouteRetrieve(id string, client *mailgun.MailgunImpl, d *schema.ResourceData) (*mailgun.Route, error) {

	route, err := client.GetRoute(context.Background(), id)

	if err != nil {
		return nil, fmt.Errorf("Error retrieving route: %s", err)
	}

	d.Set("priority", route.Priority)
	d.Set("description", route.Description)
	d.Set("expression", route.Expression)
	d.Set("actions", route.Actions)

	return &route, nil
}
