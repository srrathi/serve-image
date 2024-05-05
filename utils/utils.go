package utils

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"mime/multipart"
	"time"

	webp "github.com/chai2010/webp"
	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
)

type allCache struct {
	images *cache.Cache
}

const (
	defaultExpiration = 11 * time.Minute
	purgeTime         = 13 * time.Minute
)

func NewCache() *allCache {
	Cache := cache.New(defaultExpiration, purgeTime)
	return &allCache{
		images: Cache,
	}
}

func (c *allCache) Read(id string) (item []byte, ok bool) {
	image, ok := c.images.Get(id)
	if ok {
		log.Println("from cache")
		res, valid := image.([]byte)
		if !valid {
			return nil, false
		}
		return res, true
	}
	return nil, false
}

func (c *allCache) Update(id string, image []byte) {
	c.images.Set(id, image, cache.DefaultExpiration)
}

func LoadEnvFile() string {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
		return "Error loading .env file"
	}
	log.Println(".env file loaded")
	return ".env file loaded"
}

func ConvertImageToWebp(file multipart.File) (*bytes.Buffer, error) {
	// Decode the image from multipart file
	if _, err := file.Seek(0, 0); err != nil {
		if err != nil {
			return nil, err
		}
	}
	m, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	buff := bytes.NewBuffer(nil)
	err = webp.Encode(buff, m, &webp.Options{Lossless: true})
	if err != nil {
		return nil, err
	}
	return buff, nil
}
