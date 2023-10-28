package rawmaterial

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RawMaterial struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RawMatID string             `bson:"rawmatid" json:"rawmatid"`
	Name     string             `bson:"name" json:"name"`
	Type     string             `bson:"type" json:"type"`
	Status   bool               `bson:"status" json:"status"`
	Detail   []bson.M           `bson:"detail" json:"detail"`
	ImageID  string             `bson:"imageid" json:"imageid"`

	CreatedTime string `json:"created_time" bson:"created_time"`
	UpdatedTime string `json:"updated_time" bson:"updated_time"`
}

func (m *RawMaterial) Create(ctx context.Context, db *mongo.Database) (interface{}, error) {
	collection := db.Collection("RawMat")

	res, err := collection.InsertOne(ctx, m)
	if err != nil {
		return nil, err
	}

	m.ID = res.InsertedID.(primitive.ObjectID)
	return m.ID, nil
}

func (doc *RawMaterial) Update(ctx context.Context, db *mongo.Database, id primitive.ObjectID) (interface{}, error) {
	collection := db.Collection("RawMat")
	filter := bson.M{"_id": id}
	update := bson.D{{"$set", bson.D{{"rawmatid", doc.RawMatID}, {"name", doc.Name}, {"type", doc.Type}, {"status", doc.Status}, {"detail", doc.Detail}, {"imageid", doc.ImageID}, {"updated_time", doc.UpdatedTime}}}}
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func Delete(ctx context.Context, db *mongo.Database, id primitive.ObjectID) (interface{}, error) {
	collection := db.Collection("RawMat")
	filter := bson.M{"_id": id}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return nil, err
	}
	if result.DeletedCount == 0 {
		fmt.Println("no matching document found")
	}
	fmt.Printf("deleted %v document(s)\n", result.DeletedCount)
	return result, nil
}
