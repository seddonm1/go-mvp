package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type userkey int

var userKey userkey

func GetUserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userKey).(*User)
	return user, ok
}

func validateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

func AuthorizationInterceptor(dbpool *pgxpool.Pool) func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// The header must contain an "authorization" header
		authorization := ""
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if auth, ok := md["authorization"]; ok {
				authorization = strings.Join(auth, "")
			}
		}

		// The header must be a bearer token.
		if !strings.HasPrefix(authorization, "Bearer ey") {
			return nil, errUnauthenticated
		}

		// The Bearer token must be valid.
		tokenString := strings.TrimPrefix(authorization, "Bearer ")
		claims, err := validateToken(tokenString)
		if err != nil {
			return nil, errUnauthenticated
		}

		// The sub must be able to be retrieved.
		sub, ok := claims["sub"].(string)
		if !ok {
			return nil, errUnauthenticated
		}

		tx, err := dbpool.Begin(ctx)
		if err != nil {
			fmt.Println(err)
			return nil, errUnauthenticated
		}

		// The IdentityUser must be able to be retrieved from the database.
		identityUser, err := RetrieveIdentityUserById(ctx, tx, sub)
		if err != nil {
			return nil, errUnauthenticated
		}

		// The User must be able to be retrieved from the database.
		user, err := identityUser.RetrieveUser(ctx, tx)
		if err != nil {
			return nil, errUnauthenticated
		}

		// Attach the user to the context
		ctx = context.WithValue(ctx, userKey, user)

		// Execute the function
		return handler(ctx, req)
	}
}

func CreateToken(sub string) (string, error) {
	// Get the token instance with the Signing method
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	// Choose an expiration time. Shorter the better
	exp := time.Now().Add(time.Hour * 24)
	// Add your claims
	token.Claims = &jwt.RegisteredClaims{
		// Set the exp and sub claims. sub is usually the userID
		ExpiresAt: jwt.NewNumericDate(exp),
		Subject:   sub,
	}
	// Sign the token with your secret key
	val, err := token.SignedString(secret)

	if err != nil {
		// On error return the error
		return "", err
	}
	// On success return the token string
	return val, nil
}
