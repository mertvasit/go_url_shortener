package routes

import (
	"github.com/gofiber/fiber"
	"github.com/mertvasit/go-url-shortener/database"
)

func ResolveURL(c *fiber.Ctx) {
	url := c.Params("url")

	redisUrl := database.RedisClient{DB: 0}
	redisUser := database.RedisClient{DB: 1}

	redisUrl.CreateRedisClient()
	defer redisUrl.Client.Close()

	value, err := redisUrl.Get(url, true)
	if err != nil {
		c.Status(err.Status).JSON(fiber.Map{"error": err.Error, "message": err.Message})
		return
	}

	redisUser.CreateRedisClient()
	defer redisUser.Client.Close()

	_ = redisUser.Increment("counter")

	c.Redirect(value, fiber.StatusMovedPermanently)
	return
}
