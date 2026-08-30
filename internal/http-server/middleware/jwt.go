package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	jwtservice "CabinetREST/internal/jwt"
)

type contextKey string // Создаётся собственный тип для ключей context.

type UserStatusProvider interface {
	IsUserEnabled(ctx context.Context, userID int64) (bool, error)
}

const claimsContextKey contextKey = "jwtClaims" // Это ключ, по которому JWT-данные будут храниться внутри context. Значение собственного типа уменьшает риск случайного совпадения ключей из разных пакетов.

/*
parseClaims выполняет общую работу по получению claims из HTTP-запроса:

1. Читает заголовок Authorization.
2. Проверяет наличие префикса Bearer.
3. Извлекает строку JWT-токена.
4. Проверяет и разбирает токен через tokenService.ParseToken().
5. Возвращает claims или ошибку.

Функция parseClaims ничего не записывает в HTTP Response.

Она только возвращает результат:

	claims, err := parseClaims(r, tokenService)

Это позволяет использовать её в разных middleware:

JWT:

	при ошибке возвращает 401 Unauthorized.

OptionalJWT:

	при ошибке сможет продолжить запрос как для анонимного пользователя.
*/
func parseClaims(
	r *http.Request,
	tokenService *jwtservice.TokenService,
) (*jwtservice.Claims, error) {
	authHeader := r.Header.Get("Authorization") // Чтение заголовка Authorization. Например: Authorization: Bearer abc.def.xyz

	if authHeader == "" {
		return nil, errors.New("missing authorization header")
	}

	const bearer = "Bearer "

	if !strings.HasPrefix(authHeader, bearer) { // Проверяем, что заголовок начинается с "Bearer "
		return nil, errors.New("invalid authorization header")
	}

	tokenString := strings.TrimSpace(
		strings.TrimPrefix(authHeader, bearer),
	) // Удаляем префикс "Bearer " и пробелы, получая сам JWT-токен.

	if tokenString == "" {
		return nil, errors.New("empty token")
	}

	claims, err := tokenService.ParseToken(tokenString) // Проверяем подпись токена, срок действия и получаем claims.
	if err != nil {
		return nil, err
	}

	return claims, nil
}

/*
Предположим, React отправил запрос:

GET /api/v1/auth/me
Authorization: Bearer eyJhbGciOi...

Последовательность выполнения будет такой:

React

	│
	▼

JWT middleware

	│
	├── Вызвать parseClaims()
	├── Получить заголовок Authorization
	├── Проверить, что он начинается с "Bearer "
	├── Извлечь токен
	├── ParseToken()
	├── Получить claims
	├── Сохранить claims в Context
	│
	▼

auth.NewMe

	│
	├── ClaimsFromContext()
	├── Получить UserID
	├── Загрузить пользователя из БД
	│
	▼

JSON Response
*/
func JWT(tokenService *jwtservice.TokenService, userStatusProvider UserStatusProvider) func(http.Handler) http.Handler { // Принимает tokenService, который умеет проверять и разбирать токен. Возвращает middleware стандартного вида.
	return func(next http.Handler) http.Handler { // next — это handler, который должен выполниться после успешной проверки JWT.
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // Именно эта функция будет выполняться при запросе.
			claims, err := parseClaims(r, tokenService) // Общая функция читает Authorization, извлекает токен, проверяет его и возвращает claims.

			/*
				Условно токен содержит:

				{
					"user_id": 15,
					"login": "poweruser",
					"exp": 1785000000
				}

				После успешного вызова parseClaims() мы получаем Go-структуру:

					claims.UserID
					claims.Login
			*/

			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			enabled, err := userStatusProvider.IsUserEnabled(r.Context(), claims.UserID)
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "internal server error")
				return
			}
			if !enabled {
				writeJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx := context.WithValue(
				r.Context(),
				claimsContextKey,
				claims,
			) // Сохранение claims в context. В новый context добавляется: ключ — claimsContextKey, значение — claims.

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			) // Передача запроса дальше с новым context.

			/*
				Возникает вопрос: почему просто не передать claims параметром?

				Потому что сигнатура HTTP handler фиксирована:

					func(w http.ResponseWriter, r *http.Request)

				Мы не можем изменить её на:

					func(
						w http.ResponseWriter,
						r *http.Request,
						claims *Claims,
					)

				Поэтому Go предоставляет механизм Context.

				Context позволяет передавать дополнительные данные через всю цепочку
				вызовов без изменения сигнатур функций.
			*/
		})
	}
}

/*
OptionalJWT проверяет JWT-токен, только если он передан в запросе.

Возможны два основных варианта:

1. Токена нет:

	GET /api/v1/seasons

Middleware не возвращает ошибку и передаёт запрос дальше:

	next.ServeHTTP(w, r)

В context при этом нет claims.

2. Токен есть и корректный:

	GET /api/v1/seasons
	Authorization: Bearer eyJhbGciOi...

Middleware разбирает токен, сохраняет claims в context
и передаёт запрос следующему handler.

Таким образом один endpoint может работать:

  - для анонимного пользователя;
  - для авторизованного пользователя.
*/
func OptionalJWT(tokenService *jwtservice.TokenService) func(http.Handler) http.Handler { // Принимает tokenService и возвращает необязательный JWT middleware.
	return func(next http.Handler) http.Handler { // next — следующий middleware или конечный handler.
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // Эта функция будет выполняться при каждом HTTP-запросе.
			authHeader := r.Header.Get("Authorization") // Проверяем, передал ли клиент заголовок Authorization.

			if authHeader == "" {
				// Токен не передан.
				// Это допустимо для публичного маршрута.
				// Передаём запрос дальше без claims в context.
				next.ServeHTTP(w, r)
				return
			}

			claims, err := parseClaims(r, tokenService) // Если Authorization присутствует, пытаемся проверить токен и получить claims.
			if err != nil {
				/*
					Заголовок Authorization был передан,
					но токен оказался некорректным.

					На данном этапе считаем пользователя анонимным
					и продолжаем выполнение запроса без claims.

					Например:

						Authorization: Bearer invalid-token

					Handler будет вызван, но:

						claims, ok := ClaimsFromContext(r.Context())

					вернёт:

						claims == nil
						ok == false
				*/
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				claimsContextKey,
				claims,
			) // Токен корректный. Сохраняем claims в новый context.

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			) // Передаём запрос дальше с claims внутри context.
		})
	}
}

/*
Эта функция читает из context значение по ключу claimsContextKey.

Использование в handler:

	claims, ok := appmiddleware.ClaimsFromContext(r.Context())

Если middleware JWT успешно выполнился:

	ok == true
	claims != nil

Если claims в context нет:

	ok == false
	claims == nil
*/
func ClaimsFromContext(ctx context.Context) (*jwtservice.Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*jwtservice.Claims)
	return claims, ok
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
