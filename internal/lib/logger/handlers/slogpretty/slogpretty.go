package slogpretty

import (
	"context"
	"encoding/json"
	"io"

	stdLog "log"

	"log/slog"

	"github.com/fatih/color"
)

/*
Как slog вызывает этот код. Например:

logger.Info(
	"HTTP request",
	"method", "GET",
	"url", "/api/v1/teams",
	"status", 200,
)

log/slog примерно превращает это в:

slog.Logger
     ↓
slog.Record
     ↓
PrettyHandler.Handle()
     ↓
r.Level
r.Message
r.Attrs(...)
     ↓
форматированный вывод

И получится примерно:

[09:45:12.123] INFO: HTTP request {
  "method": "GET",
  "status": 200,
  "url": "/api/v1/teams"
}
*/

// PrettyHandlerOptions содержит настройки стандартного slog.Handler.
//
// Через SlogOpts можно, например, задать минимальный уровень логирования:
//
//	&slog.HandlerOptions{
//	    Level: slog.LevelDebug,
//	}
type PrettyHandlerOptions struct {
	SlogOpts *slog.HandlerOptions
}

// PrettyHandler — пользовательский обработчик slog, который выводит сообщения в удобном для чтения формате.
//
// Handler — стандартный slog.Handler. Используется для сохранения совместимости с интерфейсом slog.Handler.
//
// l — стандартный log.Logger, через который выполняется вывод.
//
// attrs — дополнительные атрибуты, добавленные через WithAttrs().
type PrettyHandler struct {
	opts PrettyHandlerOptions
	slog.Handler
	l     *stdLog.Logger
	attrs []slog.Attr
}

// NewPrettyHandler создаёт новый PrettyHandler.
//
// out определяет, куда будет выводиться лог.
// Например, это может быть os.Stdout.
//
// Внутри также создаётся стандартный JSONHandler.
// PrettyHandler использует собственный Handle() для форматирования вывода.
func (opts PrettyHandlerOptions) NewPrettyHandler(
	out io.Writer,
) *PrettyHandler {
	h := &PrettyHandler{
		Handler: slog.NewJSONHandler(out, opts.SlogOpts),
		// Создаём обычный Logger без стандартного префикса даты и времени.
		l: stdLog.New(out, "", 0),
	}

	return h
}

// Handle обрабатывает одну запись slog.Record.
//
// Форматирует:
//   - время;
//   - уровень логирования;
//   - сообщение;
//   - дополнительные атрибуты.
//
// Уровни и сообщение раскрашиваются для удобного чтения логов в консоли.
func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	// Например:
	//
	// INFO:
	// DEBUG:
	// WARN:
	// ERROR:
	level := r.Level.String() + ":"

	// Выбираем цвет в зависимости от уровня логирования.
	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	// Создаём map для дополнительных полей записи.
	//
	// NumAttrs() используется для предварительного выделения необходимого размера map.
	fields := make(map[string]interface{}, r.NumAttrs())

	// Получаем атрибуты непосредственно из slog.Record.
	//
	// Например:
	//
	//	log.Info(
	//	    "request completed",
	//	    "method", "GET",
	//	    "status", 200,
	//	)
	//
	// даст:
	//
	//	method = GET
	//	status = 200
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	// Добавляем постоянные атрибуты, ранее созданные через WithAttrs().
	for _, a := range h.attrs {
		fields[a.Key] = a.Value.Any()
	}

	var b []byte
	var err error

	// Если дополнительные поля существуют, преобразуем их в форматированный JSON.
	if len(fields) > 0 {
		b, err = json.MarshalIndent(fields, "", "  ")
		if err != nil {
			return err
		}
	}

	// Формат времени:
	//
	//	[09:45:12.123]
	timeStr := r.Time.Format("[15:05:05.000]")

	// Сам текст сообщения выводим голубым цвето
	msg := color.CyanString(r.Message)

	// Формируем окончательную строку лога.
	h.l.Println(
		timeStr,
		level,
		msg,
		color.WhiteString(string(b)),
	)

	return nil
}

// WithAttrs создаёт новый Handler с дополнительными атрибутами.
//
// Этот метод необходим для реализации интерфейса slog.Handler.
//
// Например:
//
//	logger := logger.With("component", "storage")
//
// После этого атрибут component будет автоматически добавляться к последующим сообщениям этого logger.
func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &PrettyHandler{
		Handler: h.Handler,
		l:       h.l,
		attrs:   attrs,
	}
}

// WithGroup создаёт Handler для группы атрибутов.
//
// Метод необходим для реализации интерфейса slog.Handler.
//
// В текущей реализации полноценная поддержка групп для PrettyHandler ещё не реализована.
func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	// TODO: implement
	return &PrettyHandler{
		Handler: h.Handler.WithGroup(name),
		l:       h.l,
	}
}
