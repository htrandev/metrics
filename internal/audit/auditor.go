package audit

import "context"

// Observer определяет интерфейс наблюдателя 
// для получения уведомлений о событиях.
type Observer interface {
	Update(ctx context.Context, info AuditInfo)
	GetID() string
}

// Auditor реализует паттерн "Наблюдатель" для распределения событий
// между зарегистрированными наблюдателями.
type Auditor struct {
	observers map[string]Observer
	info      AuditInfo
}

// AuditInfo содержит информацию о событии.
// 
//easyjson:json
type AuditInfo struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IP        string   `json:"ip_address"`
}

// NewAuditor создает и возвращает новый экземпляр Auditor.
func NewAuditor() *Auditor {
	return &Auditor{observers: make(map[string]Observer)}
}

// Register регистрирует нового наблюдателя в Auditor по его уникальному ID.
func (a *Auditor) Register(o Observer) {
	if a.observers == nil {
		a.observers = make(map[string]Observer)
	}
	a.observers[o.GetID()] = o
}

// Deregister удаляет наблюдателя из Auditor по его уникальному ID.
func (a *Auditor) Deregister(o Observer) {
	delete(a.observers, o.GetID())
}

// Update обновляет информацию о событии и уведомляет всех зарегистрированных наблюдателей.
func (a *Auditor) Update(ctx context.Context, info AuditInfo) {
	a.info = info
	a.notifyAll(ctx)
}

// notifyAll уведомляет всех зарегистрированных наблюдателей об обновлении.
func (a *Auditor) notifyAll(ctx context.Context) {
	for _, o := range a.observers {
		o.Update(ctx, a.info)
	}
}
