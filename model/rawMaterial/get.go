package rawmaterial

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func List(ctx context.Context, db *mongo.Database) (*[]*RawMaterial, error) {
	collection := db.Collection("RawMat")
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	records := make([]*RawMaterial, 0)
	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}
	return &records, nil
}

func Read(ctx context.Context, db *mongo.Database, id primitive.ObjectID) (*RawMaterial, error) {
	collection := db.Collection("RawMat")
	res := collection.FindOne(ctx, bson.M{"_id": id})
	var record RawMaterial
	err := res.Decode(&record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}
