package entity

import "time"

// CrowdTagsJobEntity 人群标签批次任务
type CrowdTagsJobEntity struct {
	TagType       int
	TagRule       string
	StatStartTime time.Time
	StatEndTime   time.Time
}
