package dto

// TagBatchJobRequestDTO 人群标签批次请求
type TagBatchJobRequestDTO struct {
	TagID   string `json:"tagId"`
	BatchID string `json:"batchId"`
}
