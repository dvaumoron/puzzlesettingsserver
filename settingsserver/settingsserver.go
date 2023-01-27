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
		return nil, err
	}
	defer disconnect(client, ctx)

	collection := client.Database(s.databaseName).Collection(collectionName)
	var result bson.M
	err = collection.FindOne(ctx, idFilter(in.Id)).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &pb.SessionInfo{Info: map[string]string{}}, nil
		}
		return nil, err
	}

	info := map[string]string{}
	for k, v := range result {
		str, _ := v.(string)
		info[k] = str
	}
	return &pb.SessionInfo{Info: info}, nil
}

func (s server) UpdateSessionInfo(ctx context.Context, in *pb.SessionUpdate) (*pb.SessionError, error) {
	client, err := mongo.Connect(ctx, s.clientOptions)
	if err != nil {
		return nil, err
	}
	defer disconnect(client, ctx)

	info := bson.M{}
	for k, v := range in.Info {
		info[k] = v
	}
	collection := client.Database(s.databaseName).Collection(collectionName)
	_, err = collection.ReplaceOne(ctx, idFilter(in.Id), info, optsCreateUnexisting)
	if err != nil {
		return &pb.SessionError{Err: err.Error()}, nil
	}
	return &pb.SessionError{}, nil
}

func disconnect(client *mongo.Client, ctx context.Context) {
	if err := client.Disconnect(ctx); err != nil {
		log.Print("Error during Disconnect :", err)
	}
}

func idFilter(id uint64) bson.D {
	return bson.D{{Key: "_id", Value: id}}
}
