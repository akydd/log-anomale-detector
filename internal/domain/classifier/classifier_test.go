package classifier

import (
	"akydd/log-anomale-detector/internal/domain"
	"context"
	"errors"
	"testing"
	"time"
)

// -- mocks --

type mockAI struct {
	results []domain.ClassifiedLogs
	err     error
}

func (m *mockAI) Classify(_ context.Context, _ []string) ([]domain.ClassifiedLogs, error) {
	return m.results, m.err
}

type mockDB struct {
	calls []domain.ClassifiedLogs
	err   error
}

func (m *mockDB) Put(_ context.Context, logs domain.ClassifiedLogs) error {
	m.calls = append(m.calls, logs)
	return m.err
}

type mockNotifier struct {
	calls []string
	err   error
}

func (m *mockNotifier) Notify(_ context.Context, msg string) error {
	m.calls = append(m.calls, msg)
	return m.err
}

// -- helpers --

func strPtr(s string) *string { return &s }

func flaggedResult(severity string) domain.ClassifiedLogs {
	now := time.Now()
	return domain.ClassifiedLogs{
		Flag:               true,
		Timestamp:          &now,
		AnomalyType:        strPtr("AUTH_FAILURE"),
		Severity:           &severity,
		RawEvidence:        []string{"failed login from 1.2.3.4"},
		BedrockExplanation: strPtr("Multiple failed auth attempts detected"),
	}
}

// -- tests --

func TestClassify_NoResults(t *testing.T) {
	ai := &mockAI{results: []domain.ClassifiedLogs{}}
	db := &mockDB{}
	n := &mockNotifier{}

	c := New(db, n, ai)
	results, err := c.Classify(context.Background(), []string{"normal log"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
	if len(db.calls) != 0 {
		t.Errorf("expected no DB writes, got %d", len(db.calls))
	}
	if len(n.calls) != 0 {
		t.Errorf("expected no notifications, got %d", len(n.calls))
	}
}

func TestClassify_UnflaggedResultsNotPersisted(t *testing.T) {
	unflagged := domain.ClassifiedLogs{Flag: false, Severity: strPtr("LOW")}
	ai := &mockAI{results: []domain.ClassifiedLogs{unflagged}}
	db := &mockDB{}
	n := &mockNotifier{}

	c := New(db, n, ai)
	_, err := c.Classify(context.Background(), []string{"normal log"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(db.calls) != 0 {
		t.Errorf("expected no DB writes for unflagged result, got %d", len(db.calls))
	}
	if len(n.calls) != 0 {
		t.Errorf("expected no notifications for unflagged result, got %d", len(n.calls))
	}
}

func TestClassify_FlaggedLowSeverity_PersistsButDoesNotNotify(t *testing.T) {
	ai := &mockAI{results: []domain.ClassifiedLogs{flaggedResult("LOW")}}
	db := &mockDB{}
	n := &mockNotifier{}

	c := New(db, n, ai)
	_, err := c.Classify(context.Background(), []string{"some log"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(db.calls) != 1 {
		t.Errorf("expected 1 DB write, got %d", len(db.calls))
	}
	if len(n.calls) != 0 {
		t.Errorf("expected no notifications for LOW severity, got %d", len(n.calls))
	}
}

func TestClassify_FlaggedHighSeverity_PersistsAndNotifies(t *testing.T) {
	result := flaggedResult("HIGH")
	ai := &mockAI{results: []domain.ClassifiedLogs{result}}
	db := &mockDB{}
	n := &mockNotifier{}

	c := New(db, n, ai)
	_, err := c.Classify(context.Background(), []string{"some log"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(db.calls) != 1 {
		t.Errorf("expected 1 DB write, got %d", len(db.calls))
	}
	if len(n.calls) != 1 {
		t.Errorf("expected 1 notification, got %d", len(n.calls))
	}
	if n.calls[0] != *result.BedrockExplanation {
		t.Errorf("notification message = %q, want %q", n.calls[0], *result.BedrockExplanation)
	}
}

func TestClassify_AIError_ReturnsError(t *testing.T) {
	ai := &mockAI{err: errors.New("bedrock unavailable")}
	db := &mockDB{}
	n := &mockNotifier{}

	c := New(db, n, ai)
	_, err := c.Classify(context.Background(), []string{"some log"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClassify_DBError_ReturnsError(t *testing.T) {
	ai := &mockAI{results: []domain.ClassifiedLogs{flaggedResult("LOW")}}
	db := &mockDB{err: errors.New("dynamodb timeout")}
	n := &mockNotifier{}

	c := New(db, n, ai)
	_, err := c.Classify(context.Background(), []string{"some log"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(n.calls) != 0 {
		t.Errorf("expected no notifications after DB failure, got %d", len(n.calls))
	}
}

func TestClassify_NotifyError_ReturnsError(t *testing.T) {
	ai := &mockAI{results: []domain.ClassifiedLogs{flaggedResult("HIGH")}}
	db := &mockDB{}
	n := &mockNotifier{err: errors.New("sns unavailable")}

	c := New(db, n, ai)
	_, err := c.Classify(context.Background(), []string{"some log"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
