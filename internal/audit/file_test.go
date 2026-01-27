package audit

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/pkg/logger"
)

type FileAuditorSuite struct {
	suite.Suite

	file   *os.File
	id     uuid.UUID
	logger *zap.Logger

	fileAudit *FileAudit
}

func (s *FileAuditorSuite) SetupSuite() {
	f, err := os.CreateTemp("", "audit-*")
	s.Require().NoError(err)

	s.file = f
	s.id = uuid.New()
	s.logger, err = logger.NewZapLogger("debug")
	s.Require().NoError(err)

	s.fileAudit = NewFile(s.id, s.file, s.logger)
}

func (s *FileAuditorSuite) TearDownSuite() {
	s.fileAudit.Close()
	os.Remove(s.file.Name())
}

func TestFileAudit(t *testing.T) {
	suite.Run(t, new(FileAuditorSuite))
}

func (s *FileAuditorSuite) TestGetId() {
	id := s.fileAudit.GetID()
	s.Require().Equal(s.id.String(), id)
}

func (s *FileAuditorSuite) TestUpdate() {
	info := AuditInfo{
		Timestamp: time.Now().Unix(),
		Metrics:   []string{"1", "2", "3"},
		IP:        "127.0.0.1",
	}

	b, err := easyjson.Marshal(info)
	s.Require().NoError(err)

	s.logger.Info(string(b))

	s.fileAudit.Update(context.Background(), info)

	// сбрасываем буфер и возвращаемся в начало файла перед чтением
	err = s.file.Sync()
	s.Require().NoError(err)
	_, err = s.file.Seek(0, 0)
	s.Require().NoError(err)

	data, err := os.ReadFile(s.file.Name())
	s.Require().NoError(err)
	s.Require().Equal(string(b)+"\n", string(data))
}
