package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"go.uber.org/zap"
)

var (
	// gzipWriterPool содержит переиспользуемые экземпляры gzip.Writer.
	gzipWriterPool = sync.Pool{
		New: func() any {
			return gzip.NewWriter(nil)
		},
	}
	// gzipReaderPool содержит переиспользуемые экземпляры gzip.Reader.
	gzipReaderPool = sync.Pool{
		New: func() any {
			return new(gzip.Reader)
		},
	}
)

// compressWriter обертка над http.ResponseWriter для сжатия ответов в формате gzip.
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// newCompressWriter создает новый экземпляр compressWriter, получая gzip.Writer из пула.
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	zw := gzipWriterPool.Get().(*gzip.Writer)
	zw.Reset(w)
	return &compressWriter{
		w:  w,
		zw: zw,
	}
}

// Header возвращает заголовки исходного ResponseWriter.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write сжимает переданные данные через gzip.Writer и записывает в исходный ResponseWriter.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader устанавливает заголовок Content-Encoding: gzip и передает статус исходному writer.
func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.Header().Set("Content-Encoding", "gzip")
	c.w.WriteHeader(statusCode)
}

// Close завершает сжатие, возвращает gzip.Writer в пул.
func (c *compressWriter) Close() error {
	err := c.zw.Close()
	gzipWriterPool.Put(c.zw)
	return err
}

// compressReader обертка над io.ReadCloser для распаковки запросов сжатых в формате gzip.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// newCompressReader создает новый compressReader, получая gzip.Reader из пула.
func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr := gzipReaderPool.Get().(*gzip.Reader)
	if err := zr.Reset(r); err != nil {
		gzipReaderPool.Put(zr)
		return nil, err
	}
	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read распаковывает gzip-данные и возвращает исходному http.Request.Body.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close завершает распаковку, возвращает gzip.Reader в пул.
func (c *compressReader) Close() error {
	err := c.zr.Close()
	gzipReaderPool.Put(c.zr)
	c.r.Close()
	return err
}

// Compress возвращает HTTP middleware для автоматического сжатия ответов gzip.
func Compress(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w

			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportGzip := strings.Contains(acceptEncoding, "gzip")
			if supportGzip {
				cw := newCompressWriter(w)
				ow = cw
				defer cw.Close()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					logger.Error("read body",
						zap.Error(err),
						zap.String("scope", "middleware"),
						zap.String("method", "compress"),
					)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer cr.Close()
			}

			next.ServeHTTP(ow, r)
		})
	}
}
