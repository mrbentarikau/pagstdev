package stdcommands

import (
	"github.com/botlabs-gg/yagpdb/bot"
	"github.com/botlabs-gg/yagpdb/bot/eventsystem"
	"github.com/botlabs-gg/yagpdb/commands"
	"github.com/botlabs-gg/yagpdb/common"
	"github.com/botlabs-gg/yagpdb/stdcommands/addspecialserver"
	"github.com/botlabs-gg/yagpdb/stdcommands/advice"
	"github.com/botlabs-gg/yagpdb/stdcommands/allocstat"
	"github.com/botlabs-gg/yagpdb/stdcommands/antifish"
	"github.com/botlabs-gg/yagpdb/stdcommands/banserver"
	"github.com/botlabs-gg/yagpdb/stdcommands/banserverowner"
	"github.com/botlabs-gg/yagpdb/stdcommands/bashquotes"
	"github.com/botlabs-gg/yagpdb/stdcommands/calc"
	"github.com/botlabs-gg/yagpdb/stdcommands/catfact"
	"github.com/botlabs-gg/yagpdb/stdcommands/ccreqs"
	"github.com/botlabs-gg/yagpdb/stdcommands/covidstats"
	"github.com/botlabs-gg/yagpdb/stdcommands/createinvite"
	"github.com/botlabs-gg/yagpdb/stdcommands/currentshard"
	"github.com/botlabs-gg/yagpdb/stdcommands/currenttime"
	"github.com/botlabs-gg/yagpdb/stdcommands/customembed"
	"github.com/botlabs-gg/yagpdb/stdcommands/dadjoke"
	"github.com/botlabs-gg/yagpdb/stdcommands/dcallvoice"
	"github.com/botlabs-gg/yagpdb/stdcommands/define"
	"github.com/botlabs-gg/yagpdb/stdcommands/dogfact"
	"github.com/botlabs-gg/yagpdb/stdcommands/editchannelpermissions"
	"github.com/botlabs-gg/yagpdb/stdcommands/editrole"
	"github.com/botlabs-gg/yagpdb/stdcommands/exportcustomcommands"
	"github.com/botlabs-gg/yagpdb/stdcommands/exportuserdatabase"
	"github.com/botlabs-gg/yagpdb/stdcommands/findserver"
	"github.com/botlabs-gg/yagpdb/stdcommands/getiplocation"
	"github.com/botlabs-gg/yagpdb/stdcommands/globalrl"
	"github.com/botlabs-gg/yagpdb/stdcommands/guildunavailable"
	"github.com/botlabs-gg/yagpdb/stdcommands/howlongtobeat"
	"github.com/botlabs-gg/yagpdb/stdcommands/info"
	"github.com/botlabs-gg/yagpdb/stdcommands/inspire"
	"github.com/botlabs-gg/yagpdb/stdcommands/invite"
	"github.com/botlabs-gg/yagpdb/stdcommands/leaveserver"
	"github.com/botlabs-gg/yagpdb/stdcommands/listflags"
	"github.com/botlabs-gg/yagpdb/stdcommands/listroles"
	"github.com/botlabs-gg/yagpdb/stdcommands/memstats"
	"github.com/botlabs-gg/yagpdb/stdcommands/openweathermap"
	"github.com/botlabs-gg/yagpdb/stdcommands/owldictionary"
	"github.com/botlabs-gg/yagpdb/stdcommands/pagstatus"
	"github.com/botlabs-gg/yagpdb/stdcommands/ping"
	"github.com/botlabs-gg/yagpdb/stdcommands/poll"
	"github.com/botlabs-gg/yagpdb/stdcommands/quack"
	"github.com/botlabs-gg/yagpdb/stdcommands/removespecialserver"
	"github.com/botlabs-gg/yagpdb/stdcommands/roll"
	"github.com/botlabs-gg/yagpdb/stdcommands/setstatus"
	"github.com/botlabs-gg/yagpdb/stdcommands/simpleembed"
	"github.com/botlabs-gg/yagpdb/stdcommands/sleep"
	"github.com/botlabs-gg/yagpdb/stdcommands/statedbg"
	"github.com/botlabs-gg/yagpdb/stdcommands/stateinfo"
	"github.com/botlabs-gg/yagpdb/stdcommands/throw"
	"github.com/botlabs-gg/yagpdb/stdcommands/toggledbg"
	"github.com/botlabs-gg/yagpdb/stdcommands/topcommands"
	"github.com/botlabs-gg/yagpdb/stdcommands/topevents"
	"github.com/botlabs-gg/yagpdb/stdcommands/topgames"
	"github.com/botlabs-gg/yagpdb/stdcommands/topic"
	"github.com/botlabs-gg/yagpdb/stdcommands/topservers"
	"github.com/botlabs-gg/yagpdb/stdcommands/unbanserver"
	"github.com/botlabs-gg/yagpdb/stdcommands/undelete"
	"github.com/botlabs-gg/yagpdb/stdcommands/viewperms"
	"github.com/botlabs-gg/yagpdb/stdcommands/voidservers"
	"github.com/botlabs-gg/yagpdb/stdcommands/weather"
	"github.com/botlabs-gg/yagpdb/stdcommands/wolframalpha"
	"github.com/botlabs-gg/yagpdb/stdcommands/wouldyourather"
	"github.com/botlabs-gg/yagpdb/stdcommands/xkcd"
)

var (
	_ bot.BotInitHandler       = (*Plugin)(nil)
	_ commands.CommandProvider = (*Plugin)(nil)
)

type Plugin struct{}

func (p *Plugin) PluginInfo() *common.PluginInfo {
	return &common.PluginInfo{
		Name:     "Standard Commands",
		SysName:  "standard_commands",
		Category: common.PluginCategoryCore,
	}
}

func (p *Plugin) AddCommands() {
	commands.AddRootCommands(p,
		// Info
		info.Command,
		invite.Command,

		// Standard
		advice.Command,
		bashquotes.Command,
		calc.Command,
		catfact.Command,
		covidstats.Command,
		currenttime.Command,
		customembed.Command,
		dadjoke.Command,
		define.Command,
		dogfact.Command,
		editchannelpermissions.Command,
		editrole.Command,
		getiplocation.Command,
		howlongtobeat.Command,
		inspire.Command,
		listroles.Command,
		openweathermap.Command,
		ping.Command,
		poll.Command,
		quack.Command,
		roll.Command,
		simpleembed.Command,
		throw.Command,
		topgames.Command,
		topic.Command,
		undelete.Command,
		viewperms.Command,
		weather.Command,
		wolframalpha.Command,
		wouldyourather.Command,
		xkcd.Command,

		// Maintenance
		antifish.Command,
		addspecialserver.Command,
		allocstat.Command,
		banserver.Command,
		banserverowner.Command,
		ccreqs.Command,
		createinvite.Command,
		currentshard.Command,
		dcallvoice.Command,
		exportcustomcommands.Command,
		exportuserdatabase.Command,
		findserver.Command,
		globalrl.Command,
		guildunavailable.Command,
		leaveserver.Command,
		listflags.Command,
		memstats.Command,
		pagstatus.Command,
		removespecialserver.Command,
		setstatus.Command,
		sleep.Command,
		stateinfo.Command,
		toggledbg.Command,
		topcommands.Command,
		topevents.Command,
		topservers.Command,
		unbanserver.Command,
		voidservers.Command,
	)

	statedbg.Commands()

	if !owldictionary.ShouldRegister() {
		common.GetPluginLogger(p).Warn("Owlbot API token not provided, skipping adding dictionary command...")
		return
	}

	commands.AddRootCommands(p, owldictionary.Command)

}

func (p *Plugin) BotInit() {
	eventsystem.AddHandlerAsyncLastLegacy(p, ping.HandleMessageCreate, eventsystem.EventMessageCreate)
}

func RegisterPlugin() {
	common.RegisterPlugin(&Plugin{})
}
