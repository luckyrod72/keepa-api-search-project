package main

import (
	"context"
	"fmt"
)

func firestoreFunction(ctx context.Context, requestID, asin string, productData *SimplifiedResponse) error {
	// delete product from Firestore
	if err := deleteFromFirestore(ctx, asin); err != nil {
		return fmt.Errorf("[RequestID: %s] Failed to delete data from Firestore for ASIN %s: %v", requestID, asin, err)
	}

	// Save to Firestore
	if err := saveToFirestore(ctx, asin, productData); err != nil {
		return fmt.Errorf("[RequestID: %s] Failed to save data to Firestore for ASIN %s: %v", requestID, asin, err)
	}

	return nil
}

func deleteFromFirestore(ctx context.Context, asin string) interface{} {
	// Delete product from Firestore
	docRef := firestoreClient.Collection("products").Doc(asin)
	_, err := docRef.Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete product from Firestore: %v", err)
	}
	return nil
}

func saveToFirestore(ctx context.Context, asin string, productData *SimplifiedResponse) error {
	// Create a new document in Firestore
	docRef := firestoreClient.Collection("products").Doc(asin)
	_, err := docRef.Set(ctx, productData)
	if err != nil {
		return fmt.Errorf("failed to save product to Firestore: %v", err)
	}
	return nil
}
