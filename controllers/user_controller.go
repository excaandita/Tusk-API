package controllers

import (
	"net/http"
	"tusk/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
}

func (u *UserController) Login(c *gin.Context) {
	user := models.User{}
	errBindJson := c.ShouldBindJSON(&user)
	if errBindJson != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errBindJson.Error()})
		return
	}

	password := user.Password

	errDb := u.DB.Where("email = ? ", user.Email).Take(&user).Error

	if errDb != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Email or Password is incorrect"})
		return
	}

	errHash := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)

	if errHash != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email or Password is incorrect"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (u *UserController) CreateAccount(c *gin.Context) {
	user := models.User{}
	errBindJson := c.ShouldBindJSON(&user)
	if errBindJson != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errBindJson.Error()})
		return
	}

	emailExist := u.DB.Where("email = ? ", user.Email).First(&user).RowsAffected != 0

	if emailExist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email Already Exists"})
		return
	}

	hashedPasswordBytes, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	user.Password = string(hashedPasswordBytes)
	user.Role = "Employee"

	errDb := u.DB.Create(&user).Error

	if errDb != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDb.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (u *UserController) DeleteAccount(c *gin.Context) {
	id := c.Param("id")

	errFind := u.DB.Where("id=?", id).First(&models.User{}).RowsAffected == 0
	if errFind {
		c.JSON(http.StatusNotFound, gin.H{"error": "User Not Found"})
		return
	}

	errDb := u.DB.Delete(&models.User{}, id).Error
	if errDb != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDb.Error()})
		return
	}

	c.JSON(http.StatusOK, "Account Deleted")
}

func (u *UserController) GetEmployee(c *gin.Context) {
	users := []models.User{}

	errDb := u.DB.Select("id, name").Where("role=?", "Employee").Find(&users).Error

	if errDb != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDb.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}
