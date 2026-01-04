package handler

import (
	"context"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/bayrambartu/taxihub-driver-service/model"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var validate = validator.New()

type DriverHandle struct {
	Collection *mongo.Collection
}

// CreateDriver godoc
// @Summary Create new driver
// @Description Create a new taxi driver
// @Tags Drivers
// @Accept json
// @Produce json
// @Param driver body model.CreateDriverDTO true "Driver information"
// @Success 201 {object} map[string]interface{}
// @Router /drivers [post]
func (h *DriverHandle) CreateDriver(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var dto model.CreateDriverDTO

	// get the incoming json
	if err := c.BodyParser(&dto); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid data format"})
	}

	if err := validate.Struct(dto); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	newDriver := model.Driver{
		ID:        primitive.NewObjectID(),
		FirstName: dto.FirstName,
		LastName:  dto.LastName,
		Plate:     dto.Plate,
		TaxiType:  dto.TaxiType,
		CarBrand:  dto.CarBrand,
		CarModel:  dto.CarModel,
		Location: model.Location{
			Lat: dto.Lat,
			Lon: dto.Lon,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// save the mongodb
	_, err := h.Collection.InsertOne(ctx, newDriver)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "not saved to database"})
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "saved successfully",
		"id":      newDriver.ID,
	})
}

// ListDriver godoc
// @Summary List drivers
// @Description List all drivers with pagination
// @Tags Drivers
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {array} model.Driver
// @Failure 500 {object} map[string]interface{}
// @Router /drivers [get]
func (h *DriverHandle) ListDriver(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.Query("pageSize", "20"))
	if err != nil || pageSize < 1 {
		pageSize = 20
	}

	// pagination logic

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)
	findOptions := options.Find().SetSkip(skip).SetLimit(limit)

	// take the datas
	cursor, err := h.Collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to retrieve data"})
	}
	defer cursor.Close(ctx)

	var drivers []model.Driver = make([]model.Driver, 0)
	if err := cursor.All(ctx, &drivers); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to convert data "})
	}
	return c.Status(http.StatusOK).JSON(drivers)

}

// UpdateDriver godoc
// @Summary Update driver
// @Description Update driver information by ID
// @Tags Drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param driver body map[string]interface{} true "Fields to update"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers/{id} [put]
func (h *DriverHandle) UpdateDriver(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// take the id and confirm
	idParam := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(idParam)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	var updateData map[string]interface{}
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	updateData["updatedAt"] = time.Now()

	update := bson.M{"$set": updateData}
	result, err := h.Collection.UpdateOne(ctx, bson.M{"_id": objID}, update)

	if err != nil || result.MatchedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "driver cant find"})
	}
	return c.JSON(fiber.Map{"message": "drivers updated successfully"})

}

func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371
	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLon := (lon2 - lon1) * (math.Pi / 180)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180))*math.Cos(lat2*(math.Pi/180))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// GetNearbyDrivers godoc
// @Summary List nearby taxis
// @Description Returns taxis within 6 km sorted by distance
// @Tags Drivers
// @Accept json
// @Produce json
// @Param lat query float64 true "Latitude"
// @Param lon query float64 true "Longitude"
// @Param taxiType query string true "Taxi type (sari, turkuaz, siyah)"
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /drivers/nearby [get]
func (h *DriverHandle) GetNearbyDrivers(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userLat, _ := strconv.ParseFloat(c.Query("lat"), 64)
	userLon, _ := strconv.ParseFloat(c.Query("lon"), 64)
	taxiType := c.Query("taxiType")

	filter := bson.M{"taxiType": taxiType}
	cursor, err := h.Collection.Find(ctx, filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query error"})
	}

	var allDrivers []model.Driver
	cursor.All(ctx, &allDrivers)

	type NearbyResponse struct {
		FirstName  string  `json:"firstName"`
		LastName   string  `json:"lastName"`
		Plate      string  `json:"plate"`
		DistanceKm float64 `json:"distanceKm"`
	}

	var results []NearbyResponse
	for _, d := range allDrivers {
		dist := calculateDistance(userLat, userLon, d.Location.Lat, d.Location.Lon)
		if dist <= 6.0 {
			results = append(results, NearbyResponse{
				FirstName:  d.FirstName,
				LastName:   d.LastName,
				Plate:      d.Plate,
				DistanceKm: math.Round(dist*100) / 100,
			})
		}
	}

	// strings -> sort
	sort.Slice(results, func(i, j int) bool {
		return results[i].DistanceKm < results[j].DistanceKm
	})
	return c.JSON(results)
}
