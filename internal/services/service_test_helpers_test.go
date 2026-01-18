package services

import "context"

type loggedEntry struct {
	action  string
	outcome string
	message *string
}

type stubLogWriter struct {
	entries []loggedEntry
}

func (s *stubLogWriter) CreateLog(ctx context.Context, action string, outcome string, message *string) error {
	var copied *string
	if message != nil {
		value := *message
		copied = &value
	}

	s.entries = append(s.entries, loggedEntry{
		action:  action,
		outcome: outcome,
		message: copied,
	})
	return nil
}
