package main

import (
	"net/http"
	"tusk/config"
	"tusk/controllers"
	"tusk/models"

	"github.com/gin-gonic/gin"
)

func main() {
	// Database
	db := config.DatabaseConnection()
	db.AutoMigrate(&models.User{}, &models.Task{})
	config.CreateOwnerAccount(db)

	//Controller
	userController := controllers.UserController{DB: db}
	TaskController := controllers.TaskController{DB: db}

	// Router
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "Welcome to Tusk API")
	})

	router.POST("/users/login", userController.Login)
	router.POST("/users", userController.CreateAccount)
	router.DELETE("/users/:id", userController.DeleteAccount)
	router.GET("/users/Employee", userController.GetEmployee)

	router.POST("/tasks", TaskController.Create)
	router.DELETE("/tasks/:id", TaskController.Delete)
	router.PATCH("/tasks/:id/submit", TaskController.Submit)
	router.PATCH("/tasks/:id/reject", TaskController.Reject)
	router.PATCH("/tasks/:id/fix", TaskController.Fix)
	router.PATCH("/tasks/:id/approve", TaskController.Approve)
	router.GET("/tasks/:id", TaskController.FindById)
	router.GET("/tasks/review/asc", TaskController.NeedToReview)
	router.GET("/tasks/progress/:userId", TaskController.ProgressTask)
	router.GET("/tasks/stat/:userId", TaskController.Statistic)
	router.GET("/tasks/user/:userId/:status", TaskController.FindByUserAndStatus)

	router.Static("/attachments", "./attachments")
	router.Run("localhost:8080")
}
