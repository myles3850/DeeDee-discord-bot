package discord

import "time"

const (
	// newMemberChannelID is where new members land before posting an intro.
	newMemberChannelID = "1483226521014636762"
	// welcomeChannelID is where new members are welcomed after their intro.
	welcomeChannelID = "1483226521014636763"
	// modLogChannelID receives moderation notices (timeouts, deleted messages).
	modLogChannelID = "1483226520951455750"
	// modmailCategoryID is excluded from delete logging to avoid flooding the mod log.
	modmailCategoryID = "1483226520637018127"
	// botSpamRoleID, when self-assigned, times the member out via timeoutBotRole.
	botSpamRoleID = "1519435918224785458"
)

// botRoleTimeoutDuration is how long a member is timed out for picking botSpamRoleID.
const botRoleTimeoutDuration = 7 * 24 * time.Hour
