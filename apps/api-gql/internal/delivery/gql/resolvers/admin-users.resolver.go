package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.45

import (
	"context"
	"log/slog"

	data_loader "github.com/twirapp/twir/apps/api-gql/internal/delivery/gql/data-loader"
	"github.com/twirapp/twir/apps/api-gql/internal/delivery/gql/gqlmodel"
	"github.com/twirapp/twir/apps/api-gql/internal/delivery/gql/graph"
	"github.com/twirapp/twir/apps/api-gql/internal/delivery/gql/mappers"
	twir_users "github.com/twirapp/twir/apps/api-gql/internal/services/twir-users"
	"github.com/twirapp/twir/apps/api-gql/internal/services/users"
)

// SwitchUserBan is the resolver for the switchUserBan field.
func (r *mutationResolver) SwitchUserBan(ctx context.Context, userID string) (bool, error) {
	user, err := r.usersService.GetByID(ctx, userID)
	if err != nil {
		r.logger.Error("failed to get user by id", slog.Any("err", err))
		return false, err
	}

	isBanned := !user.IsBanned

	_, err = r.usersService.Update(
		ctx,
		userID, users.UpdateInput{
			IsBanned: &isBanned,
		},
	)
	if err != nil {
		r.logger.Error("failed to update user", slog.Any("err", err))
		return false, err
	}

	return true, nil
}

// SwitchUserAdmin is the resolver for the switchUserAdmin field.
func (r *mutationResolver) SwitchUserAdmin(ctx context.Context, userID string) (bool, error) {
	user, err := r.usersService.GetByID(ctx, userID)
	if err != nil {
		r.logger.Error("failed to get user by id", slog.Any("err", err))
		return false, err
	}

	isBotAdmin := !user.IsBotAdmin

	_, err = r.usersService.Update(
		ctx,
		userID, users.UpdateInput{
			IsBotAdmin: &isBotAdmin,
		},
	)
	if err != nil {
		r.logger.Error("failed to update user", slog.Any("err", err))
		return false, err
	}

	return true, nil
}

// TwirUsers is the resolver for the twirUsers field.
func (r *queryResolver) TwirUsers(ctx context.Context, opts gqlmodel.TwirUsersSearchParams) (*gqlmodel.TwirUsersResponse, error) {
	var page int
	perPage := 20

	if opts.Page.IsSet() {
		page = *opts.Page.Value()
	}

	if opts.PerPage.IsSet() {
		perPage = *opts.PerPage.Value()
	}

	manyInput := twir_users.GetManyInput{
		SearchQuery:       "",
		Page:              page,
		PerPage:           perPage,
		ChannelIsEnabled:  opts.IsBotEnabled.Value(),
		ChannelIsBotAdmin: opts.IsBotAdmin.Value(),
		UserIsBanned:      opts.IsBanned.Value(),
		HasBadges:         opts.Badges.Value(),
	}

	if opts.Search.IsSet() {
		manyInput.SearchQuery = *opts.Search.Value()
	}

	dbUsers, err := r.twirUsersService.GetMany(ctx, manyInput)
	if err != nil {
		r.logger.Error("failed to get many users", slog.Any("err", err))
		return nil, err
	}

	mappedUsers := make([]gqlmodel.TwirAdminUser, 0, len(dbUsers.Users))
	for _, e := range dbUsers.Users {
		mappedUsers = append(mappedUsers, mappers.UserWithChannelToAdminUser(e))
	}

	return &gqlmodel.TwirUsersResponse{
		Users: mappedUsers,
		Total: dbUsers.Total,
	}, nil
}

// TwitchProfile is the resolver for the twitchProfile field.
func (r *twirAdminUserResolver) TwitchProfile(ctx context.Context, obj *gqlmodel.TwirAdminUser) (*gqlmodel.TwirUserTwitchInfo, error) {
	return data_loader.GetHelixUserById(ctx, obj.ID)
}

// TwirAdminUser returns graph.TwirAdminUserResolver implementation.
func (r *Resolver) TwirAdminUser() graph.TwirAdminUserResolver { return &twirAdminUserResolver{r} }

type twirAdminUserResolver struct{ *Resolver }
