package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CommentEntity struct {
	ID   string `bson:"id"`
	Text string `bson:"text"`
}

type Repository struct {
	client *mongo.Client
}

func NewRepository() *Repository {
	uri := "mongodb://localhost:27017"
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	client.Connect(ctx)

	if err != nil {
		log.Fatal(err)
	}

	return &Repository{client}
}

func (repository *Repository) AddComment(comment Comment) (*Comment, error) {
	collection := repository.client.Database("business").Collection("comments")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	commentEntity := convertCommentModelToEntity(&comment)
	_, err := collection.InsertOne(ctx, commentEntity)

	if err != nil {
		return nil, err
	}

	return repository.GetComment(commentEntity.ID)
}

func (repository *Repository) GetComment(commentID string) (*Comment, error) {
	collection := repository.client.Database("business").Collection("comments")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur := collection.FindOne(ctx, bson.M{"id": commentID})

	if cur.Err() != nil {
		return nil, cur.Err()
	}

	if cur == nil {
		return nil, CommentNotFoundError
	}

	commentEntity := CommentEntity{}
	err := cur.Decode(&commentEntity)

	if err != nil {
		return nil, err
	}

	comment := convertCommentEntityToModel(commentEntity)

	return &comment, nil
}

func (repository *Repository) GetComments() ([]Comment, error) {
	collection := repository.client.Database("business").Collection("comments")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}

	cur, err := collection.Find(ctx, filter)

	if err != nil {
		return nil, err
	}

	commentEntities := []CommentEntity{}
	for cur.Next(ctx) {
		var commentEntity CommentEntity
		err := cur.Decode(&commentEntity)
		if err != nil {
			log.Fatal(err)
		}
		commentEntities = append(commentEntities, commentEntity)
	}

	return convertCommentEntitiesToCommentModels(commentEntities), nil
}

func (repository *Repository) DeleteComment(commentID string) error {
	collection := repository.client.Database("business").Collection("comments")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"id": commentID}
	_, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	return nil
}

func (repository *Repository) UpdateComment(comment *Comment) (*Comment, error) {
	collection := repository.client.Database("business").Collection("comments")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	commentEntity := convertCommentModelToEntity(comment)

	filters := bson.M{"id": comment.ID}

	result := collection.FindOneAndReplace(ctx, filters, commentEntity)

	if result.Err() != nil {
		return nil, result.Err()
	}

	return repository.GetComment(commentEntity.ID)
}

func convertCommentModelToEntity(comment *Comment) CommentEntity {
	return CommentEntity{
		ID:   comment.ID,
		Text: comment.Text,
	}
}

func convertCommentEntityToModel(commentEntity CommentEntity) Comment {
	return Comment{
		ID:   commentEntity.ID,
		Text: commentEntity.Text,
	}
}

func convertCommentEntitiesToCommentModels(commentEntities []CommentEntity) []Comment {
	comments := []Comment{}
	for _, commentEntity := range commentEntities {
		comments = append(comments, convertCommentEntityToModel(commentEntity))
	}
	return comments
}
