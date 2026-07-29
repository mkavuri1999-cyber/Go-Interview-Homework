package graph

import (
	"context"

	"github.com/example/todo-homework/graph/model"
)

func (r *mutationResolver) CreateTask(ctx context.Context, userID string, title string, description *string, dueDate *string, tags []string) (*model.Task, error) {
	return r.Repo.CreateTask(ctx, userID, title, description, dueDate, tags)
}

func (r *mutationResolver) UpdateTaskStatus(ctx context.Context, id string, status model.TaskStatus) (*model.Task, error) {
	return r.Repo.UpdateTaskStatus(ctx, id, status)
}

func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
	return r.Repo.GetUserByID(ctx, id)
}

func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	return r.Repo.GetUsers(ctx)
}

func (r *queryResolver) Task(ctx context.Context, id string) (*model.Task, error) {
	return r.Repo.GetTaskByID(ctx, id)
}

func (r *queryResolver) Tasks(ctx context.Context, status *model.TaskStatus, userID *string) ([]*model.Task, error) {
	return r.Repo.GetTasks(ctx, status, userID)
}

func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
