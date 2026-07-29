package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/example/todo-homework/graph/model"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	err := r.DB.QueryRowContext(ctx, "SELECT id, email, name FROM users WHERE id = $1", id).
		Scan(&u.ID, &u.Email, &u.Name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}
	return &u, nil
}

func (r *Repository) GetUsers(ctx context.Context) ([]*model.User, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT id, email, name FROM users ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}

func (r *Repository) GetTaskByID(ctx context.Context, id string) (*model.Task, error) {
	var t model.Task
	var statusStr string
	var dueDate sql.NullString
	var desc sql.NullString
	var userID string

	query := `SELECT id, user_id, title, description, status, due_date FROM tasks WHERE id = $1`
	err := r.DB.QueryRowContext(ctx, query, id).Scan(&t.ID, &userID, &t.Title, &desc, &statusStr, &dueDate)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query task: %w", err)
	}

	if desc.Valid {
		t.Description = &desc.String
	}
	if dueDate.Valid {
		t.DueDate = &dueDate.String
	}
	t.Status = model.TaskStatus(strings.ToUpper(statusStr))

	tags, err := r.GetTagsByTaskID(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	t.Tags = tags

	t.User = &model.User{ID: userID}

	return &t, nil
}

func (r *Repository) GetTasks(ctx context.Context, status *model.TaskStatus, userID *string) ([]*model.Task, error) {
	query := `SELECT id, user_id, title, description, status, due_date FROM tasks WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if status != nil {
		query += fmt.Sprintf(" AND LOWER(status) = LOWER($%d)", argIdx)
		args = append(args, status.String())
		argIdx++
	}

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, *userID)
		argIdx++
	}

	query += " ORDER BY id ASC"

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		var t model.Task
		var statusStr string
		var dueDate sql.NullString
		var desc sql.NullString
		var uID string

		if err := rows.Scan(&t.ID, &uID, &t.Title, &desc, &statusStr, &dueDate); err != nil {
			return nil, err
		}

		if desc.Valid {
			t.Description = &desc.String
		}
		if dueDate.Valid {
			t.DueDate = &dueDate.String
		}
		t.Status = model.TaskStatus(strings.ToUpper(statusStr))
		t.User = &model.User{ID: uID}

		tags, err := r.GetTagsByTaskID(ctx, t.ID)
		if err != nil {
			return nil, err
		}
		t.Tags = tags

		tasks = append(tasks, &t)
	}
	return tasks, nil
}

func (r *Repository) GetTagsByTaskID(ctx context.Context, taskID string) ([]string, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT tag FROM task_tags WHERE task_id = $1", taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (r *Repository) CreateTask(ctx context.Context, userID string, title string, desc *string, dueDate *string, tags []string) (*model.Task, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var taskID string
	query := `INSERT INTO tasks (user_id, title, description, status, due_date) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	status := strings.ToLower(string(model.TaskStatusPending))

	err = tx.QueryRowContext(ctx, query, userID, title, desc, status, dueDate).Scan(&taskID)
	if err != nil {
		return nil, fmt.Errorf("insert task: %w", err)
	}

	for _, tag := range tags {
		_, err := tx.ExecContext(ctx, "INSERT INTO task_tags (task_id, tag) VALUES ($1, $2)", taskID, tag)
		if err != nil {
			return nil, fmt.Errorf("insert tag: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetTaskByID(ctx, taskID)
}

func (r *Repository) UpdateTaskStatus(ctx context.Context, id string, status model.TaskStatus) (*model.Task, error) {
	statusStr := strings.ToLower(string(status))
	res, err := r.DB.ExecContext(ctx, "UPDATE tasks SET status = $1 WHERE id = $2", statusStr, id)
	if err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return nil, fmt.Errorf("task with id %s not found", id)
	}

	return r.GetTaskByID(ctx, id)
}
