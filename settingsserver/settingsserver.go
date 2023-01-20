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
const idKey = "_id"

var optsCreateUnexisting = options.Replace().SetUpsert(true)

// Server is used to implement puzzlesessionservice.SessionServer
type Server struct {
	pb.UnimplementedSessionServer
	clientOptions *options.ClientOptions
	databaseName  string
}

func New(clientOptions *options.ClientOptions, databaseName string) *Server {
	return &Server{clientOptions: clientOptions, databaseName: databaseName}
}

func (s *Server) Generate(ctx context.Context, in *pb.SessionInfo) (*pb.SessionId, error) {
	return nil, errors.New("method Generate not supported")
}

func (s *Server) GetSessionInfo(ctx context.Context, in *pb.SessionId) (*pb.SessionInfo, error) {
	client, err := mongo.Connect(ctx, s.clientOptions)
	var sessionInfo *pb.SessionInfo
	if err == nil {
		defer disconnect(client, ctx)

		collection := client.Database(s.databaseName).Collection(collectionName)
		var result bson.M
		err = collection.FindOne(ctx, bson.D{{Key: idKey, Value: in.Id}}).Decode(&result)
		if err == nil {
			info := map[string]string{}
			for k, v := range result {
				str, _ := v.(string)
				info[k] = str
			}
		} else if err == mongo.ErrNoDocuments {
			sessionInfo = &pb.SessionInfo{Info: map[string]string{}}
			err = nil
		}
	}
	return sessionInfo, err
}

func (s *Server) UpdateSessionInfo(ctx context.Context, in *pb.SessionUpdate) (*pb.SessionError, error) {
	client, err := mongo.Connect(ctx, s.clientOptions)
	errStr := ""
	if err == nil {
		defer disconnect(client, ctx)

		info := bson.M{}
		for k, v := range in.Info {
			info[k] = v
		}
		collection := client.Database(s.databaseName).Collection(collectionName)
		filter := bson.D{{Key: idKey, Value: in.Id}}
		_, err = collection.ReplaceOne(ctx, filter, info, optsCreateUnexisting)
	}
	if err != nil {
		errStr = err.Error()
	}
	return &pb.SessionError{Err: errStr}, nil
}

func disconnect(client *mongo.Client, ctx context.Context) {
	if err := client.Disconnect(ctx); err != nil {
		log.Print("Error during Disconnect :", err)
	}
}
