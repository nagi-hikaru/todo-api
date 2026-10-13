package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/nagi-hikaru/todo-api/internal/models"
)

type DatabaseService interface {
	Create(todo *models.Todo) error
	Delete(id int) error

	UpdateStatus(id int, status string) error
	Update(todo *models.Todo) error

	GetByID(id int) (*models.Todo, error)
	List(filter models.TodoFilter) ([]models.Todo, int, error)
}

type DatabaseServiceImpl struct {
	db *sql.DB
}

var ErrNotFound = errors.New("todo 不存在")

func NewDatabaseService(db *sql.DB) DatabaseService {
	return &DatabaseServiceImpl{db: db}
}

// 格式化错误
func fmtError(message string, id *int, err error) error {
	sep := "----------"
	if id == nil {
		return fmt.Errorf("%s: \nerr: %w\n%s", message, err, sep)
	}
	return fmt.Errorf("%s: \nid: %d, \nerr: %w\n%s", message, *id, err, sep)
}

// 插入 todo ✅
func (s *DatabaseServiceImpl) Create(todo *models.Todo) error {
	insert := `
		INSERT INTO todos (title, description, status, priority, due_date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().UTC()
	todo.CreatedAt, todo.UpdatedAt = now, now
	if todo.Status == "" {
		todo.Status = "pending"
	}

	result, err := s.db.Exec(insert, todo.Title, todo.Description, todo.Status, todo.Priority,
		todo.DueDate, todo.CreatedAt, todo.UpdatedAt)
	if err != nil {
		return fmtError("创建 todo 失败", nil, err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmtError("获取 todo ID 失败", nil, err)
	}

	todo.ID = int(id)
	return nil
}

// 删除 todo ✅
func (s *DatabaseServiceImpl) Delete(id int) error {
	delete := `DELETE FROM todos WHERE id = ?`

	result, err := s.db.Exec(delete, id)
	if err != nil {
		return fmtError("删除 todo 失败", &id, err)
	}

	num, err := result.RowsAffected()
	if err != nil {
		return fmtError("获取修改行数失败", &id, err)
	}
	if num == 0 {
		return ErrNotFound
	}
	return nil
}

// 更新 todo 的 status ✅
func (s *DatabaseServiceImpl) UpdateStatus(id int, status string) error {
	update := `
		UPDATE todos SET status = ?, updated_at = ?, 
		completed_at = CASE
			WHEN ? == 'completed' THEN ?
			WHEN completed_at IS NOT NULL THEN null
			ELSE completed_at
		END
		WHERE id = ?
	`

	now := time.Now().UTC()
	result, err := s.db.Exec(update, status, now, status, now, id)
	if err != nil {
		return fmtError("更新 todo 失败", &id, err)
	}

	num, err := result.RowsAffected()
	if err != nil {
		return fmtError("获取修改行数失败", &id, err)
	}
	if num == 0 {
		return ErrNotFound
	}

	return nil
}

// 更新 todo ✅
func (s *DatabaseServiceImpl) Update(todo *models.Todo) error {
	update := `
		UPDATE todos SET title = ?, description = ?, priority = ?, due_date = ?, 
		updated_at = ? WHERE id = ?
	`
	todo.UpdatedAt = time.Now().UTC()
	result, err := s.db.Exec(update, todo.Title, todo.Description, todo.Priority, todo.DueDate, todo.UpdatedAt, todo.ID)
	if err != nil {
		return fmtError("更新 todo 失败", &todo.ID, err)
	}

	num, err := result.RowsAffected()
	if err != nil {
		return fmtError("获取修改行数失败", &todo.ID, err)
	}
	if num == 0 {
		return ErrNotFound
	}

	return nil
}

// 通过 ID 获取对应的 todo ✅
func (s *DatabaseServiceImpl) GetByID(id int) (*models.Todo, error) {
	var todo models.Todo
	query := `
		SELECT id, title, description, status, priority, due_date, created_at, 
		updated_at, completed_at From todos WHERE id = ?
	`
	var description, priority sql.NullString
	var due_date, completed_at sql.NullTime
	err := s.db.QueryRow(query, id).Scan(
		&todo.ID, &todo.Title, &description, &todo.Status, &priority,
		&due_date, &todo.CreatedAt, &todo.UpdatedAt, &completed_at,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmtError("查询任务失败", &id, err)
	}

	if description.Valid {
		todo.Description = &description.String
	}
	if priority.Valid {
		todo.Priority = &priority.String
	}
	if due_date.Valid {
		date := due_date.Time.Format("2006-01-02")
		todo.DueDate = &date
	}
	if completed_at.Valid {
		todo.CompletedAt = &completed_at.Time
	}

	return &todo, nil
}

// 筛选 todo ✅
func (s *DatabaseServiceImpl) List(filter models.TodoFilter) ([]models.Todo, int, error) {
	// 构建初始条件
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.Status != nil {
		whereClause += " AND status = ?"
		args = append(args, *filter.Status)
	}

	if filter.Priority != nil {
		whereClause += " AND priority = ?"
		args = append(args, *filter.Priority)
	}

	if filter.Search != nil {
		whereClause += " AND (title LIKE ? OR description LIKE ?)"
		searchTerm := "%" + *filter.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	countQuery := "SELECT COUNT(*) FROM todos " + whereClause
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmtError("获取 todo 总数失败", nil, err)
	}

	query := `
		SELECT id, title, description, status, priority, due_date,
		       created_at, updated_at, completed_at
		FROM todos ` + whereClause + `
		ORDER BY created_at DESC, priority DESC
		LIMIT ? OFFSET ?
	`

	offset := (filter.Page - 1) * filter.PageSize
	args = append(args, filter.PageSize, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmtError("查询 todo 列表失败", nil, err)
	}
	defer rows.Close()

	todos := make([]models.Todo, 0)
	for rows.Next() {
		var description, priority sql.NullString
		var due_date, completed_at sql.NullTime
		todo := &models.Todo{}
		err := rows.Scan(
			&todo.ID, &todo.Title, &description, &todo.Status, &priority,
			&due_date, &todo.CreatedAt, &todo.UpdatedAt, &completed_at,
		)
		if err != nil {
			return nil, 0, fmtError("扫描 todo 列表失败", nil, err)
		}

		if description.Valid {
			todo.Description = &description.String
		}
		if priority.Valid {
			todo.Priority = &priority.String
		}
		if due_date.Valid {
			date := due_date.Time.Format("2006-01-02")
			todo.DueDate = &date
		}
		if completed_at.Valid {
			todo.CompletedAt = &completed_at.Time
		}

		todos = append(todos, *todo)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmtError("遍历 todo 列表失败", nil, err)
	}

	return todos, total, nil
}
