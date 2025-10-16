package file

import (
	"context"
	"fmt"
	"os"

	"github.com/htrandev/metrics/internal/model"
	"github.com/mailru/easyjson"
)

type FileRepository struct {
	file *os.File
}

func NewRepository(fileName string) (*FileRepository, error) {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		return nil, fmt.Errorf("restore: open file: %w", err)
	}
	return &FileRepository{
		file: f,
	}, nil
}

func (f *FileRepository) Flush(ctx context.Context, metrics []model.Metric) error {
	for _, metric := range metrics {
		data, err := easyjson.Marshal(metric)
		if err != nil {
			return fmt.Errorf("marshal metric [%+v]: %w", metric, err)
		}
		if _, err := f.file.Write(data); err != nil {
			return fmt.Errorf("write data: %w", err)
		}

		if _, err := f.file.Write([]byte("\n")); err != nil {
			return fmt.Errorf("write new line: %w", err)
		}
	}
	return nil
}

func (f *FileRepository) Close() error {
	return f.file.Close()
}
