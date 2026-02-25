package metadatautil

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const (
	keyRealIP  = "real_ip"
	keyHash256 = "hash_256"
)

// SetRealIP устанавливает в контекст IP-адрес.
func SetRealIP(ctx context.Context, ip string) context.Context {
	return SetKey(ctx, keyRealIP, ip)
}

// GetRealIP возвращает IP-адрес из контекста.
func GetRealIP(ctx context.Context) string {
	return GetKey(ctx, keyRealIP)
}

// SetHash256 устанавливает в контекст hash256.
func SetHash256(ctx context.Context, hash string) context.Context {
	return SetKey(ctx, keyHash256, hash)
}

// GetHash256 возвращает hash из контекста.
func GetHash256(ctx context.Context) string {
	return GetKey(ctx, keyHash256)
}

func GetKey(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func SetKey(ctx context.Context, key, val string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	md.Set(key, val)
	return metadata.NewOutgoingContext(ctx, md)
}
