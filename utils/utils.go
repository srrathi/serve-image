package utils

import (
	"log"
	"time"

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
