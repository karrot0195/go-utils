package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go-utils/packages/config"
	"go-utils/packages/log"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	CreateAction  = "create"
	DeleteAction  = "delete"
	DestroyAction = "destroy"
)

var (
	WorkerLabel     = getEnv("WORKER_LABEL", "retranscode-worker")
	JobStatsURL     = getEnv("JOB_STATS_URL", "https:/retranscoding.cloud/job-stats")
	DefaultRegion   = getEnv("DEFAULT_REGION", "sjc")
	DefaultSSHName  = getEnv("DEFAULT_SSH_NAME", "vultr-dev-access")
	APITimeout      = parseDuration(getEnv("API_TIMEOUT", "15s"))
	MaxResponseSize = int(parseInt(getEnv("MAX_RESPONSE_SIZE", "1048576")))
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Second
	}
	return d
}

func parseInt(s string) int64 {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 1 << 20
	}
	return n
}

// sizePlans maps a CLI size name to a Vultr plan slug.
var sizePlans = map[string]string{
	"high":   "vhp-8c-16gb-intel",
	"medium": "vhp-4c-8gb-intel",
	"small":  "vhp-1c-2gb-intel",
}

var logger, _ = log.NewLogger()

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() error {
	// load
	config.NewEnv().LoadEnv()

	actionCmd := flag.String("action", "", "action to perform: create | delete | destroy")
	apiKeyInParam := flag.String("api-key", "", "API key for the job-stats endpoint")
	sizeCmd := flag.String("size", "small", "size for create action: high | medium | small")
	snapshotImageName := flag.String("img", "retranscode-worker", "snapshot image name for create action")
	sshKeyName := flag.String("ssh-name", DefaultSSHName, "SSH key name (resolved to ID via vultr-cli)")

	flag.Parse()

	apiKey := *apiKeyInParam
	if apiKey == "" {
		apiKey = os.Getenv("API_KEY")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	h := &ActionHandler{ctx: ctx, apiKey: apiKey}

	switch *actionCmd {
	case CreateAction:
		logger.Info("creating new VM", "size", *sizeCmd, "image", *snapshotImageName, "ssh", *sshKeyName)
		return h.CreateAction(*sizeCmd, *snapshotImageName, *sshKeyName)
	case DeleteAction:
		return errors.New("delete action not implemented; use -action=destroy")
	case DestroyAction:
		return h.DestroyAction()
	default:
		flag.Usage()
		return fmt.Errorf("invalid action %q", *actionCmd)
	}
}

// ---------------------------------------------------------------------------
// ActionHandler
// ---------------------------------------------------------------------------

type ActionHandler struct {
	ctx    context.Context
	apiKey string
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func (h *ActionHandler) CreateAction(vmSize, snapShotImg, sshKeyName string) error {
	snapshotID, err := h.GetSnapshotID(snapShotImg)
	if err != nil {
		return err
	}

	sshKeyID, err := h.GetSSHKeyID(sshKeyName)
	if err != nil {
		return err
	}

	return h.createVMBySnapshotID(snapshotID, vmSize, sshKeyID)
}

func (h *ActionHandler) createVMBySnapshotID(snapshotID, vmSize, sshKeyID string) error {
	plan, ok := sizePlans[vmSize]
	if !ok {
		return fmt.Errorf("invalid VM size %q (want high|medium|small)", vmSize)
	}

	maxID, err := h.getMaxWorkerID()
	if err != nil {
		return err
	}
	instanceID := maxID + 1

	cmd := exec.CommandContext(h.ctx, "vultr-cli", "instance", "create",
		"--region", DefaultRegion,
		"--plan", plan,
		"--snapshot", snapshotID,
		"--ssh-keys", sshKeyID,
		"--label", fmt.Sprintf("%s-%d", WorkerLabel, instanceID),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create instance: %w, output: %s, cmd: %v",
			err, string(output), cmd.String())
	}

	logger.Info("created instance", "output", strings.TrimSpace(string(output)))
	return nil
}

// ---------------------------------------------------------------------------
// Destroy (guarded bulk removal of worker instances)
// ---------------------------------------------------------------------------

func (h *ActionHandler) DestroyAction() error {
	if err := h.ensureNoPendingJobs(); err != nil {
		return err
	}
	return h.destroyReadyWorkers()
}

type jobStats struct {
	Accepted int   `json:"accepted"`
	Draft    int   `json:"draft"`
	Jobs     []any `json:"jobs"`
}

func (h *ActionHandler) ensureNoPendingJobs() error {
	if h.apiKey == "" {
		return errors.New("missing API key: pass -api-key or set $API_KEY")
	}

	u, err := url.Parse(JobStatsURL)
	if err != nil {
		return fmt.Errorf("invalid job-stats URL: %w", err)
	}
	q := u.Query()
	q.Set("api_key", h.apiKey)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(h.ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("build job-stats request: %w", err)
	}

	client := &http.Client{Timeout: APITimeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch job-stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("job-stats returned status %d", resp.StatusCode)
	}

	var stats jobStats
	if err := json.NewDecoder(io.LimitReader(resp.Body, int64(MaxResponseSize))).Decode(&stats); err != nil {
		return fmt.Errorf("decode job-stats: %w", err)
	}

	if stats.Accepted+stats.Draft != 0 || len(stats.Jobs) != 0 {
		return fmt.Errorf("refusing to destroy: accepted=%d draft=%d jobs=%d",
			stats.Accepted, stats.Draft, len(stats.Jobs))
	}
	return nil
}

type instanceInfo struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Status       string `json:"status"`
	ServerStatus string `json:"server_status"`
	PowerStatus  string `json:"power_status"`
}

type instanceList struct {
	Instances []instanceInfo `json:"instances"`
}

func (h *ActionHandler) destroyReadyWorkers() error {
	workers, err := h.listWorkerInstances()
	if err != nil {
		return err
	}

	var destroyed, skipped int
	for _, inst := range workers {
		if !isReady(inst) {
			logger.Info("skipping instance (not ready)",
				"label", inst.Label, "id", inst.ID, "status", inst.Status)
			skipped++
			continue
		}
		if err := h.deleteInstance(inst.ID); err != nil {
			return err
		}
		destroyed++
	}

	logger.Info("destroy complete", "destroyed", destroyed, "skipped", skipped)
	return nil
}

func (h *ActionHandler) deleteInstance(id string) error {
	cmd := exec.CommandContext(h.ctx, "vultr-cli", "instance", "delete", id)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete instance %s: %w, output: %s",
			id, err, string(output))
	}
	if msg := strings.TrimSpace(string(output)); msg != "" {
		logger.Info("deleted instance", "id", id, "output", msg)
	} else {
		logger.Info("deleted instance", "id", id)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func isReady(inst instanceInfo) bool {
	return inst.Status == "active" &&
		inst.PowerStatus == "running" &&
		inst.ServerStatus == "ok"
}

// listWorkerInstances returns only instances whose label matches
// "<WorkerLabel>-<number>".
func (h *ActionHandler) listWorkerInstances() ([]instanceInfo, error) {
	all, err := h.fetchInstances()
	if err != nil {
		return nil, err
	}

	prefix := WorkerLabel + "-"
	var workers []instanceInfo
	for _, inst := range all {
		if !strings.HasPrefix(inst.Label, prefix) {
			continue
		}
		if _, err := strconv.Atoi(strings.TrimPrefix(inst.Label, prefix)); err != nil {
			continue
		}
		workers = append(workers, inst)
	}
	return workers, nil
}

// getMaxWorkerID returns the highest numeric suffix among worker instances,
// so a new worker gets a unique ID even if some were deleted out of order.
func (h *ActionHandler) getMaxWorkerID() (int, error) {
	workers, err := h.listWorkerInstances()
	if err != nil {
		return 0, err
	}

	prefix := WorkerLabel + "-"
	maxID := 0
	for _, inst := range workers {
		n, err := strconv.Atoi(strings.TrimPrefix(inst.Label, prefix))
		if err != nil {
			continue
		}
		if n > maxID {
			maxID = n
		}
	}
	return maxID, nil
}

func (h *ActionHandler) fetchInstances() ([]instanceInfo, error) {
	cmd := exec.CommandContext(h.ctx, "vultr-cli", "instance", "list", "-o", "json")
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

// GetSnapshotID resolves a snapshot description to its ID. Filtering is done
// in Go (no shell, no jq) to avoid command injection via the label argument.
func (h *ActionHandler) GetSnapshotID(snapshotImageName string) (string, error) {
	cmd := exec.CommandContext(h.ctx, "vultr-cli", "snapshot", "list", "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error listing snapshots: %w, output: %s", err, string(output))
	}

	var res struct {
		Snapshots []struct {
			ID          string `json:"id"`
			Description string `json:"description"`
		} `json:"snapshots"`
	}
	if err := json.Unmarshal(output, &res); err != nil {
		return "", fmt.Errorf("failed to parse snapshot list: %w", err)
	}

	for _, s := range res.Snapshots {
		if s.Description == snapshotImageName {
			return s.ID, nil
		}
	}
	return "", fmt.Errorf("no snapshot found for label %q", snapshotImageName)
}

// GetSSHKeyID resolves an SSH key name to its Vultr ID. Filtering is done
// in Go (no shell, no jq) to avoid command injection via the name argument.
func (h *ActionHandler) GetSSHKeyID(name string) (string, error) {
	cmd := exec.CommandContext(h.ctx, "vultr-cli", "ssh-key", "list", "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error listing ssh keys: %w, output: %s", err, string(output))
	}

	var res struct {
		SSHKeys []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"ssh_keys"`
	}
	if err := json.Unmarshal(output, &res); err != nil {
		return "", fmt.Errorf("failed to parse ssh key list: %w", err)
	}

	for _, k := range res.SSHKeys {
		if k.Name == name {
			return k.ID, nil
		}
	}
	return "", fmt.Errorf("no ssh key found with name %q", name)
}
