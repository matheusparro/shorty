// internal/worker/clicks.go
package worker

import (
	"context"
	"encoding/json"
	"log"

	events "github.com/matheusparro/shorty/internal/events"
	"github.com/matheusparro/shorty/internal/queue"
	"github.com/matheusparro/shorty/internal/service"
)

var shortURLSvc *service.ShortURLService

func setShortURLService(svc *service.ShortURLService) {
	shortURLSvc = svc
}

func init() {
	register(func(c *queue.Consumer) {
		c.RegisterHandler(queue.TopicURLClicks, handleClickEvent)
	})
}

func handleClickEvent(ctx context.Context, msg []byte) error {
	if shortURLSvc == nil {
		return nil // ou retorna erro, mas em dev vamos ser tolerantes
	}

	var ev events.ClickEvent
	if err := json.Unmarshal(msg, &ev); err != nil {
		return err
	}

	// ✅ incrementa no banco
	if err := shortURLSvc.IncrementVisitByShortCode(ctx, ev.ShortCode); err != nil {
		return err
	}

	log.Printf("📥 CLICK consumed short=%s (visit++)", ev.ShortCode)
	return nil
}
