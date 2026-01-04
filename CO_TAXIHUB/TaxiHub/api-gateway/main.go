package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecretKey = "jwtsecretkey"

func main() {
	// err := godotenv.Load("../.env")
	// if err != nil {
	// 	log.Fatal("error loading .env file")
	// }
	// jwtSecretKey := os.Getenv("JWT_SECRET_KEY")
	// if jwtSecretKey == "" {
	// 	log.Fatal("JWT_SECRET_KEY is not set")
	// }

	app := fiber.New()

	//logging every request
	app.Use(logger.New())


	// rate limiter middleware
	app.Use(limiter.New(limiter.Config{
		Max:        20,
		Expiration: 30 * time.Second,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{"error": "too many request, please wait"})
		},
	}))

	// Login endpoint to generate JWT
	app.Post("/login", func(c *fiber.Ctx) error {

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user": "bartu",
			"exp":  time.Now().Add(time.Hour * 4).Unix(),
		})

	    // Sign token with secret key
		t, err := token.SignedString([]byte(jwtSecretKey))
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		// return token
		return c.JSON(fiber.Map{"token": t})
	})
	
	// JWT authentication middleware
	authMiddleware := func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		// if authHeader == "" {
		// 	return c.Status(401).JSON(fiber.Map{"error": "token is required"})
		// }
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return c.Status(401).JSON(fiber.Map{"error": "Token is required"})
		}
		tokenString := authHeader[7:]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(jwtSecretKey), nil
		})
		if err != nil || !token.Valid {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid or expired token"})
		}
		return c.Next()
	}
    // Proxy all driver service request
	app.All("/api/v1/drivers/*", authMiddleware, func(c *fiber.Ctx) error {
		driverServiceURL := os.Getenv("DRIVER_SERVICE_URL")

		if driverServiceURL == "" {
			driverServiceURL = "http://localhost:5001"
		}

		path := c.Path()
		queryString := string(c.Request().URI().QueryString())

		targetURL := driverServiceURL + path
		if queryString != "" {
			targetURL += "?" + queryString
		}
		log.Printf("Routing: %s", targetURL)

		return proxy.Do(c, targetURL)

	})

	log.Fatal(app.Listen(":5002"))
}
