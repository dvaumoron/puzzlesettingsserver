/*
 *
 * Copyright 2023 puzzlesettingsserver authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
package settingsserver

import (
	"context"
	"errors"
	"log"

	pb "github.com/dvaumoron/puzzlesessionservice"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collectionName = "settings"

const idKey = "userId"
const settingsKey = collectionName // currently the same

const mongoCallMsg = "Failed during MongoDB call :"

var errInternal = errors.New("internal service error")

var optsOnlySettingsField = options.FindOne().SetProjection(bson.D{{Key: settingsKey, Value: true}})
var optsCreateUnexisting = options.Replace().SetUpsert(true)

// server is used to implement puzzlesessionservice.SessionServer
type server struct {
	pb.UnimplementedSessionServer
	clientOptions *options.ClientOptions
	databaseName  string
}

func New(clientOptions *options.ClientOptions, databaseName string) pb.SessionServer {
	return server{clientOptions: clientOptions, databaseName: databaseName}
}

func (s server) Generate(ctx context.Context, in *pb.SessionInfo) (*pb.SessionId, error) {
	return nil, errors.New("method Generate not supported")
}

func (s server) GetSessionInfo(ctx context.Context, in *pb.SessionId) (*pb.SessionInfo, error) {
	client, err := mongo.Connect(ctx, s.clientOptions)
	if err != nil {
		log.Println(mongoCallMsg, err)
		return nil, errInternal
	}
	defer disconnect(client, ctx)

	collection := client.Database(s.databaseName).Collection(collectionName)
	var result bson.D
	err = collection.FindOne(
		ctx, bson.D{{Key: idKey, Value: in.Id}}, optsOnlySettingsField,
	).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &pb.SessionInfo{Info: map[string]string{}}, nil
		}
		log.Println(mongoCallMsg, err)
		return nil, errInternal
	}

	info := map[string]string{}
	// can call [0] because result has only one field
	settings, _ := result[0].Value.(bson.D)
	for _, e := range settings {
		str, _ := e.Value.(string)
		info[e.Key] = str
	}
	return &pb.SessionInfo{Info: info}, nil
}

func (s server) UpdateSessionInfo(ctx context.Context, in *pb.SessionUpdate) (*pb.Response, error) {
	client, err := mongo.Connect(ctx, s.clientOptions)
	if err != nil {
		log.Println(mongoCallMsg, err)
		return nil, errInternal
	}
	defer disconnect(client, ctx)

	id := in.Id
	info := bson.M{}
	for k, v := range in.Info {
		info[k] = v
	}
	settings := bson.M{idKey: id, settingsKey: info}
	collection := client.Database(s.databaseName).Collection(collectionName)
	_, err = collection.ReplaceOne(
		ctx, bson.D{{Key: idKey, Value: id}}, settings, optsCreateUnexisting,
	)
	if err != nil {
		log.Println(mongoCallMsg, err)
		return nil, errInternal
	}
	return &pb.Response{Success: true}, nil
}

func disconnect(client *mongo.Client, ctx context.Context) {
	if err := client.Disconnect(ctx); err != nil {
		log.Print("Error during MongoDB disconnect :", err)
	}
}
