package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"gotask/queue"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

func check(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

type PrintPayload struct {
	Text string `json:"text"`
}

type SavePayload struct {
	File *multipart.FileHeader
}

func GetStruct[T interface{}](data string) T {
	var values T
	if err := json.Unmarshal([]byte(data), &values); err != nil {
		panic("cannot unmarshal the string")
	}
	return values
}

func main() {
	jobs := &queue.Jobs{
		Jobs: queue.JobType{
			"print": func(payload string) {
				values := GetStruct[PrintPayload](payload)

				fmt.Println("working")
				fmt.Println(values.Text)
			},
			"save": func(payload string) {
				values := GetStruct[SavePayload](payload)

				fmt.Println(values.File.Filename)

				file, err := values.File.Open()
				check(err)
				defer file.Close()

				out, err := os.Create("hello.jpg")
				check(err)
				defer out.Close()

				_, err = io.Copy(out, file)
				check(err)
			},
		},
	}
	go queue.HandleJobs(*jobs)

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		message := c.Query("message")
		name := c.Query("name")
		jobType := c.Query("type")

		queue.CreateNewJob(name, PrintPayload{
			Text: message,
		}, jobType)

		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/save", func(ctx *gin.Context) {
		file, err := ctx.FormFile("file")
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"status": "failed",
			})
		}

		ctx.SaveUploadedFile(file, "hello.jpg")
		queue.CreateNewJob("save file", SavePayload{
			File: file,
		}, "save")

	})

	r.GET("/jobs", func(c *gin.Context) {
		jobs := queue.GetAllJobs()
		c.JSON(http.StatusOK, gin.H{
			"jobs": jobs,
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

}
