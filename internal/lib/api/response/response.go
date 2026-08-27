package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Response — стандартная структура ответа REST API.
//
// Status содержит результат выполнения запроса: "OK" или "Error".
//
// Error содержит текст ошибки.
// Поле не включается в JSON, если строка пустая
// благодаря тегу omitempty.
type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// Возможные значения поля Status.
const (
	StatusOK    = "OK"
	StatusError = "Error"
)

// OK создаёт стандартный успешный ответ.
//
// JSON:
//
//	{
//	    "status": "OK"
//	}
//
// Поле Error отсутствует, так как оно пустое
// и объявлено с тегом json:"error,omitempty".
func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

// Error создаёт стандартный ответ с ошибкой.
//
// Например:
//
//	Error("user not found")
//
// даст JSON:
//
//	{
//	    "status": "Error",
//	    "error": "user not found"
//	}
func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

// ValidationError преобразует ошибки библиотеки validator в стандартный Response приложения.
//
// validator.ValidationErrors может содержать сразу несколько ошибок проверки полей. Для каждой ошибки формируется понятное текстовое сообщение.
//
// Например, для структуры:
//
//	type Request struct {
//	    Login string `validate:"required"`
//	    URL   string `validate:"url"`
//	}
//
// могут быть сформированы сообщения:
//
//	field Login is a required field
//	field URL is not a valid URL
//
// Если ошибок несколько, они объединяются через ", ".
func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	for _, err := range errs {

		// ActualTag возвращает правило validator,  которое не прошло проверку: required, url и т.д.
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		case "url":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid URL", err.Field()))
		default: // Для правил, которые отдельно  здесь не обрабатываются.
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}
	// Объединяем все сообщения в одну строку и возвращаем стандартный ответ с ошибкой.
	return Response{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "),
	}
}
