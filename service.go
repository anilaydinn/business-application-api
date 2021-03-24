package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/bbalet/stopwords"
	"github.com/euskadi31/go-tokenizer"
	"github.com/google/uuid"
	"github.com/navossoc/bayesian"
)

type Service struct {
	repository          *Repository
	positiveReviewWords []string
	negativeReviewWords []string
}

type Comment struct {
	ID      string  `json:"id"`
	Text    string  `json:"text"`
	PNModel PNModel `json:"pnModel"`
}

type PNModel struct {
	PN            string  `json:"pn"`
	PositiveRatio float64 `json:"positiveRatio"`
	NegativeRatio float64 `json:"negativeRatio"`
}

type reviewData struct {
	comment string
	class   string
}

const (
	positive bayesian.Class = "positive"
	negative bayesian.Class = "negative"
)

var CommentNotFoundError error = errors.New("Comment not found!")

func NewService(repository *Repository, positiveReviewWords []string, negativeReviewWords []string) Service {
	return Service{
		repository:          repository,
		positiveReviewWords: positiveReviewWords,
		negativeReviewWords: negativeReviewWords,
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

func (service *Service) AnalyzeText(text string) (*Comment, error) {

	classifier := bayesian.NewClassifier(positive, negative) //classları belirleme

	classifier.Learn(service.positiveReviewWords, positive) //classları atama
	classifier.Learn(service.negativeReviewWords, negative)
	classifier.ConvertTermsFreqToTfIdf()

	sentenceWords := preProcessSentence(text)

	fmt.Println(classifier.ProbScores(sentenceWords))

	_, result, _ := classifier.ProbScores(sentenceWords)

	fmt.Println(classifier.ProbScores(sentenceWords))

	ratios, _, _ := classifier.ProbScores(sentenceWords)

	variable := ""
	if result == 0 {
		variable = "positive"
	}
	if result == 1 {
		variable = "negative"
	}

	return &Comment{
		Text: text,
		PNModel: PNModel{
			PN: variable, PositiveRatio: ratios[0], NegativeRatio: ratios[1]},
	}, nil
}

func preProcessSentence(sentence string) (sentenceWords []string) { //sentenceleri kelimelere ayırarak tokenleştirme
	re := regexp.MustCompile("[^a-zA-Z 0-9]+") //harf olmayanları sil
	t := tokenizer.New()
	newSentence := re.ReplaceAllString(strings.ToLower(sentence), "")  //harf olmayanları sil
	cleadnedSentence := stopwords.CleanString(newSentence, "en", true) //stopword sil
	tokenizedSentence := t.Tokenize(cleadnedSentence)
	for _, word := range tokenizedSentence {
		sentenceWords = append(sentenceWords, word)
	}
	return sentenceWords
}

func preProcessReviews(reviews []string) (reviewWords []string) { //reviewleri kelimelere ayırma
	re := regexp.MustCompile("[^a-zA-Z 0-9]+") //harf olmayanları sil
	t := tokenizer.New()
	for _, sentence := range reviews {
		newSentence := re.ReplaceAllString(strings.ToLower(sentence), "")  //harf olmayanları sil
		cleadnedSentence := stopwords.CleanString(newSentence, "en", true) //stopword sil
		tokenizedSentence := t.Tokenize(cleadnedSentence)
		for _, word := range tokenizedSentence {
			reviewWords = append(reviewWords, word)
		}
	}
	return reviewWords
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
