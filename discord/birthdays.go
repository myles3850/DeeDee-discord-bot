package discord

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	// Embeds the timezone database: the container is built FROM scratch, so
	// there's no /usr/share/zoneinfo for time.LoadLocation to read.
	_ "time/tzdata"
)

// StartBirthdayReminders posts a birthday message to birthdayChannelID at
// birthdayReminderHour every day. It blocks until ctx is cancelled, so run it
// in a goroutine.
func (d *Discord) StartBirthdayReminders(ctx context.Context) {
	loc, err := time.LoadLocation(birthdayTimezone)
	if err != nil {
		slog.Error("loading birthday timezone, birthday reminders disabled", "timezone", birthdayTimezone, "error", err)
		return
	}

	for {
		next := nextBirthdayReminder(time.Now().In(loc))
		slog.Info("next birthday check scheduled", "at", next)

		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			if _, err := d.sendBirthdayReminders(next); err != nil {
				slog.Error("sending birthday reminders", "error", err)
			}
		}
	}
}

// nextBirthdayReminder returns the next birthdayReminderHour strictly after
// now, in now's location. Building it with time.Date (rather than adding 24h)
// keeps it on the right wall-clock hour across DST changes.
func nextBirthdayReminder(now time.Time) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), birthdayReminderHour, 0, 0, 0, now.Location())
	if !next.After(now) {
		next = time.Date(now.Year(), now.Month(), now.Day()+1, birthdayReminderHour, 0, 0, 0, now.Location())
	}
	return next
}

// sendBirthdayReminders posts one happy birthday message mentioning everyone
// whose birthday is today, returning how many people it mentioned. Nothing is
// posted if there are no birthdays.
func (d *Discord) sendBirthdayReminders(today time.Time) (int, error) {
	userIDs, err := d.Database.GetUsersWithBirthdayOn(today)
	if err != nil {
		return 0, fmt.Errorf("fetching today's birthdays: %w", err)
	}
	if len(userIDs) == 0 {
		return 0, nil
	}

	mentions := make([]string, len(userIDs))
	for i, id := range userIDs {
		mentions[i] = fmt.Sprintf("<@%s>", id)
	}

	message := fmt.Sprintf("🎂 Happy birthday %s! Hope you have the best day 🎉", strings.Join(mentions, ", "))
	if _, err := d.Session.ChannelMessageSend(birthdayChannelID, message); err != nil {
		return 0, fmt.Errorf("sending birthday message to %s: %w", birthdayChannelID, err)
	}
	return len(userIDs), nil
}
