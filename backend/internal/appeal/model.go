package appeal

const (
	TargetTypeProduct = "PRODUCT"
	TargetTypeUser    = "USER"
	TargetTypeOrder   = "ORDER"
	TargetTypeReport  = "REPORT"
)

const (
	StatusPending    = "PENDING"
	StatusProcessing = "PROCESSING"
	StatusApproved   = "APPROVED"
	StatusRejected   = "REJECTED"
	StatusClosed     = "CLOSED"
)

type Appeal struct {
	ID           uint64   `json:"id"`
	AppellantID  uint64   `json:"appellantId"`
	TargetType   string   `json:"targetType"`
	TargetID     uint64   `json:"targetId"`
	Reason       string   `json:"reason"`
	Status       string   `json:"status"`
	HandleResult *string  `json:"handleResult,omitempty"`
	HandlerID    *uint64  `json:"handlerId,omitempty"`
	HandleTime   *string  `json:"handleTime,omitempty"`
	CreateTime   string   `json:"createTime"`
	UpdateTime   string   `json:"updateTime"`
	Images       []string `json:"images,omitempty"`
}

type AppealDetail struct {
	Appeal
	AppellantNickname *string `json:"appellantNickname,omitempty"`
	HandlerNickname   *string `json:"handlerNickname,omitempty"`
}

type AppealQuery struct {
	AppellantID uint64
	TargetType  string
	TargetID    uint64
	Status      string
	Page        int
	PageSize    int
}

type CreateAppealInput struct {
	AppellantID  uint64
	TargetType   string
	TargetID     uint64
	Reason       string
	EvidenceURLs []string
}

type HandleAppealInput struct {
	AppealID     uint64
	AdminID      uint64
	Status       string
	HandleResult string
	IPAddress    *string
}
