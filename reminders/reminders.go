package reminders

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/mrbentarikau/pagst/bot"
	"github.com/mrbentarikau/pagst/common"
	"github.com/mrbentarikau/pagst/common/mqueue"
	"github.com/mrbentarikau/pagst/common/scheduledevents2"
	"github.com/mrbentarikau/pagst/lib/discordgo"
	"github.com/mrbentarikau/pagst/reminders/models"

	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type ModalData struct {
	AuthorID  int64
	ChannelID int64
	Content   string
	GuildID   int64
	MessageID int64
}

//go:generate sqlboiler --no-hooks --add-soft-deletes psql

type Plugin struct {
	RemindmeData ModalData
}

func RegisterPlugin() {
	p := &Plugin{}
	common.RegisterPlugin(p)

	common.InitSchemas("reminders", DBSchemas...)
}

func (p *Plugin) PluginInfo() *common.PluginInfo {
	return &common.PluginInfo{
		Name:     "Reminders",
		SysName:  "reminders",
		Category: common.PluginCategoryMisc,
	}
}

func TriggerReminder(r *models.Reminder) error {
	reminderRepeat := "**Reminder** for"
	userID, _ := discordgo.ParseID(r.UserID)
	if r.Repeat > 0 {
		reminderRepeat = "**Repeated reminder** for"
		repeatDuration := time.Duration(r.Repeat) * time.Nanosecond
		when := time.Now().Add(repeatDuration)
		r.When = when.Unix()

		_, err := r.UpdateG(context.Background(), boil.Infer())
		if err != nil {
			return err
		}
		err = scheduledevents2.ScheduleEvent("reminders_check_user", r.GuildID, when, userID)
		if err != nil {
			logger.Info("Reminders repeat scheduler:", err)
		}
	} else {
		// remove the actual reminder
		r.DeleteG(context.Background(), false /* hardDelete */)
	}

	logger.WithFields(logrus.Fields{"channel": r.ChannelID, "user": r.UserID, "message": r.Message, "id": r.ID}).Info("Triggered reminder")
	member, _ := bot.GetMember(r.GuildID, userID)
	if member != nil {
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Reminder #%d from %s", r.ID, common.ConfBotName.GetString()),
			Description: common.ReplaceServerInvites(r.Message, r.GuildID, "(removed-invite)"),
			Color:       int(rand.Int63n(16777215)),
			Timestamp:   time.Now().Format(time.RFC3339),
		}

		var sendDM bool
		channelID, _ := discordgo.ParseID(r.ChannelID)
		if r.AppCommand {
			sendDM = true
			channel, err := common.BotSession.UserChannelCreate(userID)
			if err != nil {
				return err
			}

			msgSend := &discordgo.MessageSend{
				AllowedMentions: discordgo.AllowedMentions{
					Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeUsers},
				},
			}

			gs := bot.State.GetGuild(r.GuildID)
			gIcon := discordgo.EndpointGuildIcon(gs.ID, gs.Icon)
			embedInfo := fmt.Sprintf("DM from the server %s", gs.Name)

			embed.Footer = &discordgo.MessageEmbedFooter{
				Text:    embedInfo,
				IconURL: gIcon,
			}

			msgSend.Embeds = []*discordgo.MessageEmbed{embed}

			_, err = common.BotSession.ChannelMessageSendComplex(channel.ID, msgSend)
			if err != nil {
				embed.Footer = &discordgo.MessageEmbedFooter{}
				sendDM = false
			}
		}

		disabledMentions := discordgo.AllowedMentions{Users: []int64{userID}}
		if r.DisableMention {
			disabledMentions = discordgo.AllowedMentions{}
		}

		if !sendDM {
			mqueue.QueueMessage(&mqueue.QueuedElement{
				Source:       "reminder",
				SourceItemID: "",

				GuildID:         r.GuildID,
				ChannelID:       channelID,
				MessageEmbed:    embed,
				MessageStr:      reminderRepeat + " <@" + r.UserID + ">", // : " + common.ReplaceServerInvites(r.Message, r.GuildID, "(removed-invite)"),
				AllowedMentions: disabledMentions,

				Priority: 10, // above all feeds
			})
		}
	}
	return nil
}

func NewReminder(userID int64, guildID int64, channelID int64, message string, when time.Time, repeatDuration time.Duration, disableMention bool, isAppCmd ...bool) (*models.Reminder, error) {
	var appCmd bool

	if len(isAppCmd) > 0 {
		appCmd = true
	}

	reminder := &models.Reminder{
		UserID:         discordgo.StrID(userID),
		ChannelID:      discordgo.StrID(channelID),
		Message:        message,
		When:           when.Unix(),
		GuildID:        guildID,
		Repeat:         repeatDuration.Nanoseconds(),
		AppCommand:     appCmd,
		DisableMention: disableMention,
	}

	err := reminder.InsertG(context.Background(), boil.Infer())
	if err != nil {
		return nil, err
	}

	err = scheduledevents2.ScheduleEvent("reminders_check_user", guildID, when, userID)
	// err = scheduledevents.ScheduleEvent("reminders_check_user:"+strconv.FormatInt(whenUnix, 10), discordgo.StrID(userID), when)
	return reminder, err
}

type DisplayRemindersMode int

const (
	ModeDisplayChannelReminders DisplayRemindersMode = iota
	ModeDisplayUserReminders
)

func DisplayReminders(reminders models.ReminderSlice, mode DisplayRemindersMode) string {
	var out strings.Builder
	for _, r := range reminders {
		t := time.Unix(r.When, 0)
		timeFromNow := common.HumanizeTime(common.DurationPrecisionMinutes, t)

		switch mode {
		case ModeDisplayChannelReminders:
			// don't show the channel; do show the user
			uid, _ := discordgo.ParseID(r.UserID)
			member, _ := bot.GetMember(r.GuildID, uid)
			username := "Unknown user"
			if member != nil {
				username = member.User.Username
			}

			fmt.Fprintf(&out, "**%d**: %s: '%s' - %s from now (<t:%d:f>)\n", r.ID, username, CutReminderShort(r.Message), timeFromNow, t.Unix())

		case ModeDisplayUserReminders:
			// do show the channel; don't show the user
			channel := "<#" + r.ChannelID + ">"
			fmt.Fprintf(&out, "**%d**: %s: '%s' - %s from now (<t:%d:f>)\n", r.ID, channel, CutReminderShort(r.Message), timeFromNow, t.Unix())
		}
	}

	return out.String()
}

func CutReminderShort(msg string) string {
	return common.CutStringShort(msg, 50)
}
