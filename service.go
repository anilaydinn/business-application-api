package main

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repository *Repository
}

type Comment struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type PNModel struct {
	Negative float32 `json:"negative"`
	Positive float32 `json:"positive"`
}

var CommentNotFoundError error = errors.New("Comment not found!")

func NewService(repository *Repository) Service {
	return Service{
		repository: repository,
	}
}

func (service *Service) GetComments() ([]Comment, error) {
	comments, err := service.repository.GetComments()

	if err != nil {
		return nil, err
	}

	return comments, nil
}

func (service *Service) GetComment(commentID string) (*Comment, error) {
	comment, err := service.repository.GetComment(commentID)

	if err != nil {
		return nil, CommentNotFoundError
	}

	return comment, nil
}

func (service *Service) AddComment(text string) (*Comment, error) {

	comment := Comment{
		ID:   GenerateUUID(8),
		Text: text,
	}

	commentResponse, err := service.repository.AddComment(comment)

	if err != nil {
		return nil, err
	}

	return commentResponse, nil
}

func (service *Service) DeleteComment(commentID string) error {
	err := service.repository.DeleteComment(commentID)

	if err != nil {
		return err
	}
	return nil
}

func (service *Service) UpdateComment(commentID, text string) (*Comment, error) {

	existingComment, err := service.GetComment(commentID)

	if err != nil {
		return nil, CommentNotFoundError
	}

	existingComment = &Comment{
		ID:   existingComment.ID,
		Text: text,
	}

	_, err = service.repository.UpdateComment(existingComment)

	if err != nil {
		return nil, err
	}

	return service.GetComment(commentID)

}

func (service *Service) AnalyzeText(text string) (PNModel, error) {

	return PNModel{
		Negative: 50.02,
		Positive: 60.45,
	}, nil
}

func GenerateUUID(length int) string {
	uuid := uuid.New().String()

	uuid = strings.ReplaceAll(uuid, "-", "")

	if length < 1 {
		return uuid
	}
	if length > len(uuid) {
		length = len(uuid)
	}

	return uuid[0:length]
}
