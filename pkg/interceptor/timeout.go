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
	"time"

	"google.golang.org/grpc"
)

// UnaryClientInterceptor returns a new unary client interceptor that sets a timeout on the request context.
func UnaryClientInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		timedCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		return invoker(timedCtx, method, req, reply, cc, opts...)
	}
}
