package repository

import (
	"database/sql"
	"fiber-prjct-management-web/internal/models"
	"strconv"
)

type ProjectRepository interface {
	Create(tx *sql.Tx, project *models.ProjectInput) error
	Update(tx *sql.Tx, project *models.ProjectInput, id int) error
	UpdateStatus(tx *sql.Tx, id int, status int) error
	Delete(tx *sql.Tx, id int) error
	FindWithPagination(tx *sql.Tx, size int, page int, search string, status string, toDate string, fromDate string, userId int) ([]models.Project, int, error)
	FindByID(tx *sql.Tx, id int) (models.Project, error)
	FindIfProjectOwner(tx *sql.Tx, id int, userId int) (models.Project, error)
	FindProjectsStats(tx *sql.Tx, userId int) (models.ProjectStats, error)
	FindProjectStatusStats(tx *sql.Tx, userId int) ([]models.ProjectStatusStats, error)
}

type projectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) ProjectRepository {
	return &projectRepository{db}
}

func (r *projectRepository) FindProjectStatusStats(tx *sql.Tx, userId int) ([]models.ProjectStatusStats, error) {

	var projectStatusStats []models.ProjectStatusStats

	paramData := []interface{}{}
	query := `
		SELECT 
			ps.name,
			COUNT(p.id)
		FROM 
			projects p LEFT JOIN project_status ps ON p.status = ps.id
	`

	if userId != 0 {
		query += " WHERE created_by = $1"
		paramData = append(paramData, userId)
	}

	query += ` GROUP BY ps.name`

	rows, err := tx.Query(query, paramData...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var projectStatus models.ProjectStatusStats

		if err := rows.Scan(&projectStatus.Status, &projectStatus.Total); err != nil {
			return nil, err
		}

		projectStatusStats = append(projectStatusStats, projectStatus)
	}

	return projectStatusStats, nil
}

func (r *projectRepository) FindProjectsStats(tx *sql.Tx, userId int) (models.ProjectStats, error) {

	var projectStats models.ProjectStats

	paramData := []interface{}{}
	query := `
		SELECT
			COUNT(id) as total_projects,
			COALESCE(SUM(CASE WHEN status = 3 THEN 1 ELSE 0 END), 0) as total_projects_done,
			COALESCE(SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END), 0) as total_projects_on_going,
			COALESCE(SUM(budget), 0) as total_budget_all_projects,
			CASE WHEN COUNT(id) = 0 THEN 0 ELSE ROUND(COALESCE(SUM(budget), 0) / COUNT(id), 2) END as avg_budget_per_project,
			MAX(CASE WHEN budget = (SELECT MAX(budget) FROM projects
	`

	if userId != 0 {
		query += " WHERE created_by = $1"
		paramData = append(paramData, userId)
	}

	query += `) THEN name ELSE NULL END) as project_max_budget
		FROM projects
	`

	if userId != 0 {
		query += " WHERE created_by = $1"
	}

	if err := tx.QueryRow(query, paramData...).Scan(
		&projectStats.TotalProjects,
		&projectStats.TotalProjectsDone,
		&projectStats.TotalProjectsOnGoing,
		&projectStats.TotalBudgetAllProjects,
		&projectStats.AvgBudgetProjects,
		&projectStats.HighestBudgetProject,
	); err != nil {
		return projectStats, err
	}

	return projectStats, nil
}

func (r *projectRepository) FindIfProjectOwner(tx *sql.Tx, id int, userId int) (models.Project, error) {
	var project models.Project

	if err := tx.QueryRow("select id, name, description, status, start_date, end_date, budget, created_by, created_at, updated_at from projects where id=$1 and created_by=$2", id, userId).Scan(&project.Id, &project.Name, &project.Description, &project.Status, &project.StartDate, &project.EndDate, &project.Budget, &project.CreatedBy, &project.CreatedAt, &project.UpdatedAt); err != nil {
		return project, err
	}

	return project, nil
}

func (r *projectRepository) FindWithPagination(tx *sql.Tx, size int, page int, search string, status string, toDate string, fromDate string, userId int) ([]models.Project, int, error) {
	var (
		projects []models.Project
		total    int
	)
	offset := (page - 1) * size

	baseQueryCnt := "select count(id) from projects where 1=1"
	baseQuery := "select id, name, description, status, start_date, end_date, budget, created_by, created_at, updated_at from projects where 1=1"
	paramQuery := ""
	dataQuery := []interface{}{}
	index := 1

	if search != "" {
		paramQuery += " and name ilike $" + strconv.Itoa(index)
		dataQuery = append(dataQuery, "%"+search+"%")
		index++
	}

	if status != "" {
		paramQuery += " and status = $" + strconv.Itoa(index)
		dataQuery = append(dataQuery, status)
		index++
	}

	if fromDate != "" {
		paramQuery += " and start_date >= $" + strconv.Itoa(index)
		dataQuery = append(dataQuery, fromDate)
		index++
	}

	if toDate != "" {
		paramQuery += " and start_date <= $" + strconv.Itoa(index)
		dataQuery = append(dataQuery, toDate)
		index++
	}

	// if need separate by user like super admin can see all project
	if userId != 0 {
		paramQuery += " and created_by = $" + strconv.Itoa(index)
		dataQuery = append(dataQuery, userId)
		index++
	}

	// count total pagiantion
	baseQueryCnt += paramQuery
	if err := tx.QueryRow(baseQueryCnt, dataQuery...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// base query with limit
	orderQuery := " order by id desc"
	if (fromDate != "") || (toDate != "") {
		orderQuery = " order by start_date desc"
	}

	baseQuery += paramQuery + orderQuery + " limit $" + strconv.Itoa(index) + " offset $" + strconv.Itoa(index+1)
	dataQuery = append(dataQuery, size, offset)
	index += 2

	rows, err := tx.Query(baseQuery, dataQuery...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var project models.Project
		err := rows.Scan(&project.Id, &project.Name, &project.Description, &project.Status, &project.StartDate, &project.EndDate, &project.Budget, &project.CreatedBy, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}

		projects = append(projects, project)
	}

	return projects, total, nil
}

func (r *projectRepository) FindByID(tx *sql.Tx, id int) (models.Project, error) {
	var project models.Project

	query := `
		SELECT 	
			p.id, p.name, description, status, start_date, end_date, budget, created_by, u.username, p.created_at, p.updated_at
		FROM 
			projects p LEFT JOIN users u ON p.created_by = u.id
		WHERE
			p.id = $1
	`

	if err := tx.QueryRow(query, id).Scan(&project.Id, &project.Name, &project.Description, &project.Status, &project.StartDate, &project.EndDate, &project.Budget, &project.CreatedBy, &project.CreatedByName, &project.CreatedAt, &project.UpdatedAt); err != nil {
		return project, err
	}

	return project, nil
}

func (r *projectRepository) Create(tx *sql.Tx, project *models.ProjectInput) error {
	paramIndex := 5
	paramValues := []interface{}{project.Name, project.Description, project.Status, project.Budget, project.CreatedBy}
	baseQuery := "insert into projects (name, description, status, budget, created_by"
	if project.StartDate != "" {
		baseQuery += ", start_date"
		paramIndex++
		paramValues = append(paramValues, project.StartDate)
	}
	baseQuery += ") values ("

	for i := 0; i < paramIndex; i++ {
		baseQuery += "$" + strconv.Itoa(i+1) + ","
	}
	baseQuery = baseQuery[:len(baseQuery)-1]
	baseQuery += " )"

	if _, err := tx.Exec(baseQuery, paramValues...); err != nil {
		return err
	}

	return nil
}

func (r *projectRepository) Update(tx *sql.Tx, project *models.ProjectInput, id int) error {
	indexQuery := 4
	queryData := []interface{}{project.Name, project.Description, project.Status, project.Budget}
	query := "update projects set name=$1, description=$2, status=$3, budget=$4, updated_at=now()"

	if project.StartDate != "" {
		query += ", start_date=$" + strconv.Itoa(indexQuery+1)
		queryData = append(queryData, project.StartDate)
		indexQuery++
	} else {
		query += ", start_date=null"
	}

	if project.EndDate != "" {
		query += ", end_date=$" + strconv.Itoa(indexQuery+1)
		queryData = append(queryData, project.EndDate)
		indexQuery++
	} else {
		query += ", end_date=null"
	}

	query += " where id=$" + strconv.Itoa(indexQuery+1) + " and created_by=$" + strconv.Itoa(indexQuery+2)
	queryData = append(queryData, id, project.CreatedBy)

	if _, err := tx.Exec(query, queryData...); err != nil {
		return err
	}

	return nil
}

func (r *projectRepository) UpdateStatus(tx *sql.Tx, id int, status int) error {
	if _, err := tx.Exec("update projects set status=$1 where id=$2", status, id); err != nil {
		return err
	}

	return nil
}

func (r *projectRepository) Delete(tx *sql.Tx, id int) error {
	if _, err := tx.Exec("delete from projects where id=$1", id); err != nil {
		return err
	}

	return nil
}
