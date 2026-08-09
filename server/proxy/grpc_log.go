package proxy

import (
	"context"
	"time"

	"github.com/nusiss-capstone-project/campaign-center-api/server/log"
)

func withGRPCCall(
	ctx context.Context,
	rpc string,
	fields []any,
	fn func(context.Context) error,
) error {
	start := time.Now()
	logger := log.WithContext(ctx)
	logger.Infow("grpc_call_start", append([]any{"rpc", rpc}, fields...)...)

	err := fn(ctx)
	durationMs := float64(time.Since(start).Microseconds()) / 1000
	out := append([]any{"rpc", rpc, "duration_ms", durationMs}, fields...)
	if err != nil {
		logger.Errorw("grpc_call_failed", append(out, "error", err.Error())...)
		return err
	}
	logger.Infow("grpc_call_success", out...)
	return nil
}
