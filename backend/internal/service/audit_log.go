package service

import (
	"nagare/internal/model"
	"nagare/internal/repository"
)

// AuditResp represents an audit log response
type AuditResp struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Action    string `json:"action"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	IP        string `json:"ip"`
	Status    int    `json:"status"`
	Latency   int64  `json:"latency"`
	UserAgent string `json:"user_agent"`
	CreatedAt string `json:"created_at"`
}

// AddAuditLogServ adds a new audit log
func AddAuditLogServ(entry model.AuditLog) error {
	return repository.AddAuditLogDAO(entry)
}

// SearchAuditLogsServ searches for audit logs
func SearchAuditLogsServ(limit, offset int, query string) ([]AuditResp, int64, error) {
	logs, total, err := repository.SearchAuditLogsDAO(limit, offset, query)
	if err != nil {
		return nil, 0, err
	}

	resps := make([]AuditResp, 0, len(logs))
	for _, l := range logs {
		resp := AuditResp{
			ID:        l.ID,
			Username:  l.Username,
			Action:    l.Action,
			Method:    l.Method,
			Path:      l.Path,
			IP:        l.IP,
			Status:    l.Status,
			Latency:   l.Latency,
			UserAgent: l.UserAgent,
			CreatedAt: l.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if l.UserID != nil {
			resp.UserID = *l.UserID
		}
		resps = append(resps, resp)
	}

	return resps, total, nil
}
