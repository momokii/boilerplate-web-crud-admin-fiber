package models

import "database/sql"

type Project struct {
	Id            int            `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	StartDate     sql.NullString `json:"start_date"`
	EndDate       sql.NullString `json:"end_date"`
	Status        int            `json:"status"`
	Budget        int            `json:"budget"`
	CreatedBy     int            `json:"created_by"`
	CreatedByName string         `json:"created_by_name"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
}

type ProjectInput struct {
	Name        string `json:"name" validate:"required,min=5,max=50"`
	Description string `json:"description" validate:"required,min=5,max=255"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Status      int    `json:"status" validate:"required"`
	CreatedBy   int    `json:"created_by"`
	Budget      int    `json:"budget"`
}

type ProjectStats struct {
	TotalProjects          int            `json:"total_project"`
	TotalProjectsDone      int            `json:"total_project_done"`
	TotalProjectsOnGoing   int            `json:"total_project_ongoing"`
	TotalBudgetAllProjects int            `json:"total_budget_all_projects"`
	AvgBudgetProjects      float64        `json:"avg_budget_projects"`
	HighestBudgetProject   sql.NullString `json:"highest_budget_project"`
}

type ProjectStatusStats struct {
	Status string `json:"status"`
	Total  int    `json:"total"`
}
