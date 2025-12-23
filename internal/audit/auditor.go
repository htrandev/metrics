package audit

import "context"

type Observer interface {
	Update(ctx context.Context, info AuditInfo)
	GetID() string
}

type Auditor struct {
	observers map[string]Observer
	info      AuditInfo
}

//easyjson:json
type AuditInfo struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IP        string   `json:"ip_address"`
}

func NewAuditor() *Auditor {
	return &Auditor{observers: make(map[string]Observer)}
}

func (a *Auditor) Register(o Observer) {
	if a.observers == nil {
		a.observers = make(map[string]Observer)
	}
	a.observers[o.GetID()] = o
}

func (a *Auditor) Deregister(o Observer) {
	delete(a.observers, o.GetID())
}

func (a *Auditor) Update(ctx context.Context, info AuditInfo) {
	a.info = info
	a.notifyAll(ctx)
}

func (a *Auditor) notifyAll(ctx context.Context) {
	for _, o := range a.observers {
		o.Update(ctx, a.info)
	}
}
