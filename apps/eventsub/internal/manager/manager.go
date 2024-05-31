package manager

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"sync"

	"github.com/go-redsync/redsync/v4"
	redsyncredis "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/google/uuid"
	"github.com/nicklaw5/helix/v2"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"github.com/satont/twir/apps/eventsub/internal/tunnel"
	cfg "github.com/satont/twir/libs/config"
	model "github.com/satont/twir/libs/gomodels"
	"github.com/satont/twir/libs/logger"
	"github.com/satont/twir/libs/twitch"
	"github.com/twirapp/twir/libs/grpc/tokens"
	eventsub_framework "github.com/twirapp/twitch-eventsub-framework"
	"go.uber.org/atomic"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type Manager struct {
	*eventsub_framework.SubClient

	config     cfg.Config
	logger     logger.Logger
	tokensGrpc tokens.TokensClient
	gorm       *gorm.DB
	tunnel     *tunnel.AppTunnel
}

type Opts struct {
	fx.In
	Lc fx.Lifecycle

	Config     cfg.Config
	Logger     logger.Logger
	Creds      *Creds
	TokensGrpc tokens.TokensClient
	Gorm       *gorm.DB
	Tunnel     *tunnel.AppTunnel
	Redis      *redis.Client
}

func NewManager(opts Opts) (*Manager, error) {
	client := eventsub_framework.NewSubClient(opts.Creds)

	manager := &Manager{
		SubClient:  client,
		config:     opts.Config,
		logger:     opts.Logger,
		tokensGrpc: opts.TokensGrpc,
		gorm:       opts.Gorm,
		tunnel:     opts.Tunnel,
	}

	locker := redsync.New(redsyncredis.NewPool(opts.Redis))
	startDistributedLock := locker.NewMutex("eventsub:startDistributedLock")

	opts.Lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				if opts.Config.AppEnv != "production" {
					if err := manager.
						gorm.
						Session(&gorm.Session{AllowGlobalUpdate: true}).
						Delete(&model.EventsubSubscription{}).
						Error; err != nil {
						return err
					}
				}

				go func() {
					if opts.Config.AppEnv != "production" {
						twitchClient, err := twitch.NewAppClient(opts.Config, opts.TokensGrpc)
						if err != nil {
							panic(err)
						}

						var subscriptions []helix.EventSubSubscription
						cursor := ""
						for {
							subs, err := twitchClient.GetEventSubSubscriptions(
								&helix.EventSubSubscriptionsParams{
									After: cursor,
								},
							)
							if err != nil {
								panic(err)
							}

							subscriptions = append(subscriptions, subs.Data.EventSubSubscriptions...)

							if subs.Data.Pagination.Cursor == "" {
								break
							}

							cursor = subs.Data.Pagination.Cursor
						}

						var unsubWg sync.WaitGroup

						for _, sub := range subscriptions {
							sub := sub
							unsubWg.Add(1)
							go func() {
								defer unsubWg.Done()
								manager.Unsubscribe(ctx, sub.ID)
							}()
						}

						unsubWg.Wait()
					}

					requestContext := context.Background()
					var channels []model.Channels
					err := manager.gorm.Where(
						`"channels"."isEnabled" = ? AND "User"."is_banned" = ? AND "channels"."isTwitchBanned" = ?`,
						true,
						false,
						false,
					).Joins("User").Find(&channels).Error
					if err != nil {
						panic(err)
					}

					var topics []model.EventsubTopic
					if err := opts.Gorm.WithContext(requestContext).Find(&topics).Error; err != nil {
						panic(err)
					}

					startDistributedLock.Lock()

					channelsWg := sync.WaitGroup{}

					for _, channel := range channels {
						channelsWg.Add(1)

						channel := channel

						go func() {
							defer channelsWg.Done()
							err = manager.SubscribeToNeededEvents(
								requestContext,
								topics,
								channel.ID,
								channel.BotID,
							)
							if err != nil {
								opts.Logger.Error(
									"failed to subscribe to needed events",
									slog.Any("err", err),
								)
							}
						}()
					}

					manager.SubscribeWithLimits(
						requestContext,
						&eventsub_framework.SubRequest{
							Type: "user.authorization.revoke",
							Condition: map[string]string{
								"client_id": opts.Config.TwitchClientId,
							},
							Callback: opts.Tunnel.GetAddr(),
							Secret:   opts.Config.TwitchClientSecret,
							Version:  "1",
						},
					)

					channelsWg.Wait()
					startDistributedLock.Unlock()
				}()

				return nil
			},
		},
	)

	return manager, nil
}

var statusesForSkip = []string{
	"enabled",
	"webhook_callback_verification_pending",
	"authorization_revoked",
	"user_removed",
	"version_removed",
}

func (c *Manager) SubscribeToNeededEvents(
	ctx context.Context,
	topics []model.EventsubTopic,
	broadcasterId,
	botId string,
) error {
	var existedSubscriptions []model.EventsubSubscription
	if err := c.gorm.
		WithContext(ctx).
		Where(&model.EventsubSubscription{UserID: broadcasterId}).
		Find(&existedSubscriptions).
		Error; err != nil {
		return err
	}

	var wg sync.WaitGroup
	newSubsCount := atomic.NewInt64(0)

	for _, topic := range topics {
		existedSubForTopic, subExists := lo.Find(
			existedSubscriptions,
			func(item model.EventsubSubscription) bool {
				return item.TopicID == topic.ID
			},
		)

		if subExists && slices.Contains(statusesForSkip, existedSubForTopic.Status) {
			continue
		}

		wg.Add(1)

		topic := topic
		go func() {
			defer wg.Done()
			condition := getTypeCondition(topic.ConditionType, topic.Topic, broadcasterId, botId)
			if condition == nil {
				c.logger.Error(
					"failed to get condition",
					slog.String("topic", topic.Topic),
					slog.String("channel_id", broadcasterId),
					slog.String("condition_type", string(topic.ConditionType)),
				)
				return
			}

			status, err := c.SubscribeWithLimits(
				ctx,
				&eventsub_framework.SubRequest{
					Type:      topic.Topic,
					Condition: condition,
					Callback:  c.tunnel.GetAddr(),
					Secret:    c.config.TwitchClientSecret,
					Version:   topic.Version,
				},
			)

			var casterErr *eventsub_framework.TwitchError
			if err != nil && !errors.As(err, &casterErr) {
				c.logger.Error(
					"failed to subscribe to event",
					slog.Any("err", err),
					slog.Any("topic", topic.Topic),
					slog.Any("condition", condition),
					slog.String("version", topic.Version),
					slog.String("callback", c.tunnel.GetAddr()),
				)
				return
			}

			subStatus := "unknown"
			subId := uuid.New()
			if status != nil && len(status.Data) > 0 {
				subStatus = status.Data[0].Status
				subId = uuid.MustParse(status.Data[0].ID)
			}

			if err := c.gorm.Create(
				&model.EventsubSubscription{
					ID:          subId,
					TopicID:     topic.ID,
					UserID:      broadcasterId,
					Status:      subStatus,
					Version:     topic.Version,
					CallbackUrl: c.tunnel.GetAddr(),
				},
			).Error; err != nil {
				c.logger.Error("failed to create subscription", slog.Any("err", err))
			}

			newSubsCount.Inc()
		}()
	}

	wg.Wait()

	if newSubsCount.Load() > 0 {
		c.logger.Info(
			"New subscriptions created for channel",
			slog.String("channel_id", broadcasterId),
			slog.String("bot_id", botId),
			slog.Int64("count", newSubsCount.Load()),
		)
	}

	return nil
}
