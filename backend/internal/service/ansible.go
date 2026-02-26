package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"nagare/internal/model"
	"nagare/internal/repository"
)

type PlaybookReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Content     string `json:"content" binding:"required"`
	Tags        string `json:"tags"`
}

type PlaybookResp struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
	Tags        string `json:"tags"`
	CreatedAt   string `json:"created_at"`
}

type AnsibleJobResp struct {
	ID         uint   `json:"id"`
	PlaybookID uint   `json:"playbook_id"`
	PlaybookName string `json:"playbook_name"`
	Status     string `json:"status"`
	Output     string `json:"output"`
	CreatedAt  string `json:"created_at"`
	HostFilter string `json:"host_filter"`
}

// ============= Inventory Service =============

func GetAnsibleDynamicInventory() (map[string]interface{}, error) {
	hosts, err := repository.GetAllHostsDAO()
	if err != nil {
		return nil, err
	}

	inventory := make(map[string]interface{})
	meta := make(map[string]interface{})
	hostvars := make(map[string]interface{})

	allHosts := []string{}
	groupMap := make(map[uint][]string)

	for _, h := range hosts {
		if h.Enabled == 0 {
			continue
		}
		
		hostKey := h.Name
		if hostKey == "" {
			hostKey = h.IPAddr
		}
		
		allHosts = append(allHosts, hostKey)
		
		// Map by group
		if h.GroupID > 0 {
			groupMap[h.GroupID] = append(groupMap[h.GroupID], hostKey)
		}

		// SSH Connection details
		vars := map[string]interface{}{
			"ansible_host": h.IPAddr,
			"ansible_port": h.SSHPort,
			"ansible_user": h.SSHUser,
		}
		
		// If we have decrypted password in context, we could use it
		// But usually it's better to use SSH Keys on the control node
		// For this implementation, we assume the Control Node (Backend) has access
		
		hostvars[hostKey] = vars
	}

	inventory["all"] = map[string]interface{}{"hosts": allHosts}
	
	// Add groups
	for gid, gHosts := range groupMap {
		group, err := repository.GetGroupByIDDAO(gid)
		if err == nil {
			groupName := strings.ReplaceAll(strings.ToLower(group.Name), " ", "_")
			inventory[groupName] = map[string]interface{}{"hosts": gHosts}
		}
	}

	meta["hostvars"] = hostvars
	inventory["_meta"] = meta

	return inventory, nil
}

// ============= Playbook Service =============

func CreatePlaybookServ(req PlaybookReq) (PlaybookResp, error) {
	pb := model.AnsiblePlaybook{
		Name:        req.Name,
		Description: req.Description,
		Content:     req.Content,
		Tags:        req.Tags,
	}
	if err := repository.CreatePlaybookDAO(&pb); err != nil {
		return PlaybookResp{}, err
	}
	return playbookToResp(pb), nil
}

func ListPlaybooksServ(query string) ([]PlaybookResp, error) {
	pbs, err := repository.ListPlaybooksDAO(query)
	if err != nil {
		return nil, err
	}
	res := make([]PlaybookResp, 0, len(pbs))
	for _, pb := range pbs {
		res = append(res, playbookToResp(pb))
	}
	return res, nil
}

func GetPlaybookServ(id uint) (PlaybookResp, error) {
	pb, err := repository.GetPlaybookByIDDAO(id)
	if err != nil {
		return PlaybookResp{}, err
	}
	return playbookToResp(pb), nil
}

func UpdatePlaybookServ(id uint, req PlaybookReq) error {
	return repository.UpdatePlaybookDAO(id, model.AnsiblePlaybook{
		Name:        req.Name,
		Description: req.Description,
		Content:     req.Content,
		Tags:        req.Tags,
	})
}

func DeletePlaybookServ(id uint) error {
	return repository.DeletePlaybookDAO(id)
}

// ============= Execution Service =============

func RunAnsiblePlaybookServ(playbookID uint, hostFilter string, userID *uint) (uint, error) {
	pb, err := repository.GetPlaybookByIDDAO(playbookID)
	if err != nil {
		return 0, err
	}

	job := model.AnsibleJob{
		PlaybookID:  playbookID,
		Status:      "running",
		TriggeredBy: userID,
		HostFilter:  hostFilter,
	}
	if err := repository.CreateAnsibleJobDAO(&job); err != nil {
		return 0, err
	}

	go executePlaybook(job.ID, pb.Content, hostFilter)

	return job.ID, nil
}

func executePlaybook(jobID uint, content string, hostFilter string) {
	// 1. Create temporary playbook file
	tmpDir := "temp/ansible"
	_ = os.MkdirAll(tmpDir, 0755)
	
	tmpFile, err := os.CreateTemp(tmpDir, "playbook-*.yml")
	if err != nil {
		updateJobStatus(jobID, "failed", "Failed to create temp file: "+err.Error())
		return
	}
	defer os.Remove(tmpFile.Name())
	
	if _, err := tmpFile.WriteString(content); err != nil {
		updateJobStatus(jobID, "failed", "Failed to write playbook: "+err.Error())
		return
	}
	tmpFile.Close()

	// 2. Prepare dynamic inventory
	inv, _ := GetAnsibleDynamicInventory()
	invFile, _ := os.CreateTemp(tmpDir, "inventory-*.json")
	defer os.Remove(invFile.Name())
	
	invData, _ := json.Marshal(inv)
	invFile.Write(invData)
	invFile.Close()

	// 3. Build command
	args := []string{"-i", invFile.Name(), tmpFile.Name()}
	if hostFilter != "" && hostFilter != "all" {
		args = append(args, "-l", hostFilter)
	}

	var cmd *exec.Cmd
	
	if runtime.GOOS == "windows" {
		// On Windows, we force the code page to 65001 (UTF-8) before running ansible.
		// We use cmd /c to execute multiple commands in one shell session.
		quotedArgs := make([]string, len(args))
		for i, arg := range args {
			// Basic quoting for Windows shell
			quotedArgs[i] = "\"" + arg + "\""
		}
		fullCmd := "chcp 65001 >nul && ansible-playbook " + strings.Join(quotedArgs, " ")
		cmd = exec.Command("cmd", "/c", fullCmd)
		
		cmd.Stdin = strings.NewReader("")
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "ANSIBLE_STDOUT_CALLBACK=json")
		cmd.Env = append(cmd.Env, "PYTHONUTF8=1")
		cmd.Env = append(cmd.Env, "PYTHONIOENCODING=UTF-8")
	} else {
		cmd = exec.Command("ansible-playbook", args...)
		cmd.Stdin = strings.NewReader("")
	}
	
	// Create pipes for stdout/stderr
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	multi := io.MultiReader(stdout, stderr)

	if err := cmd.Start(); err != nil {
		// If native start fails on Windows, suggest WSL
		errMsg := err.Error()
		if runtime.GOOS == "windows" {
			errMsg += "\n\nTip: Native Ansible on Windows is experimental. Consider using WSL."
		}
		updateJobStatus(jobID, "failed", "Failed to start ansible: "+errMsg)
		return
	}

	// Read output in real-time
	scanner := bufio.NewScanner(multi)
	var fullOutput strings.Builder
	
	for scanner.Scan() {
		line := scanner.Text() + "\n"
		fullOutput.WriteString(line)
		
		// Stream via WebSocket
		BroadcastMessage(map[string]interface{}{
			"event": "ansible_log",
			"job_id": jobID,
			"data": line,
		})
		
		// Update DB periodically or at end? For now, we'll update at end to save DB load
		// but for long jobs, periodic update is better.
	}

	if err := cmd.Wait(); err != nil {
		updateJobStatus(jobID, "failed", fullOutput.String())
		LogService("error", fmt.Sprintf("Ansible Job Failed: Job #%d failed.", jobID), map[string]interface{}{"job_id": jobID}, nil, "")
	} else {
		updateJobStatus(jobID, "success", fullOutput.String())
		LogService("info", fmt.Sprintf("Ansible Job Success: Job #%d finished successfully.", jobID), map[string]interface{}{"job_id": jobID}, nil, "")
	}
}

func updateJobStatus(id uint, status string, output string) {
	_ = repository.UpdateAnsibleJobDAO(id, status, output)
	
	// Notify frontend of status change
	BroadcastMessage(map[string]interface{}{
		"event": "ansible_status",
		"job_id": id,
		"status": status,
	})
}

func GetAnsibleJobServ(id uint) (AnsibleJobResp, error) {
	job, err := repository.GetAnsibleJobByIDDAO(id)
	if err != nil {
		return AnsibleJobResp{}, err
	}
	return jobToResp(job), nil
}

func ListAnsibleJobsServ(playbookID uint, limit int) ([]AnsibleJobResp, error) {
	jobs, err := repository.ListAnsibleJobsDAO(playbookID, limit)
	if err != nil {
		return nil, err
	}
	res := make([]AnsibleJobResp, 0, len(jobs))
	for _, job := range jobs {
		res = append(res, jobToResp(job))
	}
	return res, nil
}

// ============= AI Helpers =============

func RecommendAnsiblePlaybookServ(context string) (string, error) {
	// Logic to use LLM to generate a playbook based on error message or requirement
	prompt := fmt.Sprintf(`You are an Ansible expert specializing in Huawei network automation. 
%s

Generate a YAML playbook based on the following requirement:
%s

Respond ONLY with the YAML content. Use 'all' as the default hosts if not specified.
Prefer using Huawei-specific modules (e.g., community.network.huawei_vrp) or generic network modules if appropriate.`, systemContextPrompt(), context)
	
	// Re-using SendChatServ logic
	res, err := SendChatServ(ChatReq{
		ProviderID: 1, // Default provider
		Content: prompt,
	})
	if err != nil {
		return "", err
	}
	
	return res.Content, nil
}

// ============= Mappers =============

func playbookToResp(pb model.AnsiblePlaybook) PlaybookResp {
	return PlaybookResp{
		ID:          pb.ID,
		Name:        pb.Name,
		Description: pb.Description,
		Content:     pb.Content,
		Tags:        pb.Tags,
		CreatedAt:   pb.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func jobToResp(job model.AnsibleJob) AnsibleJobResp {
	return AnsibleJobResp{
		ID:           job.ID,
		PlaybookID:   job.PlaybookID,
		PlaybookName: job.Playbook.Name,
		Status:       job.Status,
		Output:       job.Output,
		HostFilter:   job.HostFilter,
		CreatedAt:    job.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
