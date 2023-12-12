package impl_protected

import (
	"github.com/alexedwards/scs/v2"
	"github.com/redis/go-redis/v9"
	"github.com/satont/twir/apps/api/internal/impl_deps"
	"github.com/satont/twir/apps/api/internal/impl_protected/alerts"
	"github.com/satont/twir/apps/api/internal/impl_protected/auth"
	"github.com/satont/twir/apps/api/internal/impl_protected/bot"
	"github.com/satont/twir/apps/api/internal/impl_protected/build_in_variables"
	"github.com/satont/twir/apps/api/internal/impl_protected/commands"
	"github.com/satont/twir/apps/api/internal/impl_protected/commands_group"
	"github.com/satont/twir/apps/api/internal/impl_protected/community"
	"github.com/satont/twir/apps/api/internal/impl_protected/dashboard"
	"github.com/satont/twir/apps/api/internal/impl_protected/events"
	"github.com/satont/twir/apps/api/internal/impl_protected/files"
	"github.com/satont/twir/apps/api/internal/impl_protected/games"
	"github.com/satont/twir/apps/api/internal/impl_protected/giveaways"
	"github.com/satont/twir/apps/api/internal/impl_protected/greetings"
	"github.com/satont/twir/apps/api/internal/impl_protected/integrations"
	"github.com/satont/twir/apps/api/internal/impl_protected/keywords"
	"github.com/satont/twir/apps/api/internal/impl_protected/moderation"
	"github.com/satont/twir/apps/api/internal/impl_protected/modules"
	"github.com/satont/twir/apps/api/internal/impl_protected/overlays"
	"github.com/satont/twir/apps/api/internal/impl_protected/rewards"
	"github.com/satont/twir/apps/api/internal/impl_protected/roles"
	"github.com/satont/twir/apps/api/internal/impl_protected/timers"
	"github.com/satont/twir/apps/api/internal/impl_protected/twitch"
	"github.com/satont/twir/apps/api/internal/impl_protected/users"
	"github.com/satont/twir/apps/api/internal/impl_protected/variables"
	config "github.com/satont/twir/libs/config"
	"github.com/satont/twir/libs/grpc/generated/bots"
	"github.com/satont/twir/libs/grpc/generated/discord"
	giveawaysGrpc "github.com/satont/twir/libs/grpc/generated/giveaways"
	integrationsGrpc "github.com/satont/twir/libs/grpc/generated/integrations"
	"github.com/satont/twir/libs/grpc/generated/parser"
	"github.com/satont/twir/libs/grpc/generated/scheduler"
	timersGrpc "github.com/satont/twir/libs/grpc/generated/timers"
	"github.com/satont/twir/libs/grpc/generated/tokens"
	"github.com/satont/twir/libs/grpc/generated/websockets"
	"github.com/satont/twir/libs/logger"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type Protected struct {
	*integrations.Integrations
	*keywords.Keywords
	*modules.Modules
	*bot.Bot
	*commands.Commands
	*community.Community
	*events.Events
	*greetings.Greetings
	*rewards.Rewards
	*roles.Roles
	*timers.Timers
	*auth.Auth
	*variables.Variables
	*commands_group.CommandsGroup
	*build_in_variables.BuildInVariables
	*dashboard.Dashboard
	*twitch.Twitch
	*files.Files
	*alerts.Alerts
	*games.Games
	*overlays.Overlays
	*moderation.Moderation
	*users.Users
	*giveaways.Giveaways
}

type Opts struct {
	fx.In

	Redis          *redis.Client
	DB             *gorm.DB
	Config         config.Config
	SessionManager *scs.SessionManager

	TokensGrpc       tokens.TokensClient
	BotsGrpc         bots.BotsClient
	IntegrationsGrpc integrationsGrpc.IntegrationsClient
	ParserGrpc       parser.ParserClient
	WebsocketsGrpc   websockets.WebsocketClient
	SchedulerGrpc    scheduler.SchedulerClient
	TimersGrpc       timersGrpc.TimersClient
	DiscordGrpc      discord.DiscordClient
	Logger           logger.Logger
	GiveawaysGrpc    giveawaysGrpc.GiveawaysClient
}

func New(opts Opts) *Protected {
	d := &impl_deps.Deps{
		Redis:          opts.Redis,
		Db:             opts.DB,
		Config:         opts.Config,
		SessionManager: opts.SessionManager,
		Grpc: &impl_deps.Grpc{
			Tokens:       opts.TokensGrpc,
			Bots:         opts.BotsGrpc,
			Integrations: opts.IntegrationsGrpc,
			Parser:       opts.ParserGrpc,
			Websockets:   opts.WebsocketsGrpc,
			Scheduler:    opts.SchedulerGrpc,
			Timers:       opts.TimersGrpc,
			Discord:      opts.DiscordGrpc,
			Giveaways:    opts.GiveawaysGrpc,
		},
		Logger: opts.Logger,
	}

	return &Protected{
		Integrations:     &integrations.Integrations{Deps: d},
		Keywords:         &keywords.Keywords{Deps: d},
		Modules:          &modules.Modules{Deps: d},
		Bot:              &bot.Bot{Deps: d},
		Commands:         &commands.Commands{Deps: d},
		Community:        &community.Community{Deps: d},
		Events:           &events.Events{Deps: d},
		Greetings:        &greetings.Greetings{Deps: d},
		Rewards:          &rewards.Rewards{Deps: d},
		Roles:            &roles.Roles{Deps: d},
		Timers:           &timers.Timers{Deps: d},
		Auth:             &auth.Auth{Deps: d},
		Variables:        &variables.Variables{Deps: d},
		CommandsGroup:    &commands_group.CommandsGroup{Deps: d},
		BuildInVariables: &build_in_variables.BuildInVariables{Deps: d},
		Dashboard:        &dashboard.Dashboard{Deps: d},
		Twitch:           &twitch.Twitch{Deps: d},
		Files:            files.New(d),
		Alerts:           &alerts.Alerts{Deps: d},
		Games:            &games.Games{Deps: d},
		Overlays:         &overlays.Overlays{Deps: d},
		Moderation:       &moderation.Moderation{Deps: d},
		Users:            &users.Users{Deps: d},
		Giveaways:        &giveaways.Giveaways{Deps: d},
	}
}
