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
	grpcserver "github.com/dvaumoron/puzzlegrpcserver"
	mongoclient "github.com/dvaumoron/puzzlemongoclient"
	pb "github.com/dvaumoron/puzzlesessionservice"
	"github.com/dvaumoron/puzzlesettingsserver/settingsserver"
)

func main() {
	s := grpcserver.Make()
	clientOptions, databaseName := mongoclient.Create()
	pb.RegisterSessionServer(s, settingsserver.New(clientOptions, databaseName, s.Logger))
	s.Start()
}
