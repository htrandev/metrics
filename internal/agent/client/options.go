package client

import (
	"crypto/rsa"

	"go.uber.org/zap"
)

type Option func(*CommonOptions)

type CommonOptions struct {
	maxRetry  int
	addr      string
	ip        string
	signature string
	key       *rsa.PublicKey
	logger    *zap.Logger
}

func WithMaxRetry(retry int) Option {
	return func(opt *CommonOptions) {
		opt.maxRetry = retry
	}
}

func WithAddr(addr string) Option {
	return func(opt *CommonOptions) {
		opt.addr = addr
	}
}

func WithPublicKey(key *rsa.PublicKey) Option {
	return func(opt *CommonOptions) {
		opt.key = key
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(opt *CommonOptions) {
		opt.logger = logger
	}
}

func WithIP(ip string) Option {
	return func(opt *CommonOptions) {
		opt.ip = ip
	}
}

func WithSignature(signature string) Option {
	return func(opt *CommonOptions) {
		opt.signature = signature
	}
}
