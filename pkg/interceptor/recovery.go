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
	"fmt"
	"runtime"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerRecoveryInterceptor returns a new unary server recovery for panic recovery.
func UnaryServerRecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {

				var msg string
				for i := 2; ; i++ {
					_, file, line, ok := runtime.Caller(i)
					if !ok {
						break
					}
					msg = msg + fmt.Sprintf("%s:%d\n", file, line)
				}
				log.Error().Msgf("%s\n↧↧↧↧↧↧ PANIC ↧↧↧↧↧↧\n%s↥↥↥↥↥↥ PANIC ↥↥↥↥↥↥", r, msg)

				resp = nil
				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		resp, err = handler(ctx, req)
		return
	}
}
