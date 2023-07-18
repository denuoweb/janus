package transformer

import (
	"github.com/labstack/echo"
	"github.com/denuoweb/janus/pkg/notifier"
)

func getNotifier(c echo.Context) *notifier.Notifier {
	storedValue := c.Get("notifier")
	notifier, ok := storedValue.(*notifier.Notifier)
	if !ok {
		return nil
	}
	return notifier
}
