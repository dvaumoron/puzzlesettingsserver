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
package main

import (
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/dvaumoron/puzzlesessionservice"
	"github.com/dvaumoron/puzzlesettingsserver/mongoclient"
	"github.com/dvaumoron/puzzlesettingsserver/settingsserver"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	if godotenv.Overload() == nil {
		fmt.Println("Loaded .env file")
	}

	lis, err := net.Listen("tcp", ":"+os.Getenv("SERVICE_PORT"))
	if err != nil {
		log.Fatalf("Failed to listen : %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterSessionServer(s, settingsserver.New(mongoclient.Create()))
	log.Printf("Listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve : %v", err)
	}
}
