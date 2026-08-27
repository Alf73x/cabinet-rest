package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

/*
Login/password
      ↓
пользователь найден в БД
      ↓
CreateToken()
      ↓
Claims
      ↓
JWT + secret
      ↓
HS256
      ↓
eyJ...xxxxx.yyyyy
      ↓
отправляем клиенту


Следующий HTTP request:

Authorization: Bearer eyJ...
                ↓
           ParseToken()
                ↓
       проверка подписи HS256
                ↓
       проверка exp и claims
                ↓
        получаем UserID
                ↓
     запрос считается авторизованным
*/

type TokenService struct {
	secret []byte        // секретный ключ, которым подписывается токен.
	ttl    time.Duration //  время жизни токена.
}

// Claims содержит данные (claims), которые сохраняются внутри JWT.
//
// Помимо собственных данных UserID и LoginName,
// в JWT включаются стандартные RegisteredClaims:
// Subject, IssuedAt, ExpiresAt и другие стандартные поля JWT.
type Claims struct {
	UserID    int64  `json:"user_id"`
	LoginName string `json:"login_name"`

	jwt.RegisteredClaims
}

// NewTokenService создаёт сервис для работы с JWT.
//
// secret используется для подписи и последующей проверки токенов.
// ttl определяет, сколько времени созданный токен будет действителе
func NewTokenService(secret string, ttl time.Duration) *TokenService {
	return &TokenService{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

// CreateToken создаёт новый JWT для пользователя.
//
// В токен записываются:
//   - UserID и LoginName — данные приложения;
//   - Subject — идентификатор пользователя в стандартном поле JWT;
//   - IssuedAt — время создания токена;
//   - ExpiresAt — время окончания действия токена.
//
// Токен подписывается алгоритмом HMAC-SHA256 (HS256)
// с использованием секретного ключа TokenService.
//
// Возвращает строку JWT, время окончания его действия и ошибку.
func (s *TokenService) CreateToken(userID int64, loginName string) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.ttl)

	// Формируем содержимое (payload) JWT.
	claims := Claims{
		UserID:    userID,
		LoginName: loginName,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	// Создаём объект JWT с алгоритмом подписи HS256 и подготовленными claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем JWT секретным ключом.
	// Результатом является готовая строка:
	// header.payload.signature
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign JWT: %w", err)
	}

	return tokenString, expiresAt, nil
}

// ParseToken разбирает и проверяет JWT.
//
// Проверяется:
//   - корректность структуры токена;
//   - подпись с использованием secret;
//   - допустимый алгоритм подписи (только HS256);
//   - стандартные claims, включая срок действия ExpiresAt.
//
// При успешной проверке возвращаются данные Claims.
func (s *TokenService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		// Указываем тип структуры, в которую библиотека должна разобрать payload JWT.
		&Claims{},
		// Функция возвращает ключ, которым библиотека проверит подпись токена.
		func(token *jwt.Token) (any, error) {
			return s.secret, nil
		},
		// Разрешаем только HS256. Это не позволяет принять токен, использующий неожиданный алгоритм подписи.
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("parse JWT: %w", err)
	}

	// ParseWithClaims возвращает Claims через интерфейс. Проверяем, что внутри действительно находится *Claims.
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid JWT claims")
	}

	// Дополнительная проверка общего состояния токена.
	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	return claims, nil
}
