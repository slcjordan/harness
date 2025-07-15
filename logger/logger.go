package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Init() {
	cfg := zap.NewProductionEncoderConfig()
	encoder := zapcore.NewConsoleEncoder(cfg)

	infoLevel := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= zapcore.DebugLevel && l < zapcore.ErrorLevel
	})
	errorLevel := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= zapcore.ErrorLevel
	})
	infoCore := zapcore.NewCore(encoder, os.Stdout, infoLevel)
	errorCore := zapcore.NewCore(encoder, os.Stderr, errorLevel)

	logger = zap.New(zapcore.NewTee(infoCore, errorCore))
}

var logger *zap.Logger

type contextKey struct{}

type contextValue struct {
	Fields []zap.Field
}

func getContext(ctx context.Context) contextValue {
	val, ok := ctx.Value(contextKey{}).(contextValue)
	if !ok {
		return contextValue{}
	}
	return val
}

func With(ctx context.Context, key string, raw any) context.Context {
	v := getContext(ctx)
	switch val := raw.(type) {
	case zapcore.ArrayMarshaler:
		v.Fields = append(v.Fields, zap.Array(key, val))
	case []byte:
		v.Fields = append(v.Fields, zap.Binary(key, val))
	case bool:
		v.Fields = append(v.Fields, zap.Bool(key, val))
	case *bool:
		v.Fields = append(v.Fields, zap.Boolp(key, val))
	case []bool:
		v.Fields = append(v.Fields, zap.Bools(key, val))
	case [][]byte:
		v.Fields = append(v.Fields, zap.ByteStrings(key, val))
	case complex128:
		v.Fields = append(v.Fields, zap.Complex128(key, val))
	case *complex128:
		v.Fields = append(v.Fields, zap.Complex128p(key, val))
	case []complex128:
		v.Fields = append(v.Fields, zap.Complex128s(key, val))
	case complex64:
		v.Fields = append(v.Fields, zap.Complex64(key, val))
	case *complex64:
		v.Fields = append(v.Fields, zap.Complex64p(key, val))
	case []complex64:
		v.Fields = append(v.Fields, zap.Complex64s(key, val))
	case time.Duration:
		v.Fields = append(v.Fields, zap.Duration(key, val))
	case *time.Duration:
		v.Fields = append(v.Fields, zap.Durationp(key, val))
	case []time.Duration:
		v.Fields = append(v.Fields, zap.Durations(key, val))
	case error:
		v.Fields = append(v.Fields, zap.NamedError(key, val))
	case []error:
		v.Fields = append(v.Fields, zap.Errors(key, val))
	case float32:
		v.Fields = append(v.Fields, zap.Float32(key, val))
	case *float32:
		v.Fields = append(v.Fields, zap.Float32p(key, val))
	case []float32:
		v.Fields = append(v.Fields, zap.Float32s(key, val))
	case float64:
		v.Fields = append(v.Fields, zap.Float64(key, val))
	case *float64:
		v.Fields = append(v.Fields, zap.Float64p(key, val))
	case []float64:
		v.Fields = append(v.Fields, zap.Float64s(key, val))
	case int:
		v.Fields = append(v.Fields, zap.Int(key, val))
	case int16:
		v.Fields = append(v.Fields, zap.Int16(key, val))
	case *int16:
		v.Fields = append(v.Fields, zap.Int16p(key, val))
	case []int16:
		v.Fields = append(v.Fields, zap.Int16s(key, val))
	case int32:
		v.Fields = append(v.Fields, zap.Int32(key, val))
	case *int32:
		v.Fields = append(v.Fields, zap.Int32p(key, val))
	case []int32:
		v.Fields = append(v.Fields, zap.Int32s(key, val))
	case int64:
		v.Fields = append(v.Fields, zap.Int64(key, val))
	case *int64:
		v.Fields = append(v.Fields, zap.Int64p(key, val))
	case []int64:
		v.Fields = append(v.Fields, zap.Int64s(key, val))
	case int8:
		v.Fields = append(v.Fields, zap.Int8(key, val))
	case *int8:
		v.Fields = append(v.Fields, zap.Int8p(key, val))
	case []int8:
		v.Fields = append(v.Fields, zap.Int8s(key, val))
	case *int:
		v.Fields = append(v.Fields, zap.Intp(key, val))
	case []int:
		v.Fields = append(v.Fields, zap.Ints(key, val))
	case zapcore.ObjectMarshaler:
		v.Fields = append(v.Fields, zap.Object(key, val))
	case string:
		v.Fields = append(v.Fields, zap.String(key, val))
	case fmt.Stringer:
		v.Fields = append(v.Fields, zap.Stringer(key, val))
	case *string:
		v.Fields = append(v.Fields, zap.Stringp(key, val))
	case []string:
		v.Fields = append(v.Fields, zap.Strings(key, val))
	case time.Time:
		v.Fields = append(v.Fields, zap.Time(key, val))
	case *time.Time:
		v.Fields = append(v.Fields, zap.Timep(key, val))
	case []time.Time:
		v.Fields = append(v.Fields, zap.Times(key, val))
	case uint:
		v.Fields = append(v.Fields, zap.Uint(key, val))
	case uint16:
		v.Fields = append(v.Fields, zap.Uint16(key, val))
	case *uint16:
		v.Fields = append(v.Fields, zap.Uint16p(key, val))
	case []uint16:
		v.Fields = append(v.Fields, zap.Uint16s(key, val))
	case uint32:
		v.Fields = append(v.Fields, zap.Uint32(key, val))
	case *uint32:
		v.Fields = append(v.Fields, zap.Uint32p(key, val))
	case []uint32:
		v.Fields = append(v.Fields, zap.Uint32s(key, val))
	case uint64:
		v.Fields = append(v.Fields, zap.Uint64(key, val))
	case *uint64:
		v.Fields = append(v.Fields, zap.Uint64p(key, val))
	case []uint64:
		v.Fields = append(v.Fields, zap.Uint64s(key, val))
	case uint8:
		v.Fields = append(v.Fields, zap.Uint8(key, val))
	case *uint8:
		v.Fields = append(v.Fields, zap.Uint8p(key, val))
	case *uint:
		v.Fields = append(v.Fields, zap.Uintp(key, val))
	case uintptr:
		v.Fields = append(v.Fields, zap.Uintptr(key, val))
	case *uintptr:
		v.Fields = append(v.Fields, zap.Uintptrp(key, val))
	case []uintptr:
		v.Fields = append(v.Fields, zap.Uintptrs(key, val))
	case []uint:
		v.Fields = append(v.Fields, zap.Uints(key, val))
	default:
		v.Fields = append(v.Fields, zap.Any(key, val))
	}
	return context.WithValue(ctx, contextKey{}, v)
}

func Infof(ctx context.Context, f string, a ...any) {
	logger.Info(fmt.Sprintf(f, a...), getContext(ctx).Fields...)
}

func Errorf(ctx context.Context, f string, a ...any) {
	logger.Error(fmt.Sprintf(f, a...), getContext(ctx).Fields...)
}
