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

// GetAllUsers retrieves all user documents from the SpotifyUser collection
func GetAllUsers(ctx context.Context, client *firestore.Client) ([]User, error) {
	var users []User

	// Reference the SpotifyUser collection
	iter := client.Collection("SpotifyUser").Documents(ctx)
	defer iter.Stop() // Close iterator after use

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate documents: %w", err)
		}

		// Map document data to a User struct
		var user User
		if err := doc.DataTo(&user); err != nil {
			return nil, fmt.Errorf("failed to map document data: %w", err)
		}
		user.ID = doc.Ref.ID // Optionally add Firestore document ID
		users = append(users, user)
	}

	return users, nil
}

func GetUserByID(ctx context.Context, client *firestore.Client, userID string) (*User, error) {
	// Query the collection for documents where the "id" field matches the provided userID
	query := client.Collection("SpotifyUser").Where("id", "==", userID).Limit(1)

	// Execute the query
	iter := query.Documents(ctx)
	defer iter.Stop() // Ensure iterator is properly closed

	doc, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return nil, fmt.Errorf("user with ID %s not found", userID)
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	// Map document data to a User struct
	var user User
	if err := doc.DataTo(&user); err != nil {
		return nil, fmt.Errorf("failed to map document data: %w", err)
	}
	user.ID = doc.Ref.ID // Optionally set the Firestore document ID if needed

	return &user, nil
}
