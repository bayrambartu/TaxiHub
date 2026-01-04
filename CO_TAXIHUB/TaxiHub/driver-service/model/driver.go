package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Location struct {
	Lat float64 `json:"lat" bson:"lat"`
	Lon float64 `json:"lon" bson:"lon"`
}

type Driver struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	FirstName string             `json:"firstName" bson:"firstName"`
	LastName  string             `json:"lastName" bson:"lastName"`
	Plate     string             `json:"plate" bson:"plate"`
	TaxiType  string             `json:"taxiType" bson:"taxiType"`
	CarBrand  string             `json:"carBrand" bson:"carBrand"`
	CarModel  string             `json:"carModel" bson:"carModel"`
	Location  Location           `json:"location" bson:"location"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type CreateDriverDTO struct {
	FirstName string  `json:"firstName" validate:"required,min=2"`
	LastName  string  `json:"lastName" validate:"required,min=2"`
	Plate     string  `json:"plate" validate:"required"`
	TaxiType  string  `json:"taxiType" validate:"required,oneof=sari turkuaz siyah"`
	CarBrand  string  `json:"carBrand" validate:"required"`
	CarModel  string  `json:"carModel" validate:"required"`
	Lat       float64 `json:"lat" validate:"required"`
	Lon       float64 `json:"lon" validate:"required"`
}

// type UpdateDriverDTO struct {
// 	FirstName *string  `json:"firstName"`
// 	LastName  *string  `json:"lastName"`
// 	Plate     *string  `json:"plate"`
// 	TaxiType  *string  `json:"taksiType"`
// 	CarBrand  *string  `json:"carBrand"`
// 	CarModel  *string  `json:"carModel"`
// 	Lat       *float64 `json:"lat"`
// 	Lon       *float64 `json:"lon"`
// 	UpdatedAt time.Time
// }
