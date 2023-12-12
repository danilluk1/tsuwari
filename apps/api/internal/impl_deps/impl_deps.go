package impl_deps

import (
	"github.com/alexedwards/scs/v2"
	"github.com/redis/go-redis/v9"
	cfg "github.com/satont/twir/libs/config"
	"github.com/satont/twir/libs/grpc/generated/bots"
	"github.com/satont/twir/libs/grpc/generated/discord"
	"github.com/satont/twir/libs/grpc/generated/eventsub"
	"github.com/satont/twir/libs/grpc/generated/giveaways"
	"github.com/satont/twir/libs/grpc/generated/integrations"
	"github.com/satont/twir/libs/grpc/generated/parser"
	"github.com/satont/twir/libs/grpc/generated/scheduler"
	"github.com/satont/twir/libs/grpc/generated/timers"
	"github.com/satont/twir/libs/grpc/generated/tokens"
	"github.com/satont/twir/libs/grpc/generated/websockets"
	"github.com/satont/twir/libs/logger"
	"gorm.io/gorm"
)

type Grpc struct {
	Tokens       tokens.TokensClient
	Integrations integrations.IntegrationsClient
	Bots         bots.BotsClient
	Parser       parser.ParserClient
	Websockets   websockets.WebsocketClient
	Scheduler    scheduler.SchedulerClient
	Timers       timers.TimersClient
	EventSub     eventsub.EventSubClient
	Discord      discord.DiscordClient
	Giveaways    giveaways.GiveawaysClient
}

type Deps struct {
	Config         cfg.Config
	Redis          *redis.Client
	Db             *gorm.DB
	Grpc           *Grpc
	SessionManager *scs.SessionManager
	Logger         logger.Logger
}
