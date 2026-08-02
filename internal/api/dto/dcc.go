package dto

// DCCUpdateRequestDTO 动态配置更新（JSON 体）
type DCCUpdateRequestDTO struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// DCCSnapshot 动态配置快照
type DCCSnapshot map[string]string
