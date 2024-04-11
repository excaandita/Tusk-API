package controllers

import (
	"net/http"
	"os"
	"strconv"
	"tusk/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TaskController struct {
	DB *gorm.DB
}

func (t *TaskController) Create(c *gin.Context) {
	task := models.Task{}
	errBindJson := c.ShouldBindJSON(&task)
	if errBindJson != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errBindJson.Error()})
		return
	}

	errDb := t.DB.Create(&task).Error

	if errDb != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDb.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (t *TaskController) Delete(c *gin.Context) {
	id := c.Param("id")
	task := models.Task{}

	if errFind := t.DB.First(&task, id).Error; errFind != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task Not Found"})
		return
	}

	errDb := t.DB.Delete(&models.Task{}, id).Error
	if errDb != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDb.Error()})
		return
	}

	if task.Attachment != "" {
		os.Remove("attachments/" + task.Attachment)
	}

	c.JSON(http.StatusOK, "Task Deleted")
}

func (t *TaskController) Submit(c *gin.Context) {
	id := c.Param("id")
	task := models.Task{}

	submitDate := c.PostForm("submitDate")
	file, errFile := c.FormFile("attachment")

	if errFile != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errFile.Error()})
		return
	}

	if errFind := t.DB.First(&task, id).Error; errFind != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task Not Found"})
		return
	}

	//Remove attachments
	attachment := task.Attachment
	fileInfo, _ := os.Stat("/attachments/" + attachment)
	if fileInfo != nil {
		os.Remove("attachments/" + attachment)
	}

	//New attachments
	attachment = file.Filename
	errSave := c.SaveUploadedFile(file, "attachment/"+attachment)
	if errSave != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errSave.Error()})
		return
	}

	errDb := t.DB.Where("id=?", id).Updates(models.Task{
		Status:     "Review",
		SubmitDate: submitDate,
		Attachment: attachment,
	}).Error
	if errDb != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDb.Error()})
		return
	}

	c.JSON(http.StatusOK, "Task Submited wait to Review")
}

func (t *TaskController) Reject(c *gin.Context) {
	id := c.Param("id")
	task := models.Task{}

	rejectedDate := c.PostForm("rejectedDate")
	reason := c.PostForm("reason")

	if errFind := t.DB.First(&task, id).Error; errFind != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task Not Found"})
		return
	}

	errDb := t.DB.Where("id=?", id).Updates(models.Task{
		Status:       "Rejected",
		Reason:       reason,
		RejectedDate: rejectedDate,
	}).Error

	if errDb != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDb.Error()})
		return
	}

	c.JSON(http.StatusOK, "Task Rejected")
}

func (t *TaskController) Fix(c *gin.Context) {
	id := c.Param("id")

	revision, errConv := strconv.Atoi(c.PostForm("revision"))

	if errConv != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errConv.Error()})
		return
	}

	if errFind := t.DB.First(&models.Task{}, id).Error; errFind != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task Not Found"})
		return
	}

	errDb := t.DB.Where("id=?", id).Updates(models.Task{
		Status:   "Queue",
		Revision: int8(revision),
	}).Error

	if errDb != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDb.Error()})
		return
	}

	c.JSON(http.StatusOK, "Task Fix to Queue with Revision")
}

func (t *TaskController) Approve(c *gin.Context) {
	id := c.Param("id")

	approvedDate := c.PostForm("approvedDate")

	if errFind := t.DB.First(&models.Task{}, id).Error; errFind != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task Not Found"})
		return
	}

	errDb := t.DB.Where("id=?", id).Updates(models.Task{
		Status:       "Approved",
		ApprovedDate: approvedDate,
	}).Error

	if errDb != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errDb.Error()})
		return
	}

	c.JSON(http.StatusOK, "Task Approved")
}

func (t *TaskController) FindById(c *gin.Context) {
	task := models.Task{}
	id := c.Param("id")

	if errFind := t.DB.First(&models.Task{}, id).Error; errFind != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task Not Found"})
		return
	}

	errDb := t.DB.Preload("User").Find(&task, id).Error
	if errDb != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": errDb.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (t *TaskController) NeedToReview(c *gin.Context) {
	tasks := []models.Task{}

	errDb := t.DB.Preload("User").Where("status=?", "Review").Order("submit_date ASC").Limit(2).Find(&tasks).Error
	if errDb != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": errDb.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (t *TaskController) ProgressTask(c *gin.Context) {
	tasks := []models.Task{}
	userId := c.Param("userId")

	errDb := t.DB.Preload("User").Where(
		"(status!=? and user_id=?) OR (revision!=? and user_id=?) ", "Queue", userId, 0, userId,
	).Order("updated_at DESC").Limit(5).Find(&tasks).Error

	if errDb != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": errDb.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (t *TaskController) Statistic(c *gin.Context) {
	userId := c.Param("userId")

	stat := []map[string]interface{}{}

	errDb := t.DB.Model(models.Task{}).Select("status, count(status) as total").Where("user_id=? ", userId).Group("status").Order("updated_at DESC").Limit(5).Find(&stat).Error

	if errDb != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": errDb.Error()})
		return
	}

	c.JSON(http.StatusOK, stat)
}

func (t *TaskController) FindByUserAndStatus(c *gin.Context) {
	tasks := []models.Task{}

	userId := c.Param("userId")
	status := c.Param("status")

	errDb := t.DB.Preload("User").Where("status=? AND user_id=?", status, userId).Find(&tasks).Error
	if errDb != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": errDb.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}
