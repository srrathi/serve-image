package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	database "github.com/srrathi/image-server/db"
	"github.com/srrathi/image-server/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client instance
var DB *mongo.Client = database.ConnectDB()
var name = "photos"
var opt = options.GridFSBucket().SetName(name)
var Cache = utils.NewCache()

// Getting database collections
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("image-server").Collection(collectionName)
	return collection
}

// Upload image handler
func uploadImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, header, err := c.Request.FormFile("image")
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}
		defer file.Close()

		bucket, err := gridfs.NewBucket(
			DB.Database("image-server"), opt,
		)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, file); err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		filename := time.Now().Format(time.RFC3339) + "_" + header.Filename
		uploadStream, err := bucket.OpenUploadStream(
			filename,
		)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}
		defer uploadStream.Close()

		fileSize, err := uploadStream.Write(buf.Bytes())
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		fileId, err := json.Marshal(uploadStream.FileID)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		Cache.Update(string(fileId), buf.Bytes())
		c.JSON(http.StatusOK, map[string]interface{}{"fileId": strings.Trim(string(fileId), `"`), "fileSize": fileSize})
	}
}

// Serving image over http REST handler
func serveImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		imageId := strings.TrimPrefix(c.Request.URL.Path, "/image/")
		img, valid := Cache.Read(imageId)
		if valid {
			imgContentType := http.DetectContentType(img)
			if imgContentType == "text/plain; charset=utf-8" {
				log.Println(string(img))
				c.JSON(http.StatusBadRequest, string(img))
				return
			}

			bytesData := img
			contentType := http.DetectContentType(bytesData)

			c.Writer.Header().Add("Content-Type", contentType)
			c.Writer.Header().Add("Content-Length", strconv.Itoa(len(bytesData)))

			c.Writer.Write(bytesData)
			return
		}

		objID, err := primitive.ObjectIDFromHex(imageId)
		if err != nil {
			log.Println(err)
			Cache.Update(imageId, []byte(err.Error()))
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		bucket, _ := gridfs.NewBucket(
			DB.Database("image-server"), opt,
		)

		var buf bytes.Buffer
		dStream, err := bucket.DownloadToStream(objID, &buf)
		if err != nil {
			log.Println(err)
			Cache.Update(imageId, []byte(err.Error()))
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		log.Printf("File size to download: %v kb\n", dStream/1024)
		bytesData := buf.Bytes()
		contentType := http.DetectContentType(bytesData)

		Cache.Update(imageId, bytesData)
		c.Writer.Header().Add("Content-Type", contentType)
		c.Writer.Header().Add("Content-Length", strconv.Itoa(len(bytesData)))

		c.Writer.Write(bytesData)
	}
}

// Home endpoint to check server status
func homePage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, "Image server working fine")
	}
}

func Route(router *gin.Engine) {
	//All routes will be added here
	router.GET("/", homePage())
	router.POST("/upload", uploadImage())
	router.GET("/image/:imageId", serveImage())
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}
	router := gin.Default()
	Route(router)

	router.Run(":" + port)
}
