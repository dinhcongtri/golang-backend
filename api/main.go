package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	prisma "golang-prisma/api/db"
	"net/http"
	"strconv"
)

//func HashPassword(password string) (string, error) {
//	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
//	return string(bytes), err
//}
//
//func CheckPasswordHash(password, hash string) bool {
//	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
//	return err == nil
//}
//
//type SignUp struct {
//	Username   string `json:"username"`
//	Password   string `json:"password"`
//	RePassword string `json:"rePassword"`
//}
//
//type Login struct {
//	Username string `json:"username"`
//	Password string `json:"password"`
//}

type CreateNote struct {
	UserId   int    `json:"userId"`
	Content  string `json:"content"`
	CreateAt string `json:"createAt"`
	UpdateAt string `json:"updateAt"`
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
	//===============================REST API=====================================
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

		data, errCreate := client.User.CreateOne(prisma.User.Username.Set(signup.Username), prisma.User.Password.Set(hashedPassword)).Exec(ctx)

		if errCreate != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

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

	router.POST("/create-note", func(c *gin.Context) {
		var createNote CreateNote
		if err := c.BindJSON(&createNote); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		data, errCreate := client.Note.CreateOne(prisma.Note.Content.Set(createNote.Content)).Exec(ctx)
		if errCreate != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": errCreate.Error()})
			return
		}

		dataRoot, errRoot := client.UserNote.CreateOne(prisma.UserNote.UserID.Cursor(createNote.UserId), prisma.UserNote.NoteID.Cursor(data.ID)).Exec(ctx)
		if errRoot != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": errRoot.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Note created",
			"data":    dataRoot,
		})
		return
	})

	router.GET("/note", func(c *gin.Context) {
		param := c.Request.URL.Query()
		noteId := param["id"][0]
		userId, _ := strconv.ParseInt(c.Request.Header.Get("userId"), 10, 0)

		_, errFind := client.UserNote.FindMany(prisma.UserNote.NoteID.Equals(noteId), prisma.UserNote.UserID.Equals(int(userId))).Exec(ctx)
		if errFind != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": errFind.Error()})
			return
		}

		data, errNote := client.Note.FindUnique(prisma.Note.ID.Equals(noteId)).Exec(ctx)
		if errNote != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": errNote.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": data,
		})
	})

	//errRouter := router.Run("0.0.0.0:8080")
	//if errRouter != nil {
	//	return
	//}
}
