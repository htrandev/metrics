package audit

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"
)

var _ Observer = (*URLAudit)(nil)

// URLAudit реализует Observer для отправки событий по HTTP на указанный URL.
type URLAudit struct {
	id     uuid.UUID
	url    string
	client *resty.Client

	logger *zap.Logger
}

// NewURL возвращает новый экземпляр URLAudit.
func NewURL(id uuid.UUID, url string, client *resty.Client, l *zap.Logger) *URLAudit {
	if l == nil {
		l = zap.NewNop()
	}
	l.With(zap.String("scope", "urlAudit"))

	if client == nil {
		client = resty.New().
			SetTimeout(30 * time.Second)
	}

	return &URLAudit{
		id:     id,
		url:    url,
		client: client,
		logger: l,
	}
}

// GetID возвращает уникальный идентификатор URLAudit.
func (u *URLAudit) GetID() string {
	return u.id.String()
}

// Update сериализует полученную информацию и отправляет POST-апрос на указанный URL.
func (u *URLAudit) Update(ctx context.Context, info AuditInfo) {
	u.logger.Debug("send info to url")
	b, err := easyjson.Marshal(info)
	if err != nil {
		u.logger.Error("marshal audit info", zap.Error(err))
		return
	}

	r := u.client.R().
		SetBody(b).
		SetContext(ctx)

	if _, err := r.Post(u.url); err != nil {
		if errors.Is(err, io.EOF) {
			return
		}
		u.logger.Error("send request to url auditor", zap.Error(err))
	}
}
