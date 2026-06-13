package dto

import "time"

type ServerNodeResp struct {
	ID            uint       `json:"id"`
	NodeKey       string     `json:"nodeKey"`
	Name          string     `json:"name"`
	Hostname      string     `json:"hostname"`
	Roles         string     `json:"roles"`
	OS            string     `json:"os"`
	Arch          string     `json:"arch"`
	Version       string     `json:"version"`
	Status        string     `json:"status"`
	ResourceJSON  string     `json:"resourceJson"`
	LastHeartbeat *time.Time `json:"lastHeartbeat"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type ListServerNodeReq struct {
	PaginationRequest
	Status string `json:"status"`
	OS     string `json:"os"`
}

type ListServerNodeResp struct {
	PaginatedResponse[ServerNodeResp]
}

type CreateAgentTokenReq struct {
	Name      string     `json:"name" binding:"required,min=1,max=255"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

type CreateAgentTokenResp struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	Token     string     `json:"token"`
	ExpiresAt *time.Time `json:"expiresAt"`
	CreatedAt time.Time  `json:"createdAt"`
}

type AgentRegisterReq struct {
	NodeKey      string `json:"nodeKey" binding:"required,min=1,max=100"`
	Name         string `json:"name" binding:"max=255"`
	Hostname     string `json:"hostname" binding:"required,min=1,max=255"`
	Roles        string `json:"roles" binding:"required,min=1,max=100"`
	OS           string `json:"os" binding:"required,min=1,max=50"`
	Arch         string `json:"arch" binding:"required,min=1,max=50"`
	Version      string `json:"version" binding:"max=100"`
	ResourceJSON string `json:"resourceJson"`
}

type AgentHeartbeatReq struct {
	NodeKey      string        `json:"nodeKey" binding:"required,min=1,max=100"`
	ResourceJSON string        `json:"resourceJson"`
	Runners      []RunnerState `json:"runners"`
}

type RunnerState struct {
	ID        uint   `json:"id"`
	Status    string `json:"status"`
	ProcessID int    `json:"processId"`
	LastError string `json:"lastError"`
}

type PollRunnerTasksReq struct {
	NodeKey string `json:"nodeKey" binding:"required,min=1,max=100"`
	Limit   int    `json:"limit"`
}

type ReportRunnerTaskReq struct {
	NodeKey      string `json:"nodeKey" binding:"required,min=1,max=100"`
	Status       string `json:"status" binding:"required,min=1,max=50"`
	ResultJSON   string `json:"resultJson"`
	LogSummary   string `json:"logSummary"`
	LastError    string `json:"lastError"`
	RunnerStatus string `json:"runnerStatus"`
	ProcessID    int    `json:"processId"`
}

type RunnerResp struct {
	ID                  uint       `json:"id"`
	ProjectID           uint       `json:"projectId"`
	NodeID              uint       `json:"nodeId"`
	Name                string     `json:"name"`
	Labels              string     `json:"labels"`
	WorkDir             string     `json:"workDir"`
	InstallDir          string     `json:"installDir"`
	OS                  string     `json:"os"`
	Arch                string     `json:"arch"`
	Status              string     `json:"status"`
	ProcessID           int        `json:"processId"`
	LastError           string     `json:"lastError"`
	LastSeenAt          *time.Time `json:"lastSeenAt"`
	CreaterID           uint       `json:"createrId"`
	CreaterUserName     string     `json:"createrUserName"`
	LastUpdaterID       uint       `json:"lastUpdaterId"`
	LastUpdaterUserName string     `json:"lastUpdaterUserName"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

type CreateRunnerReq struct {
	ProjectID         uint   `json:"projectId" binding:"required,min=1"`
	NodeID            uint   `json:"nodeId" binding:"required,min=1"`
	Name              string `json:"name" binding:"required,min=1,max=255"`
	Labels            string `json:"labels" binding:"max=1000"`
	WorkDir           string `json:"workDir" binding:"max=1000"`
	RegistrationToken string `json:"registrationToken" binding:"required,min=1"`
}

type ListRunnerReq struct {
	PaginationRequest
	ProjectID uint   `json:"projectId"`
	NodeID    uint   `json:"nodeId"`
	Status    string `json:"status"`
}

type ListRunnerResp struct {
	PaginatedResponse[RunnerResp]
}

type RunnerTaskResp struct {
	ID           uint       `json:"id"`
	RunnerID     uint       `json:"runnerId"`
	NodeID       uint       `json:"nodeId"`
	Type         string     `json:"type"`
	Status       string     `json:"status"`
	PayloadJSON  string     `json:"payloadJson"`
	ResultJSON   string     `json:"resultJson"`
	LogSummary   string     `json:"logSummary"`
	LastError    string     `json:"lastError"`
	AttemptCount int        `json:"attemptCount"`
	ClaimedAt    *time.Time `json:"claimedAt"`
	StartedAt    *time.Time `json:"startedAt"`
	FinishedAt   *time.Time `json:"finishedAt"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}
