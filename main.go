package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	prisma "golang-prisma/db"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type SignUp struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	RePassword string `json:"rePassword"`
}

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	client := prisma.NewClient()
	errPrisma := client.Prisma.Connect()
	if errPrisma != nil {
		return
	}

	defer func() {
		if err := client.Prisma.Disconnect(); err != nil {
			panic(err)
		}
	}()

	ctx := context.Background()

	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	router.POST("/signup", func(c *gin.Context) {
		var signup SignUp

		if err := c.BindJSON(&signup); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if signup.Password != signup.RePassword {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password does not match"})
			return
		}

		hashedPassword, err := HashPassword(signup.Password)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		data, err := client.User.CreateOne(prisma.User.Username.Set(signup.Username), prisma.User.Password.Set(hashedPassword)).Exec(ctx)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fmt.Println(data)

		c.JSON(http.StatusCreated, gin.H{
			"message": "User created",
			"data":    data,
		})
	})

	router.POST("/login", func(c *gin.Context) {
		var login Login
		if err := c.BindJSON(&login); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		data, err := client.User.FindUnique(prisma.User.Username.Equals(login.Username)).Exec(ctx)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if CheckPasswordHash(login.Password, data.Password) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Login success",
				"userId":  data.ID,
			})
			return
		}
	})

	errRouter := router.Run("localhost:8080")
	if errRouter != nil {
		return
	}
}
