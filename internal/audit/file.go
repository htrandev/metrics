package audit

import (
	"context"
	"os"

	"github.com/google/uuid"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"
)

var _ Observer = (*FileAudit)(nil)

type FileAudit struct {
	id   uuid.UUID
	file *os.File

	logger *zap.Logger
}

func NewFile(id uuid.UUID, f *os.File, l *zap.Logger) *FileAudit {
	if l == nil {
		l = zap.NewNop()
	}
	l.With(zap.String("scope", "fileAudit"))

	return &FileAudit{
		id:     id,
		file:   f,
		logger: l,
	}
}

func (f *FileAudit) GetID() string {
	return f.id.String()
}

func (f *FileAudit) Update(ctx context.Context, info AuditInfo) {
	f.logger.Debug("write info to file")

	b, err := easyjson.Marshal(info)
	if err != nil {
		f.logger.Error("marshal audit info", zap.Error(err))
		return
	}

	if _, err := f.file.Write(b); err != nil {
		f.logger.Error("write data", zap.Error(err))
		return
	}

	if _, err := f.file.WriteString("\n"); err != nil {
		f.logger.Error("write new line", zap.Error(err))
		return
	}

}

func (f *FileAudit) Close() error {
	return f.file.Close()
}
