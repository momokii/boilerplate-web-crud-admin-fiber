package repository

import (
	"database/sql"
	"fiber-prjct-management-web/internal/models"
	"strconv"
)

type DailyLogRepository interface {
	Create(tx *sql.Tx, log *models.DailyLogInput) error
	Update(tx *sql.Tx, log *models.DailyLogInput, logId int) error
	Delete(tx *sql.Tx, id int) error
	FindWithPagination(tx *sql.Tx, size int, page int, search string, projectId int, fromDate string, toDate string, userId int, userRole int) ([]models.DailyLog, int, error)
	FindByID(tx *sql.Tx, id int) (models.DailyLog, error)
	FindByDate(tx *sql.Tx, date string, projectId int) (models.DailyLog, error)
	FindIfProjectAndLogOwner(tx *sql.Tx, projectId int, logId int, userId int) (models.DailyLog, error)
	FindStats(tx *sql.Tx, projectId int) (models.DailyLogStats, error)
	FindStatsCumulative(tx *sql.Tx, projectId int) ([]models.DailyLogStatsCumulative, error)
}

type dailyLogRepository struct {
	db *sql.DB
}

func NewDailyLogRepository(db *sql.DB) DailyLogRepository {
	return &dailyLogRepository{db}
}

func (r *dailyLogRepository) FindWithPagination(tx *sql.Tx, size int, page int, search string, projectId int, fromDate string, toDate string, userId int, userRole int) ([]models.DailyLog, int, error) {
	var (
		logs  []models.DailyLog
		total int
	)
	offset := (page - 1) * size

	baseQueryCnt := "select count(dl.id) from daily_logs dl left join projects p on dl.project_id = p.id where 1=1"
	baseQuery := `
		select 
			dl.id, dl.project_id, dl.log_date, dl.description, dl.issues, dl.income, dl.expense, dl.file, dl.created_at, dl.updated_at, p.name
		from 
			daily_logs dl left join projects p on dl.project_id = p.id
		where 1=1`

	paramQuery := ""
	dataQuery := []interface{}{}
	index := 1

	if search != "" {
		paramQuery += " and dl.description ilike $" + strconv.Itoa(index)
		dataQuery = append(dataQuery, "%"+search+"%")
		index++
	}

	if fromDate != "" {
		paramQuery += " and dl.log_date >= $" + strconv.Itoa(index)
		dataQuery = append(dataQuery, fromDate)
		index++
	}

	if toDate != "" {
		paramQuery += " and dl.log_date <= $" + strconv.Itoa(index)
		dataQuery = append(dataQuery, toDate)
		index++
	}

	if projectId != 0 {
		paramQuery += " and dl.project_id=$" + strconv.Itoa(index)
		dataQuery = append(dataQuery, projectId)
		index++
	} else {
		if userRole != 3 { // 3 is superadmin
			paramQuery += " and p.created_by=$" + strconv.Itoa(index)
			dataQuery = append(dataQuery, userId)
			index++
		}
	}

	// count
	baseQueryCnt += paramQuery
	if err := tx.QueryRow(baseQueryCnt, dataQuery...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// all data
	baseQuery += paramQuery + " order by log_date desc limit $" + strconv.Itoa(index) + " offset $" + strconv.Itoa(index+1)
	dataQuery = append(dataQuery, size, offset)
	index += 2
	rows, err := tx.Query(baseQuery, dataQuery...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var log models.DailyLog

		err := rows.Scan(&log.Id, &log.ProjectId, &log.LogDate, &log.Description, &log.Issues, &log.Income, &log.Expense, &log.File, &log.CreatedAt, &log.UpdatedAt, &log.ProjectName)
		if err != nil {
			return nil, 0, err
		}

		logs = append(logs, log)
	}

	return logs, total, nil
}

func (r *dailyLogRepository) FindStats(tx *sql.Tx, projectId int) (models.DailyLogStats, error) {

	var logStats models.DailyLogStats

	query := `
		WITH project_stats AS (
			SELECT 
				COALESCE(SUM(income), 0) as total_income,
				COALESCE(SUM(expense), 0) total_expense,
				COUNT(id) as total_working_days,
				MAX(CASE WHEN income = (SELECT MAX(income) FROM daily_logs WHERE project_id = $1) THEN log_date END) as highest_income_day,
				MAX(CASE WHEN expense = (SELECT MAX(expense) FROM daily_logs WHERE project_id = $1) THEN log_date END) as highest_expense_day
			FROM daily_logs
			WHERE project_id = $1
		), project_info AS (
			SELECT budget
			FROM projects
			WHERE id = $1
		)
		SELECT
			ps.total_income,
			ps.total_expense,
			pi.budget,
			ps.total_income - ps.total_expense as balance,
			CASE WHEN pi.budget > 0 THEN ROUND((ps.total_expense::numeric / pi.budget::numeric) * 100, 2) ELSE 0 END as budget_usage_percentage,
			ps.total_working_days,
			CASE WHEN ps.total_working_days > 0 THEN ROUND((ps.total_income::numeric / ps.total_working_days::numeric), 2) ELSE 0 END as avg_daily_income,
			CASE WHEN ps.total_working_days > 0 THEN ROUND((ps.total_expense::numeric / ps.total_working_days::numeric), 2) ELSE 0 END as avg_daily_expense,
			ps.highest_income_day,
			ps.highest_expense_day
		FROM 
			project_stats ps, 
			project_info pi
	`

	if err := tx.QueryRow(query, projectId).Scan(
		&logStats.TotalIncome,
		&logStats.TotalExpense,
		&logStats.Budget,
		&logStats.Balance,
		&logStats.BudgetUsagePercentage,
		&logStats.TotalWorkingDays,
		&logStats.AvgDailyIncome,
		&logStats.AvgDailyExpense,
		&logStats.HighestIncomeDay,
		&logStats.HighestExpenseDay,
	); err != nil {
		return logStats, err
	}

	return logStats, nil
}

func (r *dailyLogRepository) FindStatsCumulative(tx *sql.Tx, projectId int) ([]models.DailyLogStatsCumulative, error) {
	logStats := []models.DailyLogStatsCumulative{}

	query := `
		SELECT 
			log_date,
			income,
			expense,
			SUM(income - expense) OVER (ORDER BY log_date) as cumulative_saldo
		FROM daily_logs
		WHERE project_id = $1
		ORDER BY log_date
	`

	rows, err := tx.Query(query, projectId)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var logStat models.DailyLogStatsCumulative

		if err = rows.Scan(&logStat.LogDate, &logStat.Income, &logStat.Expense, &logStat.CumulativeSaldo); err != nil {
			return nil, err
		}

		logStats = append(logStats, logStat)
	}

	return logStats, nil
}

func (r *dailyLogRepository) FindByID(tx *sql.Tx, id int) (models.DailyLog, error) {
	var log models.DailyLog

	if err := tx.QueryRow("select id, project_id, log_date, description, issues, income, expense, file from daily_logs where id=$1", id).Scan(&log.Id, &log.ProjectId, &log.LogDate, &log.Description, &log.Issues, &log.Income, &log.Expense, &log.File); err != nil {
		return log, err
	}

	return log, nil
}

func (r *dailyLogRepository) FindByDate(tx *sql.Tx, date string, projectId int) (models.DailyLog, error) {
	var log models.DailyLog

	if err := tx.QueryRow("select id, project_id, log_date, description, issues, income, expense, file from daily_logs where log_date=$1 and project_id=$2", date, projectId).Scan(&log.Id, &log.ProjectId, &log.LogDate, &log.Description, &log.Issues, &log.Income, &log.Expense, &log.File); err != nil {
		return log, err
	}

	return log, nil
}

func (r *dailyLogRepository) FindIfProjectAndLogOwner(tx *sql.Tx, projectId int, logId int, userId int) (models.DailyLog, error) {
	var log models.DailyLog
	var createdBy int

	query := `
		select 
			dl.id, dl.project_id, dl.log_date, dl.description, dl.issues, dl.income, dl.expense, dl.file, p.created_by
		from daily_logs dl left join projects p on dl.project_id = p.id
		where 
			dl.id= $1
			and dl.project_id = $2
			and p.created_by = $3
	`

	if err := tx.QueryRow(query, logId, projectId, userId).Scan(&log.Id, &log.ProjectId, &log.LogDate, &log.Description, &log.Issues, &log.Income, &log.Expense, &log.File, &createdBy); err != nil {
		return log, err
	}

	return log, nil
}

func (r *dailyLogRepository) Create(tx *sql.Tx, log *models.DailyLogInput) error {
	if _, err := tx.Exec("insert into daily_logs (project_id, log_date, description, issues, income, expense, file) values ($1, $2, $3, $4, $5, $6, $7)", log.ProjectId, log.LogDate, log.Description, log.Issues, log.Income, log.Expense, log.File); err != nil {
		return err
	}

	return nil
}

func (r *dailyLogRepository) Update(tx *sql.Tx, log *models.DailyLogInput, logId int) error {
	if _, err := tx.Exec("update daily_logs set project_id=$1, log_date=$2, description=$3, issues=$4, income=$5, expense=$6, file=$7, updated_at=now() where id=$8", log.ProjectId, log.LogDate, log.Description, log.Issues, log.Income, log.Expense, log.File, logId); err != nil {
		return err
	}

	return nil
}

func (r *dailyLogRepository) Delete(tx *sql.Tx, id int) error {
	if _, err := tx.Exec("delete from daily_logs where id=$1", id); err != nil {
		return err
	}

	return nil
}
