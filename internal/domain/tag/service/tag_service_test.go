package service

import (
	"context"
	"testing"

	"group-buy-market/internal/domain/tag/model/entity"
)

type mockTagRepo struct {
	job    *entity.CrowdTagsJobEntity
	users  []string
	stats  int
}

func (m *mockTagRepo) QueryCrowdTagsJobEntity(ctx context.Context, tagID, batchID string) (*entity.CrowdTagsJobEntity, error) {
	return m.job, nil
}
func (m *mockTagRepo) AddCrowdTagsUserID(ctx context.Context, tagID, userID string) error {
	m.users = append(m.users, userID)
	return nil
}
func (m *mockTagRepo) UpdateCrowdTagsStatistics(ctx context.Context, tagID string, count int) error {
	m.stats = count
	return nil
}

func TestExecTagBatchJob(t *testing.T) {
	repo := &mockTagRepo{job: &entity.CrowdTagsJobEntity{TagType: 0, TagRule: "100"}}
	svc := NewTagService(repo)
	if err := svc.ExecTagBatchJob(context.Background(), "tag", "10001"); err != nil {
		t.Fatal(err)
	}
	if len(repo.users) == 0 || repo.stats != len(repo.users) {
		t.Fatalf("users=%d stats=%d", len(repo.users), repo.stats)
	}
}

func TestExecTagBatchJobMissing(t *testing.T) {
	repo := &mockTagRepo{job: nil}
	svc := NewTagService(repo)
	if err := svc.ExecTagBatchJob(context.Background(), "tag", "x"); err != nil {
		t.Fatal(err)
	}
	if len(repo.users) != 0 {
		t.Fatal("should not write")
	}
}
