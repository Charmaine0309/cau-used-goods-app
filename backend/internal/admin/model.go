package admin

import "time"

const (
	TargetTypeUser     = "USER"
	TargetTypeProduct  = "PRODUCT"
	TargetTypeOrder    = "ORDER"
	TargetTypeReport   = "REPORT"
	TargetTypeNotice   = "NOTICE"
	TargetTypeWord     = "WORD"
	TargetTypeAppeal   = "APPEAL"
	TargetTypeCategory = "CATEGORY"
)

const (
	AnnouncementStatusDraft     = "DRAFT"
	AnnouncementStatusPublished = "PUBLISHED"
	AnnouncementStatusOffline   = "OFFLINE"
)

const (
	OperationUserDisable         = "USER_DISABLE"
	OperationUserEnable          = "USER_ENABLE"
	OperationProductOffShelf     = "PRODUCT_OFF_SHELF"
	OperationReportResolve       = "REPORT_RESOLVE"
	OperationReportReject        = "REPORT_REJECT"
	OperationReportClose         = "REPORT_CLOSE"
	OperationNoticePublish       = "NOTICE_PUBLISH"
	OperationNoticeOffline       = "NOTICE_OFFLINE"
	OperationWordCreate          = "WORD_CREATE"
	OperationWordDisable         = "WORD_DISABLE"
	OperationOrderExceptionClose = "ORDER_EXCEPTION_CLOSE"
	OperationCategoryCreate      = "CATEGORY_CREATE"
	OperationCategoryUpdate      = "CATEGORY_UPDATE"
	OperationCategoryEnable      = "CATEGORY_ENABLE"
	OperationCategoryDisable     = "CATEGORY_DISABLE"

	OperationCreateNotice = "CREATE_NOTICE"
	OperationUpdateNotice = "UPDATE_NOTICE"
	OperationStatusNotice = "STATUS_NOTICE"
	OperationUpdateWord   = "UPDATE_WORD"
	OperationHandleAppeal = "HANDLE_APPEAL"
)

type AdminLog struct {
	ID            uint64    `json:"id"`
	AdminID       uint64    `json:"adminId"`
	OperationType string    `json:"operationType"`
	TargetType    string    `json:"targetType"`
	TargetID      uint64    `json:"targetId"`
	Description   *string   `json:"description,omitempty"`
	IPAddress     *string   `json:"ipAddress,omitempty"`
	CreateTime    time.Time `json:"createTime"`
}

type LogActionInput struct {
	AdminID       uint64
	OperationType string
	TargetType    string
	TargetID      uint64
	Description   *string
	IPAddress     *string
}

type LogQuery struct {
	AdminID       uint64
	OperationType string
	TargetType    string
	TargetID      uint64
	StartTime     string
	EndTime       string
	Page          int
	PageSize      int
}

type Announcement struct {
	ID          uint64     `json:"id"`
	Title       string     `json:"title"`
	Content     *string    `json:"content,omitempty"`
	CoverURL    *string    `json:"coverUrl,omitempty"`
	Status      string     `json:"status"`
	PublishTime *time.Time `json:"publishTime,omitempty"`
	CreateBy    uint64     `json:"createBy"`
	CreateTime  time.Time  `json:"createTime"`
	UpdateTime  time.Time  `json:"updateTime"`
}

type AnnouncementQuery struct {
	Status   string
	Keyword  string
	Page     int
	PageSize int
}

type CreateAnnouncementInput struct {
	AdminID   uint64
	Title     string
	Content   *string
	CoverURL  *string
	Status    string
	IPAddress *string
}

type UpdateAnnouncementInput struct {
	AdminID   uint64
	ID        uint64
	Title     string
	Content   *string
	CoverURL  *string
	IPAddress *string
}

type UpdateAnnouncementStatusInput struct {
	AdminID   uint64
	ID        uint64
	Status    string
	IPAddress *string
}
