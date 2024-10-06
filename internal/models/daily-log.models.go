package models

import "database/sql"

type DailyLog struct {
	Id          int            `json:"id"`
	ProjectId   int            `json:"project_id"`
	LogDate     string         `json:"log_date"`
	Description string         `json:"description"`
	Issues      string         `json:"issues"`
	Income      int            `json:"income"`
	Expense     int            `json:"expense"`
	File        sql.NullString `json:"file"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
	ProjectName string         `json:"project_name"`
}

type DailyLogInput struct {
	ProjectId   int    `form:"project_id" json:"project_id" validate:"required"`
	LogDate     string `form:"log_date" json:"log_date" validate:"required"`
	Description string `form:"description" json:"description"`
	Issues      string `form:"issues" json:"issues"`
	Income      int    `form:"income" json:"income" validate:"min=0"`
	Expense     int    `form:"expense" json:"expense" validate:"min=0"`
	File        string `form:"file" json:"file"`
}

type DailyLogStats struct {
	TotalIncome           int            `json:"total_income"`
	TotalExpense          int            `json:"total_expense"`
	Budget                int            `json:"budget"`
	Balance               int            `json:"balance"`
	BudgetUsagePercentage float64        `json:"budget_usage_percentage"`
	TotalWorkingDays      int            `json:"total_working_days"`
	AvgDailyIncome        float64        `json:"avg_daily_income"`
	AvgDailyExpense       float64        `json:"avg_daily_expense"`
	HighestIncomeDay      sql.NullString `json:"highest_income_day"`
	HighestExpenseDay     sql.NullString `json:"highest_expense_day"`
}

type DailyLogStatsCumulative struct {
	LogDate         sql.NullString `json:"log_date"`
	Income          int            `json:"income"`
	Expense         int            `json:"expense"`
	CumulativeSaldo int            `json:"cumulative_saldo"`
}
