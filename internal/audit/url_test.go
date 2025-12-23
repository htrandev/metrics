package audit

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	easyjson "github.com/mailru/easyjson"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/htrandev/metrics/pkg/logger"
)

type URLAuditorSuite struct {
	suite.Suite

	id     uuid.UUID
	client *resty.Client
	url    string
	logger *zap.Logger

	urlAudit *URLAudit
}

func (s *URLAuditorSuite) SetupSuite() {
	var err error

	s.id = uuid.New()
	s.logger, err = logger.NewZapLogger("debug")
	s.Require().NoError(err)

	s.client = resty.New()

	s.urlAudit = NewURL(s.id, s.url, s.client, s.logger)
}

func TestURLAudit(t *testing.T) {
	suite.Run(t, new(URLAuditorSuite))
}

func (s *URLAuditorSuite) TestGetId() {
	id := s.urlAudit.GetID()
	s.Require().Equal(s.id.String(), id)
}

func (s *URLAuditorSuite) TestUpdate() {
	ctx := context.Background()

	info := AuditInfo{
		Timestamp: time.Now().Unix(),
		Metrics:   []string{"1", "2", "3"},
		IP:        "127.0.0.1",
	}

	var gotMethod string
	var gotBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		gotMethod = r.Method

		body, err := io.ReadAll(r.Body)
		s.Require().NoError(err)

		gotBody = body
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s.urlAudit.url = srv.URL
	s.urlAudit.Update(ctx, info)

	s.Require().Equal(http.MethodPost, gotMethod)

	expectedJSON, err := easyjson.Marshal(info)
	s.Require().NoError(err)
	s.Require().Equal(string(expectedJSON), string(gotBody))

}
