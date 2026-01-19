package services

import "context"

type loggedEntry struct {
	eventID *string
	action  string
	outcome string
	message *string
}

type stubLogWriter struct {
	entries []loggedEntry
}

func (s *stubLogWriter) CreateLog(ctx context.Context, eventID *string, action string, outcome string, message *string) error {
	var copied *string
	if message != nil {
		value := *message
		copied = &value
	}
	var copiedEventID *string
	if eventID != nil {
		value := *eventID
		copiedEventID = &value
	}

	s.entries = append(s.entries, loggedEntry{
		eventID: copiedEventID,
		action:  action,
		outcome: outcome,
		message: copied,
	})
	return nil
}
