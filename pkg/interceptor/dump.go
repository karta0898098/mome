/*
 * Copyright (c) 2023.
 * D-Link Corporation.
 * All rights reserved.
 *
 * The information contained herein is confidential and proprietary to
 * D-Link. Use of this information by anyone other than authorized employees
 * of D-Link is granted only under a written non-disclosure agreement,
 * expressly prescribing the scope and manner of such use.
 */

package interceptor

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// requestDump log grpc request dump
func requestDump(ctx context.Context, info *grpc.UnaryServerInfo, logger *zerolog.Logger, msg interface{}, err error) {
	dict := zerolog.Dict()

	// load header form context
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		header, err := json.Marshal(&md)
		if err == nil {
			dict.RawJSON("header", header)
		}
	}

	protoMsg, ok := msg.(proto.Message)
	if ok {
		buf, err := protojson.Marshal(protoMsg)
		if err == nil {
			dict.RawJSON("body", buf)
			logger.Info().Dict("dump", dict).Msg("grpc request dump.")
		}
	}
}

// replayDump log grpc reply dump
func replayDump(ctx context.Context, info *grpc.UnaryServerInfo, logger *zerolog.Logger, msg interface{}, err error) {
	protoMsg, ok := msg.(proto.Message)
	if ok {
		buf, err := protojson.Marshal(protoMsg)
		if err == nil {
			logger.Info().Dict("dump", zerolog.Dict().RawJSON("body", buf)).Msg("grpc replay dump.")
		}
	}
}
