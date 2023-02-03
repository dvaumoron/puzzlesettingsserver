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

	mongoclient "github.com/dvaumoron/puzzlemongoclient"
	pb "github.com/dvaumoron/puzzlesessionservice"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collectionName = "settings"

const userIdKey = "userId"
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

func (server) Generate(context.Context, *pb.SessionInfo) (*pb.SessionId, error) {
	return nil, errors.New("method Generate not supported")
}

func (s server) GetSessionInfo(ctx context.Context, request *pb.SessionId) (*pb.SessionInfo, error) {
	client, err := mongo.Connect(ctx, s.clientOptions)
	if err != nil {
		log.Println(mongoCallMsg, err)
		return nil, errInternal
	}
	defer mongoclient.Disconnect(client, ctx)

	collection := client.Database(s.databaseName).Collection(collectionName)
	var result bson.D
	err = collection.FindOne(
		ctx, bson.D{{Key: userIdKey, Value: request.Id}}, optsOnlySettingsField,
	).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &pb.SessionInfo{Info: map[string]string{}}, nil
		}

		log.Println(mongoCallMsg, err)
		return nil, errInternal
	}

	// can call [0] to get settings because result has only one field
	return &pb.SessionInfo{Info: mongoclient.ExtractStringMap(result[0].Value)}, nil
}

func (s server) UpdateSessionInfo(ctx context.Context, request *pb.SessionUpdate) (*pb.Response, error) {
	client, err := mongo.Connect(ctx, s.clientOptions)
	if err != nil {
		log.Println(mongoCallMsg, err)
		return nil, errInternal
	}
	defer mongoclient.Disconnect(client, ctx)

	id := request.Id
	settings := bson.M{userIdKey: id, settingsKey: request.Info}
	collection := client.Database(s.databaseName).Collection(collectionName)
	_, err = collection.ReplaceOne(
		ctx, bson.D{{Key: userIdKey, Value: id}}, settings, optsCreateUnexisting,
	)
	if err != nil {
		log.Println(mongoCallMsg, err)
		return nil, errInternal
	}
	return &pb.Response{Success: true}, nil
}
