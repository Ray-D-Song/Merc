package model

import "time"

const (
	NodeRoleServer = "server"
	NodeRoleAgent  = "agent"

	NodeStatusOnline   = "online"
	NodeStatusOffline  = "offline"
	NodeStatusDraining = "draining"

	RunnerStatusPending    = "pending"
	RunnerStatusInstalling = "installing"
	RunnerStatusConfigured = "configured"
	RunnerStatusRunning    = "running"
	RunnerStatusStopped    = "stopped"
	RunnerStatusError      = "error"
	RunnerStatusRemoving   = "removing"
	RunnerStatusRemoved    = "removed"

	RunnerTaskTypeCreate      = "create"
	RunnerTaskTypeStart       = "start"
	RunnerTaskTypeStop        = "stop"
	RunnerTaskTypeRemove      = "remove"
	RunnerTaskTypeReconfigure = "reconfigure"

	RunnerTaskStatusQueued    = "queued"
	RunnerTaskStatusClaimed   = "claimed"
	RunnerTaskStatusRunning   = "running"
	RunnerTaskStatusSucceeded = "succeeded"
	RunnerTaskStatusFailed    = "failed"
	RunnerTaskStatusCanceled  = "canceled"
)

type ServerNode struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	NodeKey       string     `gorm:"size:100;uniqueIndex" json:"nodeKey"`
	Name          string     `gorm:"size:255;index" json:"name"`
	Hostname      string     `gorm:"size:255;index" json:"hostname"`
	Roles         string     `gorm:"size:100" json:"roles"`
	OS            string     `gorm:"size:50;index" json:"os"`
	Arch          string     `gorm:"size:50" json:"arch"`
	Version       string     `gorm:"size:100" json:"version"`
	Status        string     `gorm:"size:50;default:offline;index" json:"status"`
	ResourceJSON  string     `gorm:"type:text" json:"resourceJson"`
	LastHeartbeat *time.Time `json:"lastHeartbeat"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type AgentToken struct {
	ID                  uint       `gorm:"primaryKey" json:"id"`
	Name                string     `gorm:"size:255;index" json:"name"`
	TokenHash           string     `gorm:"size:255;uniqueIndex" json:"-"`
	LastUsedAt          *time.Time `json:"lastUsedAt"`
	ExpiresAt           *time.Time `json:"expiresAt"`
	RevokedAt           *time.Time `json:"revokedAt"`
	CreaterID           uint       `json:"createrId"`
	CreaterUserName     string     `gorm:"size:255" json:"createrUserName"`
	LastUpdaterID       uint       `json:"lastUpdaterId"`
	LastUpdaterUserName string     `gorm:"size:255" json:"lastUpdaterUserName"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

type Runner struct {
	ID                  uint       `gorm:"primaryKey" json:"id"`
	ProjectID           uint       `gorm:"index" json:"projectId"`
	NodeID              uint       `gorm:"index" json:"nodeId"`
	Name                string     `gorm:"size:255;index" json:"name"`
	Labels              string     `gorm:"size:1000" json:"labels"`
	WorkDir             string     `gorm:"size:1000" json:"workDir"`
	InstallDir          string     `gorm:"size:1000" json:"installDir"`
	OS                  string     `gorm:"size:50;index" json:"os"`
	Arch                string     `gorm:"size:50" json:"arch"`
	Status              string     `gorm:"size:50;default:pending;index" json:"status"`
	ProcessID           int        `json:"processId"`
	LastError           string     `gorm:"type:text" json:"lastError"`
	LastSeenAt          *time.Time `json:"lastSeenAt"`
	CreaterID           uint       `json:"createrId"`
	CreaterUserName     string     `gorm:"size:255" json:"createrUserName"`
	LastUpdaterID       uint       `json:"lastUpdaterId"`
	LastUpdaterUserName string     `gorm:"size:255" json:"lastUpdaterUserName"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

type RunnerTask struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	RunnerID     uint       `gorm:"index" json:"runnerId"`
	NodeID       uint       `gorm:"index" json:"nodeId"`
	Type         string     `gorm:"size:50;index" json:"type"`
	Status       string     `gorm:"size:50;default:queued;index" json:"status"`
	PayloadJSON  string     `gorm:"type:text" json:"payloadJson"`
	ResultJSON   string     `gorm:"type:text" json:"resultJson"`
	LogSummary   string     `gorm:"type:text" json:"logSummary"`
	LastError    string     `gorm:"type:text" json:"lastError"`
	AttemptCount int        `json:"attemptCount"`
	ClaimedAt    *time.Time `json:"claimedAt"`
	StartedAt    *time.Time `json:"startedAt"`
	FinishedAt   *time.Time `json:"finishedAt"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}
