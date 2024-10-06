package repository

import (
	"database/sql"
	"fiber-prjct-management-web/internal/models"
	"fmt"
	"strconv"
)

type UserRepository interface {
	Create(tx *sql.Tx, user *models.CreateUserInput) error
	Update(tx *sql.Tx, user *models.User) error
	UpdatePassword(tx *sql.Tx, user *models.UpdatePasswordInput) error
	Delete(tx *sql.Tx, id int) error
	SoftDelete(tx *sql.Tx, id int) error
	FindWithPagination(tx *sql.Tx, size int, page int, search string, role string, toDate string, fromDate string) ([]models.User, int, error)
	FindByID(tx *sql.Tx, id int) (models.User, error)
	FindByUsername(tx *sql.Tx, username string) (models.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) FindWithPagination(tx *sql.Tx, size int, page int, search string, role string, toDate string, fromDate string) ([]models.User, int, error) {
	var (
		users []models.User
		total int
	)
	offset := (page - 1) * size

	baseQueryCnt := "select count(id) from users where 1=1"
	baseQuery := "select id, username, role, created_at, updated_at from users where 1=1 and is_deleted = FALSE"
	paramQuery := ""
	dataQuery := []interface{}{}
	index := 1

	if search != "" {
		paramQuery += " and username like $" + strconv.Itoa(index)
		dataQuery = append(dataQuery, "%"+search+"%")
		index++
	}

	if role != "" {
		paramQuery += " and role = $" + strconv.Itoa(index)
		dataQuery = append(dataQuery, role)
		index++
	}

	if fromDate != "" {
		paramQuery += " and created_at >= $" + strconv.Itoa(index)
		dataQuery = append(dataQuery, fromDate)
		index++
	}

	if toDate != "" {
		paramQuery += " and created_at <= $" + strconv.Itoa(index)
		dataQuery = append(dataQuery, toDate)
		index++
	}

	// count total row before pagination
	baseQueryCnt += paramQuery
	err := tx.QueryRow(baseQueryCnt, dataQuery...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// modify baseQuery to include limit and offset
	baseQuery += paramQuery + " order by id desc limit $" + strconv.Itoa(index) + " offset $" + strconv.Itoa(index+1)
	dataQuery = append(dataQuery, size, offset)
	index += 2

	rows, err := tx.Query(baseQuery, dataQuery...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.Id, &user.Username, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}

		users = append(users, user)
	}

	return users, total, nil
}

func (r *userRepository) FindByID(tx *sql.Tx, id int) (models.User, error) {
	var model models.User

	err := tx.QueryRow("select id, username, role, password, created_at, updated_at, is_deleted from users where id = $1 and is_deleted = FALSE", id).Scan(&model.Id, &model.Username, &model.Role, &model.Password, &model.CreatedAt, &model.UpdatedAt, &model.IsDeleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return model, fmt.Errorf("user with id %d not found", id)
		}
		return model, err
	}

	return model, nil
}

func (r *userRepository) FindByUsername(tx *sql.Tx, username string) (models.User, error) {
	var model models.User

	err := tx.QueryRow("select id, username, role, password, is_deleted from users where username = $1 and is_deleted = FALSE", username).Scan(&model.Id, &model.Username, &model.Role, &model.Password, &model.IsDeleted)
	if err != nil {
		return model, err
	}

	return model, nil
}

func (r *userRepository) Create(tx *sql.Tx, user *models.CreateUserInput) error {
	if _, err := tx.Exec("insert into users (username, password, role) values ($1, $2, $3)", user.Username, user.Password, user.Role); err != nil {
		return err
	}

	return nil
}

func (r *userRepository) Update(tx *sql.Tx, user *models.User) error {
	if _, err := tx.Exec("update users set username = $1, password = $2, role = $3, updated_at = NOW() where id = $4", user.Username, user.Password, user.Role, user.Id); err != nil {
		return err
	}

	return nil
}

func (r *userRepository) UpdatePassword(tx *sql.Tx, user *models.UpdatePasswordInput) error {
	if _, err := tx.Exec("update users set password = $1, updated_at=now() where id = $2", user.Password, user.Id); err != nil {
		return err
	}

	return nil
}

func (r *userRepository) Delete(tx *sql.Tx, id int) error {
	if _, err := tx.Exec("delete from users where id = $1", id); err != nil {
		return err
	}

	return nil
}

func (r *userRepository) SoftDelete(tx *sql.Tx, id int) error {
	query := `
		update users 
		set 
			is_deleted = TRUE, 
			updated_at = now(),
			username = concat(username, ' (deleted)-', now()) 
		where id = $1
	`
	if _, err := tx.Exec(query, id); err != nil {
		return err
	}

	return nil
}
