package hitalentapp

import (
	"fmt"
	"hitalent/internal/department"
	departmentapi "hitalent/internal/department/api"
	"hitalent/internal/employee"
	employeeapi "hitalent/internal/employee/api"
	"log"
	"net/http"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Run() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		env("DB_HOST", "localhost"),
		env("DB_USER", "postgres"),
		env("DB_PASSWORD", "postgres"),
		env("DB_NAME", "hitalent"),
		env("DB_PORT", "5432"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})

	if err != nil {
		panic("failed to connect database")
	}

	var currentDB string
	db.Raw("SELECT current_database()").Scan(&currentDB)
	log.Println("current db:", currentDB)

	var currentSchema string
	db.Raw("SELECT current_schema()").Scan(&currentSchema)
	log.Println("current schema:", currentSchema)

	departmentService := department.NewDepartmentService(db)
	departmentsHandler := departmentapi.NewDepartmentsHandler(departmentService)

	employeeService := employee.NewEmployeeService(db)
	employeeHandler := employeeapi.NewEmployeeHandler(employeeService)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /departments", departmentsHandler.CreateDepartment)
	mux.HandleFunc("GET /departments/{id}", departmentsHandler.GetDepartment)
	mux.HandleFunc("PATCH /departments/{id}/update", departmentsHandler.UpdateDepartment)
	mux.HandleFunc("DELETE /departments/{id}/delete", departmentsHandler.DeleteDepartment)
	mux.HandleFunc("POST /departments/{id}/employees", employeeHandler.CreateEmployee)

	addr := env("APP_ADDR", ":8080")
	log.Println("listening on", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		panic(err)
	}
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
