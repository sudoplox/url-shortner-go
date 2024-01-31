package routes

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/sudoplox/url-shortner-go/database"
	"github.com/sudoplox/url-shortner-go/helpers"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func ShortenURL(c *fiber.Ctx) error {
	body := new(request)

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "cannot parse JSON",
			},
		)
	}

	// implement rate limit
	// check if IP has already been there in DB

	r2 := database.CreateClient(1)
	defer r2.Close()

	val, err := r2.Get(database.Ctx, c.IP()).Result()
	if err == redis.Nil {
		// If not found in DB -> new user
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*time.Minute).Err()
	} else if err != nil {
		fmt.Println(err.Error())
		return c.Status(fiber.StatusServiceUnavailable).JSON(
			fiber.Map{
				"error": "Redis Error: " + err.Error(),
			},
		)
	} else {
		// Found the user IP in DB
		val, _ := r2.Get(database.Ctx, c.IP()).Result()
		// Get the API QUOTA for the IP -> how many api calls left
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(
				fiber.Map{
					"error":            "rate limit exceeded",
					"rate_limit_reset": limit / time.Nanosecond / time.Minute,
				},
			)
		}
	}

	// check if the input is an actual url or not
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "invalid URL",
			},
		)
	}
	// check for domain error
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(
			fiber.Map{
				"error": "haha... nice try",
			},
		)
	}
	// enforce https -> SSL
	body.URL = helpers.EnforceHTTP(body.URL)

	var id string

	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	r := database.CreateClient(0)
	defer r.Close()

	val, _ = r.Get(database.Ctx, id).Result()
	if val != "" {
		// something is found in the db for the custom short
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{
				"error": "url custom short is already in use",
			},
		)
	}

	// check expiry or set to default expiry of 24 hours
	if body.Expiry == 0 {
		body.Expiry = 24
	}

	rErr := r.Set(database.Ctx, id, body.URL, body.Expiry*time.Hour)
	if rErr != nil {
		c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{
				"error": "unable to connect to the server",
			},
		)
	}

	// Default response
	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	r2.Decr(database.Ctx, c.IP())

	val, _ = r2.Get(database.Ctx, c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	return c.Status(fiber.StatusOK).JSON(resp)
}
