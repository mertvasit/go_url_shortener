package routes

import (
	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber"
	"github.com/google/uuid"
	"github.com/mertvasit/go-url-shortener/database"
	"github.com/mertvasit/go-url-shortener/helpers"
	"os"
	"strconv"
	"time"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate-limit-rest"`
}

func ShortenURL(c *fiber.Ctx) {
	body := new(request)
	if err := c.BodyParser(&body); err != nil {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse json"})
		return
	}

	if body.Expiry == 0 {
		body.Expiry = 24 // default expiry of 24 hours
	}

	redisUrl := database.RedisClient{DB: 0}
	redisUser := database.RedisClient{DB: 1, SetValue: os.Getenv("API_QUOTA"), TTL: 30 * 60 * time.Second}

	// rate limiting
	redisUser.CreateRedisClient()
	defer redisUser.Client.Close()

	ipVisitCount, err := redisUser.Get(c.IP(), true)
	if err != nil {
		c.Status(err.Status).JSON(fiber.Map{"error": err.Error, "message": err.Message})
		return
	}

	ipVisitCountInt, _ := strconv.Atoi(ipVisitCount)
	if ipVisitCountInt <= 0 {
		limit, _ := redisUser.Client.TTL(database.Ctx, c.IP()).Result()
		c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":            "Rate limit exceeded",
			"rate_limit_reset": limit / time.Nanosecond / time.Minute,
		})
		return
	}

	//
	if !govalidator.IsURL(body.URL) {
		c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid url"})
		return
	}

	if !helpers.RemoveDomainError(body.URL) {
		c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "oops"})
		return
	}

	body.URL = helpers.EnforceHttp(body.URL)

	var shortId string
	if body.CustomShort == "" {
		shortId = uuid.New().String()[:6]
	} else {
		shortId = body.CustomShort
	}

	redisUrl.CreateRedisClient()
	defer redisUrl.Client.Close()

	redisUrl.SetValue = body.URL
	redisUrl.TTL = body.Expiry * 3600 * time.Second

	value, _ := redisUrl.Get(shortId, true)
	if value != "" {
		c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "URL short already in use",
		})
		return
	}

	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	_ = redisUser.Decrement(c.IP())

	xRateRemaining, _ := redisUser.Get(c.IP(), false)
	resp.XRateRemaining, _ = strconv.Atoi(xRateRemaining)

	ttlRemaining, _ := redisUser.Client.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = ttlRemaining / time.Nanosecond / time.Minute

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + shortId

	c.Status(fiber.StatusOK).JSON(resp)
	return
}
