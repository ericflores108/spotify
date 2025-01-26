package db

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// User represents a user document in Firestore
type User struct {
	ID           string `firestore:"id"`
	DisplayName  string `firestore:"display_name"`
	AccessToken  string `firestore:"access_token"`
	RefreshToken string `firestore:"refresh_token"`
}

func GetAllUsers(ctx context.Context, client *firestore.Client) ([]User, error) {
	var users []User

	iter := client.Collection("SpotifyUser").Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate documents: %w", err)
		}

		var user User
		if err := doc.DataTo(&user); err != nil {
			return nil, fmt.Errorf("failed to map document data: %w", err)
		}
		user.ID = doc.Ref.ID
		users = append(users, user)
	}

	return users, nil
}

func GetUserByID(ctx context.Context, client *firestore.Client, userID string) (*User, error) {
	query := client.Collection("SpotifyUser").Where("id", "==", userID).Limit(1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return nil, fmt.Errorf("user with ID %s not found", userID)
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	var user User
	if err := doc.DataTo(&user); err != nil {
		return nil, fmt.Errorf("failed to map document data: %w", err)
	}
	user.ID = doc.Ref.ID

	return &user, nil
}

func CreateUser(ctx context.Context, client *firestore.Client, user User) (string, error) {
	// Check if the user already exists
	query := client.Collection("SpotifyUser").Where("id", "==", user.ID).Limit(1)
	iter := query.Documents(ctx)
	defer iter.Stop()

	docSnap, err := iter.Next()
	if err == nil {
		// If the user exists, update the document
		_, err := docSnap.Ref.Set(ctx, user)
		if err != nil {
			return "", fmt.Errorf("failed to update user with ID %s: %w", user.ID, err)
		}
		return docSnap.Ref.ID, nil
	} else if err != iterator.Done {
		// If there's an error other than no documents, return the error
		return "", fmt.Errorf("failed to query user: %w", err)
	}

	// If the user does not exist, create a new document
	docRef, _, err := client.Collection("SpotifyUser").Add(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return docRef.ID, nil
}
