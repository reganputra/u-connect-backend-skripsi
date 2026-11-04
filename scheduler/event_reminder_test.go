package scheduler

import (
	"context"
	"errors"
	"testing"
)

func TestReminderNotification(t *testing.T) {
	for _, tc := range []struct {
		name, title string
		user, event uint
	}{
		{"normal", "Alumni meetup", 7, 42},
		{"unicode", "Temu 🎓", 9, 99},
		{"empty title", "", 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			n := reminderNotification(reminderRow{UserID: tc.user, EventID: tc.event, EventTitle: tc.title})
			if n.ID != 0 || n.UserID != tc.user || n.ReferenceID != tc.event || n.NotificationType != "event_reminder" || n.Title != "Pengingat event" || n.Body != tc.title+" dimulai dalam 24 jam" || n.ReferenceType != "event" || n.RedirectURL == "" || n.IsRead {
				t.Fatalf("incorrect reminder: %+v", n)
			}
		})
	}
}

func TestCancelledSchedulersDoNotTouchDatabase(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := RunEventRemindersOnceContext(ctx, nil, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("reminders: %v", err)
	}
	if err := RunEventStatusSyncOnceContext(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("status: %v", err)
	}
	// Startup with an already-cancelled context must not run an initial DB job.
	StartEventReminderScheduler(ctx, nil, nil)
	StartEventStatusScheduler(ctx, nil)
}
