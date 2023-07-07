package main

import (
	"github.com/botlabs-gg/yagpdb/analytics"
	"github.com/botlabs-gg/yagpdb/antiphishing"
	"github.com/botlabs-gg/yagpdb/common/featureflags"
	"github.com/botlabs-gg/yagpdb/common/prom"
	"github.com/botlabs-gg/yagpdb/common/run"
	"github.com/botlabs-gg/yagpdb/web/discorddata"

	// Core yagpdb packages

	"github.com/botlabs-gg/yagpdb/admin"
	"github.com/botlabs-gg/yagpdb/bot/paginatedmessages"
	"github.com/botlabs-gg/yagpdb/common/internalapi"
	"github.com/botlabs-gg/yagpdb/common/scheduledevents2"

	// Plugin imports
	"github.com/botlabs-gg/yagpdb/automod"
	"github.com/botlabs-gg/yagpdb/automod_basic"
	"github.com/botlabs-gg/yagpdb/autorole"

	// "github.com/botlabs-gg/yagpdb/aylien"
	"github.com/botlabs-gg/yagpdb/cah"
	"github.com/botlabs-gg/yagpdb/commands"
	"github.com/botlabs-gg/yagpdb/customcommands"
	"github.com/botlabs-gg/yagpdb/discordlogger"
	"github.com/botlabs-gg/yagpdb/logs"
	"github.com/botlabs-gg/yagpdb/moderation"
	"github.com/botlabs-gg/yagpdb/notifications"
	"github.com/botlabs-gg/yagpdb/premium"
	"github.com/botlabs-gg/yagpdb/premium/patreonpremiumsource"
	"github.com/botlabs-gg/yagpdb/reddit"
	"github.com/botlabs-gg/yagpdb/reminders"
	"github.com/botlabs-gg/yagpdb/reputation"
	"github.com/botlabs-gg/yagpdb/rolecommands"
	"github.com/botlabs-gg/yagpdb/rss"
	"github.com/botlabs-gg/yagpdb/rsvp"
	"github.com/botlabs-gg/yagpdb/safebrowsing"
	"github.com/botlabs-gg/yagpdb/serverstats"
	"github.com/botlabs-gg/yagpdb/soundboard"
	"github.com/botlabs-gg/yagpdb/stdcommands"
	"github.com/botlabs-gg/yagpdb/streaming"
	"github.com/botlabs-gg/yagpdb/tickets"
	"github.com/botlabs-gg/yagpdb/timezonecompanion"
	"github.com/botlabs-gg/yagpdb/twitter"
	"github.com/botlabs-gg/yagpdb/verification"
	"github.com/botlabs-gg/yagpdb/youtube"
	// External plugins
)

func main() {

	run.Init()

	//BotSession.LogLevel = discordgo.LogInformational
	paginatedmessages.RegisterPlugin()
	discorddata.RegisterPlugin()

	// Setup plugins
	admin.RegisterPlugin()
	analytics.RegisterPlugin()
	antiphishing.RegisterPlugin()
	automod.RegisterPlugin()
	automod_basic.RegisterPlugin()
	autorole.RegisterPlugin()
	// aylien.RegisterPlugin()
	cah.RegisterPlugin()
	commands.RegisterPlugin()
	customcommands.RegisterPlugin()
	discordlogger.Register()
	featureflags.RegisterPlugin()
	internalapi.RegisterPlugin()
	logs.RegisterPlugin()
	moderation.RegisterPlugin()
	notifications.RegisterPlugin()
	patreonpremiumsource.RegisterPlugin()
	premium.RegisterPlugin()
	prom.RegisterPlugin()
	reddit.RegisterPlugin()
	reminders.RegisterPlugin()
	reputation.RegisterPlugin()
	rolecommands.RegisterPlugin()
	rss.RegisterPlugin()
	rsvp.RegisterPlugin()
	safebrowsing.RegisterPlugin()
	scheduledevents2.RegisterPlugin()
	serverstats.RegisterPlugin()
	soundboard.RegisterPlugin()
	stdcommands.RegisterPlugin()
	streaming.RegisterPlugin()
	tickets.RegisterPlugin()
	timezonecompanion.RegisterPlugin()
	twitter.RegisterPlugin()
	verification.RegisterPlugin()
	youtube.RegisterPlugin()

	run.Run()
}
