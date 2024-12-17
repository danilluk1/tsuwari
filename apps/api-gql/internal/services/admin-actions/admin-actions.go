package admin_actions

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	model "github.com/satont/twir/libs/gomodels"
	"github.com/twirapp/twir/apps/api-gql/internal/entity"
	buscore "github.com/twirapp/twir/libs/bus-core"
	"github.com/twirapp/twir/libs/bus-core/eventsub"
	"github.com/twirapp/twir/libs/bus-core/timers"
	"github.com/twirapp/twir/libs/repositories/channels"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type Opts struct {
	fx.In

	Redis              *redis.Client
	ChannelsRepository channels.Repository
	TwirBus            *buscore.Bus
	Gorm               *gorm.DB
}

func New(opts Opts) *Service {
	return &Service{
		redis:              opts.Redis,
		channelsRepository: opts.ChannelsRepository,
		twirbus:            opts.TwirBus,
		gorm:               opts.Gorm,
	}
}

type Service struct {
	redis              *redis.Client
	channelsRepository channels.Repository
	twirbus            *buscore.Bus
	gorm               *gorm.DB
}

func (c *Service) DropAllAuthSessions(ctx context.Context) error {
	keys, err := c.redis.Keys(ctx, "scs:*").Result()
	if err != nil {
		return fmt.Errorf("failed to get session keys: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	err = c.redis.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete sessions: %w", err)
	}

	return nil
}

type EventSubSubscribeInput struct {
	Type      string                            `json:"type"`
	Version   string                            `json:"version"`
	Condition entity.EventsubSubscribeCondition `json:"condition"`
}

func (c *Service) EventSubSubscribe(ctx context.Context, input EventSubSubscribeInput) error {
	ch, err := c.channelsRepository.GetMany(
		ctx,
		channels.GetManyInput{
			Enabled: lo.ToPtr(true),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to get channels: %w", err)
	}

	for _, channel := range ch {
		go func() {
			c.twirbus.EventSub.Subscribe.Publish(
				eventsub.EventsubSubscribeRequest{
					ChannelID:     channel.ID,
					Topic:         input.Type,
					ConditionType: string(input.Condition),
					Version:       input.Version,
				},
			)
		}()
	}

	return nil
}

func (c *Service) RescheduleTimers(ctx context.Context) error {
	var entities []model.ChannelsTimers
	if err := c.gorm.Find(&entities).Error; err != nil {
		return fmt.Errorf("failed to get timers: %w", err)
	}

	for _, timer := range entities {
		c.twirbus.Timers.RemoveTimer.Publish(
			timers.AddOrRemoveTimerRequest{
				TimerID: timer.ID,
			},
		)

		c.twirbus.Timers.AddTimer.Publish(
			timers.AddOrRemoveTimerRequest{
				TimerID: timer.ID,
			},
		)
	}

	return nil
}
