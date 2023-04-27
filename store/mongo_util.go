package store

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DataUpdateSertWithTime(collection string, uidObj primitive.ObjectID, sets *bson.D, incs *bson.D, unsets *bson.D, pulls *bson.D, pushs *bson.D, pops *bson.D) error {
	err := DataUpdateSert(collection, uidObj, sets, incs, unsets, pulls, pushs, pops)
	return err
}

func DataUpdateSert(collection string, uidObj primitive.ObjectID, sets *bson.D, incs *bson.D, unsets *bson.D, pulls *bson.D, pushs *bson.D, pops *bson.D) error {
	update := bson.D{}

	if sets != nil {
		update = append(update, bson.E{"$set", sets})
	}

	if incs != nil {
		update = append(update, bson.E{"$inc", incs})
	}

	if unsets != nil {
		update = append(update, bson.E{"$unset", unsets})
	}

	if pulls != nil {
		update = append(update, bson.E{"$pull", pulls})
	}

	if pushs != nil {
		update = append(update, bson.E{"$push", pushs})
	}
	if pops != nil {
		update = append(update, bson.E{"$pop", pops})
	}

	opts := options.Update().SetUpsert(true)
	filter := bson.D{{"_id", uidObj}}
	_, err := GetColl(collection).UpdateOne(context.TODO(), filter, update, opts)
	return err
}

func DataUpdateSertSets(collection string, uidObj primitive.ObjectID, sets *bson.D) error {
	return DataUpdateSert(collection, uidObj, sets, nil, nil, nil, nil, nil)
}

func DataInrSertSets(collection string, uidObj primitive.ObjectID, incs *bson.D) error {
	return DataUpdateSert(collection, uidObj, nil, incs, nil, nil, nil, nil)
}

func DataGetType(collection string, uidObj primitive.ObjectID, fields []string, result interface{}) error {
	projection := bson.D{{"_id", 0}}
	for _, field := range fields {
		e := bson.E{Key: field, Value: 1}
		projection = append(projection, e)
	}
	filter := bson.D{{"_id", uidObj}}

	opts := options.FindOne().SetProjection(projection)
	err := GetColl(collection).FindOne(context.TODO(), filter, opts).Decode(result)
	return err
}

func DataRemove(collection string, uidObj primitive.ObjectID) error {
	filter := bson.D{{"_id", uidObj}}
	_, err := GetColl(collection).DeleteOne(context.TODO(), filter)
	return err
}

func DataGet(collection string, uidObj primitive.ObjectID, fields ...string) (bson.M, error) {
	result := bson.M{}

	projection := bson.D{{"_id", 0}}
	for _, field := range fields {
		e := bson.E{Key: field, Value: 1}
		projection = append(projection, e)
	}
	filter := bson.D{{"_id", uidObj}}

	opts := options.FindOne().SetProjection(projection)
	err := GetColl(collection).FindOne(context.TODO(), filter, opts).Decode(&result)
	return result, err
}

const db_timeout = time.Millisecond * 500

func DataGetMore(collection string, uidObj []primitive.ObjectID, fields []string, result interface{}) error {
	projection := bson.D{}
	for _, field := range fields {
		e := bson.E{Key: field, Value: 1}
		projection = append(projection, e)
	}

	filter := bson.D{{"_id", bson.D{{"$in", uidObj}}}}

	opts := options.Find().SetProjection(projection)

	cursor, err := GetColl(collection).Find(context.TODO(), filter, opts)
	if err != nil {
		return err
	}

	err = cursor.All(context.TODO(), result)
	if err != nil {
		return err
	}
	cursor.Close(context.TODO())

	return nil
}

func DataGetMoreByUIDStr(collection string, uidObj []string, fields []string, result interface{}) error {
	uidObjList := []primitive.ObjectID{}
	for _, v := range uidObj {
		uidObj, err := primitive.ObjectIDFromHex(v)
		if err != nil {
			return err
		}
		uidObjList = append(uidObjList, uidObj)
	}
	err := DataGetMore(collection, uidObjList, fields, result)
	return err
}

func DataAddToSets(collection string, objectID primitive.ObjectID, fields string, value string) error {
	update := bson.D{bson.E{"$addToSet", bson.D{{fields, value}}}}
	opts := options.Update().SetUpsert(true)
	filter := bson.D{{"_id", objectID}}
	_, err := GetColl(collection).UpdateOne(context.TODO(), filter, update, opts)
	return err
}
