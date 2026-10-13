package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nagi-hikaru/todo-api/internal/database"
	"github.com/nagi-hikaru/todo-api/internal/models"
)

type TodoHandler interface {
	CreateTodoHandler(c *gin.Context)
	DeleteTodoHandler(c *gin.Context)

	UpdateStatusHandler(c *gin.Context)
	UpdateTodoHandler(c *gin.Context)

	GetTodoHandler(c *gin.Context)
	ListTodosHandler(c *gin.Context)
}

type TodoHandlerImpl struct {
	dbService database.DatabaseService
}

// 发送成功消息
func success(c *gin.Context, status int, message string, data any) {
	c.JSON(status, models.APIResponse{
		Success: true, Message: message, Data: data, Timestamp: time.Now(),
	})
}

// 发送失败消息
func failure(c *gin.Context, status int, message string, err error) {
	response := models.APIResponse{
		Success: false, Message: message, Timestamp: time.Now(),
	}
	if err != nil {
		response.Error = err.Error()
	}
	c.JSON(status, response)
}

func databaseError(c *gin.Context, err error) {
	if errors.Is(err, database.ErrNotFound) {
		failure(c, http.StatusNotFound, "todo 不存在", err)
		return
	}
	failure(c, http.StatusInternalServerError, "服务器内部错误", err)
}

// 判断日期字符串格式是否正确
func validDate(date string) bool {
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}

// 解析 ID
func parseID(c *gin.Context) (int, error) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return 0, err
	}
	if id <= 0 {
		return 0, errors.New("无效的 ID")
	}
	return id, nil
}

// 创建 todo 处理函数
func NewTodoHandler(databaseService database.DatabaseService) TodoHandler {
	return &TodoHandlerImpl{dbService: databaseService}
}

// CreateTodoHandler 创建任务。
// @Summary 创建任务
// @Description 创建待办任务，初始状态为 pending；截止日期格式为 YYYY-MM-DD。
// @Tags Todos
// @Accept json
// @Produce json
// @Param body body models.TodoCreateRequest true "任务信息"
// @Success 201 {object} models.APIResponse{data=models.Todo} "创建成功"
// @Failure 400 {object} models.APIResponse "请求参数或截止日期无效"
// @Failure 500 {object} models.APIResponse "服务器内部错误"
// @Router /todos [post]
func (h *TodoHandlerImpl) CreateTodoHandler(c *gin.Context) {
	var req models.TodoCreateRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		failure(c, http.StatusBadRequest, "请求参数无效", err)
		return
	}

	if req.DueDate != nil && !validDate(*req.DueDate) {
		failure(c, http.StatusBadRequest, "截止日期无效", errors.New("due_date 必须是有效的 YYYY-MM-DD 日期"))
		return
	}

	todo := &models.Todo{
		Title: req.Title, Status: "pending", Description: req.Description,
		Priority: req.Priority, DueDate: req.DueDate,
	}

	if err := h.dbService.Create(todo); err != nil {
		databaseError(c, err)
		return
	}

	success(c, http.StatusCreated, "todo 创建成功", todo)
}

// DeleteTodoHandler 删除任务。
// @Summary 删除任务
// @Tags Todos
// @Produce json
// @Param id path int true "任务 ID" minimum(1)
// @Success 200 {object} models.APIResponse "删除成功"
// @Failure 400 {object} models.APIResponse "任务 ID 无效"
// @Failure 404 {object} models.APIResponse "任务不存在"
// @Failure 500 {object} models.APIResponse "服务器内部错误"
// @Router /todos/{id} [delete]
func (h *TodoHandlerImpl) DeleteTodoHandler(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		failure(c, http.StatusBadRequest, "todo ID 不正确", err)
		return
	}

	if err := h.dbService.Delete(id); err != nil {
		databaseError(c, err)
		return
	}

	success(c, http.StatusOK, "todo 删除成功", nil)
}

// UpdateStatusHandler 更新任务状态。
// @Summary 更新任务状态
// @Description 通过查询参数传入目标状态；设为 completed 时记录完成时间，设为其他合法状态时清除完成时间。
// @Tags Todos
// @Produce json
// @Param id path int true "任务 ID" minimum(1)
// @Param status query string true "目标状态" Enums(pending,completed,cancelled)
// @Success 200 {object} models.APIResponse "状态更新成功"
// @Failure 400 {object} models.APIResponse "任务 ID 无效或状态为空"
// @Failure 404 {object} models.APIResponse "任务不存在"
// @Failure 500 {object} models.APIResponse "数据库更新失败，包括非法状态触发的约束错误"
// @Router /todos/{id}/status [patch]
func (h *TodoHandlerImpl) UpdateStatusHandler(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		failure(c, http.StatusBadRequest, "todo ID 不正确", err)
		return
	}

	status := c.Query("status")
	if status == "" {
		failure(c, http.StatusBadRequest, "没有传入有效 status", nil)
		return
	}
	if err := h.dbService.UpdateStatus(id, status); err != nil {
		databaseError(c, err)
		return
	}

	success(c, http.StatusOK, "更新状态成功", nil)
}

// UpdateTodoHandler 更新任务内容。
// @Summary 更新任务内容
// @Description 标题必填；未提供的描述、优先级和截止日期会被清空；状态保持不变；截止日期格式为 YYYY-MM-DD。
// @Tags Todos
// @Accept json
// @Produce json
// @Param id path int true "任务 ID" minimum(1)
// @Param body body models.TodoUpdateRequest true "任务信息"
// @Success 200 {object} models.APIResponse{data=models.Todo} "更新后的完整任务"
// @Failure 400 {object} models.APIResponse "任务 ID、请求参数或截止日期无效"
// @Failure 404 {object} models.APIResponse "任务不存在"
// @Failure 500 {object} models.APIResponse "服务器内部错误"
// @Router /todos/{id} [put]
func (h *TodoHandlerImpl) UpdateTodoHandler(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		failure(c, http.StatusBadRequest, "todo ID 不正确", err)
		return
	}

	var req models.TodoUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failure(c, http.StatusBadRequest, "请求参数无效", err)
		return
	}

	if req.DueDate != nil && !validDate(*req.DueDate) {
		failure(c, http.StatusBadRequest, "截止日期无效", errors.New("due_date 必须是有效的 YYYY-MM-DD 日期"))
		return
	}

	todo := &models.Todo{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
	}

	if err := h.dbService.Update(todo); err != nil {
		databaseError(c, err)
		return
	}

	updated, err := h.dbService.GetByID(id)
	if err != nil {
		databaseError(c, err)
		return
	}

	success(c, http.StatusOK, "todo 更新成功", updated)
}

// GetTodoHandler 根据 ID 查询任务。
// @Summary 根据 ID 查询任务
// @Tags Todos
// @Produce json
// @Param id path int true "任务 ID" minimum(1)
// @Success 200 {object} models.APIResponse{data=models.Todo} "任务详情"
// @Failure 400 {object} models.APIResponse "任务 ID 无效"
// @Failure 404 {object} models.APIResponse "任务不存在"
// @Failure 500 {object} models.APIResponse "服务器内部错误"
// @Router /todos/{id} [get]
func (h *TodoHandlerImpl) GetTodoHandler(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		failure(c, http.StatusBadRequest, "todo ID 不正确", err)
		return
	}

	todo, err := h.dbService.GetByID(id)
	if err != nil {
		databaseError(c, err)
		return
	}

	success(c, http.StatusOK, "获取 todo 成功", todo)
}

// ListTodosHandler 筛选并分页查询任务。
// @Summary 查询任务列表
// @Description 支持按状态、优先级及标题或描述关键词筛选；返回当前页任务与符合筛选条件的总数。
// @Tags Todos
// @Produce json
// @Param status query string false "任务状态" Enums(pending,completed,cancelled)
// @Param priority query string false "优先级" Enums(low,medium,high)
// @Param search query string false "标题或描述关键词，可为空"
// @Param page query int false "页码" default(1) minimum(1)
// @Param page_size query int false "每页条数" default(20) minimum(1) maximum(100)
// @Success 200 {object} models.APIResponse{data=models.PaginateResponse{items=[]models.Todo}} "分页查询结果"
// @Failure 400 {object} models.APIResponse "查询参数无效"
// @Failure 500 {object} models.APIResponse "服务器内部错误"
// @Router /todos [get]
func (h *TodoHandlerImpl) ListTodosHandler(c *gin.Context) {
	var filter models.TodoFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		failure(c, http.StatusBadRequest, "请求参数无效", err)
		return
	}

	todos, total, err := h.dbService.List(filter)
	if err != nil {
		databaseError(c, err)
		return
	}

	data := models.PaginateResponse{
		Items: todos,
		Total: total,
	}
	success(c, http.StatusOK, "筛选 todos 成功", data)

}
