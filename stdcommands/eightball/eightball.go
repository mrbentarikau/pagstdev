// using code from https://github.com/xbilldozer/8ball

package eightball

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/botlabs-gg/yagpdb/commands"
	"github.com/botlabs-gg/yagpdb/lib/dcmd"
)

var magicAnswers = [...]string{
	// Positive outcomes
	"It is certain",
	"It is decidedly so",
	"Without a doubt",
	"Yes definitely",
	"You may rely on it",
	"As I see it, yes",
	"Most likely",
	"Outlook good",
	"Yes",
	"Signs point to yes",

	// Neutral outcomes
	"Reply hazy try again",
	"Ask again later",
	"Better not tell you now",
	"Cannot predict now",
	"Concentrate and ask again",

	// Negative outcomes
	"Don't count on it",
	"My reply is no",
	"My sources say no",
	"Outlook not so good",
	"Very doubtful",
}

var ShakeFailureMessage = "You can't just shake me bro, ask a question!"

var Command = &commands.YAGCommand{
	Cooldown:                  2,
	CmdCategory:               commands.CategoryFun,
	Name:                      "8Ball",
	Description:               "Wisdom",
	DefaultEnabled:            true,
	ApplicationCommandEnabled: true,
	Arguments: []*dcmd.ArgDef{
		{Name: "What-to-ask", Type: dcmd.String},
	},

	RunFunc: func(data *dcmd.Data) (interface{}, error) {
		if data.SlashCommandTriggerData != nil {
			return "You asked: " + data.Args[0].Str() + "\n" + Shake(data), nil
		}
		return Shake(data), nil
	},
}

func Shake(data *dcmd.Data) string {
	if data.Args[0].Str() == "" {
		return ShakeFailureMessage
	}

	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("8-Ball says: %s", magicAnswers[rand.Intn(len(magicAnswers))])
}
