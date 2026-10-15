package vultrcmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// ---------------------------------------------------------------------------
// VultrCmd
// ---------------------------------------------------------------------------

// VultrCmd wraps vultr-cli operations.
type VultrCmd struct {
	ctx context.Context
}

// NewVultrCmd creates a new VultrCmd. If ctx is nil, context.Background() is used.
func NewVultrCmd(ctx context.Context) *VultrCmd {
	if ctx == nil {
		ctx = context.Background()
	}
	return &VultrCmd{ctx: ctx}
}

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// Instance represents a Vultr instance.
type Instance struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Status       string `json:"status"`
	ServerStatus string `json:"server_status"`
	PowerStatus  string `json:"power_status"`
	MainIP       string `json:"main_ip"`
	IP           string `json:"ip"` // fallback for some CLI versions
}

type instanceList struct {
	Instances []Instance `json:"instances"`
}

// IP returns the primary IP for the instance, checking both possible fields.
func (i Instance) PrimaryIP() string {
	if i.MainIP != "" {
		return i.MainIP
	}
	return i.IP
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

// GetInstances returns all Vultr instances.
func (v *VultrCmd) GetInstances() ([]Instance, error) {
	cmd := exec.CommandContext(v.ctx, "vultr-cli", "instance", "list", "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w, output: %s", err, string(output))
	}

	var list instanceList
	if err := json.Unmarshal(output, &list); err != nil {
		return nil, fmt.Errorf("failed to parse instance list: %w", err)
	}
	return list.Instances, nil
}

// ---------------------------------------------------------------------------
// Delete by IP
// ---------------------------------------------------------------------------

// RemoveInstanceByIP deletes the instance whose primary IP matches ip.
// Returns the deleted instance's ID, or an error if not found / delete fails.
func (v *VultrCmd) RemoveInstanceByIP(ip string) (string, error) {
	instances, err := v.GetInstances()
	if err != nil {
		return "", err
	}

	for _, inst := range instances {
		if inst.PrimaryIP() == ip {
			if err := v.deleteInstance(inst.ID); err != nil {
				return "", err
			}
			return inst.ID, nil
		}
	}
	return "", fmt.Errorf("no instance found with IP %s of %v", ip, instances)
}

func (v *VultrCmd) deleteInstance(id string) error {
	cmd := exec.CommandContext(v.ctx, "vultr-cli", "instance", "delete", id)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete instance %s: %w, output: %s ",
			id, err, string(output))
	}
	fmt.Printf("Deleted instance %s: %s", id, string(output))
	return nil
}
