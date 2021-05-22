package main

import (
	"errors"
	"regexp"
	"strings"

	"github.com/bbalet/stopwords"
	"github.com/euskadi31/go-tokenizer"
	"github.com/google/uuid"
	"github.com/navossoc/bayesian"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
}

type Product struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Price         float64   `json:"price"`
	Comments      []Comment `json:"comments"`
	CommentIDList []string  `json:"-"`
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

type ReviewData struct {
	Comment string `json:"comment"`
	Class   string `json:"class"`
}

type Service struct {
	repository          *Repository
	positiveReviewWords []string
	negativeReviewWords []string
}

const (
	positive bayesian.Class = "positive"
	negative bayesian.Class = "negative"
)

var CommentNotFoundError error = errors.New("Comment not found!")
var ProductNotFoundError error = errors.New("Product not found!")
var UserNotFoundError error = errors.New("User not found!")
var UserAlreadyRegisteredError error = errors.New("User already registered!")

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

	comment, err := service.AnalyzeText(text)

	if err != nil {
		return nil, err
	}

	comment.ID = GenerateUUID(8)

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

func (service *Service) GetProducts() ([]Product, error) {

	products, err := service.repository.GetProducts()

	if err != nil {
		return nil, err
	}

	for i, product := range products {
		comments, err := service.repository.GetCommentsByIDList(product.CommentIDList)
		if err != nil {
			return nil, err
		}
		products[i].Comments = comments
	}

	return products, nil
}

func (service *Service) GetProduct(productID string) (*Product, error) {

	product, err := service.repository.GetProduct(productID)

	if err != nil {
		return nil, err
	}

	return product, nil
}

func (service *Service) AddProduct(productDTO ProductDTO) (*Product, error) {

	product := &Product{
		ID:    GenerateUUID(8),
		Name:  productDTO.Name,
		Price: productDTO.Price,
	}

	product, err := service.repository.AddProduct(*product)

	if err != nil {
		return nil, err
	}

	return product, nil
}

func (service *Service) UpdateProduct(productID string, productDTO ProductDTO) (*Product, error) {
	existingProduct, err := service.repository.GetProduct(productID)

	if err != nil {
		return nil, ProductNotFoundError
	}

	existingProduct.Name = productDTO.Name
	existingProduct.Price = productDTO.Price

	_, err = service.repository.UpdateProduct(*existingProduct)

	if err != nil {
		return nil, err
	}

	return service.GetProduct(productID)
}

func (service *Service) CreateUser(userDTO UserDTO) (*User, error) {

	existingUser, _ := service.repository.GetUserByUsername(userDTO.Username)

	if existingUser != nil {
		return nil, UserAlreadyRegisteredError
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userDTO.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	user := User{
		ID:       GenerateUUID(8),
		Username: userDTO.Username,
		Password: string(hashedPassword),
	}

	newUser, err := service.repository.CreateUser(user)

	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (service *Service) AnalyzeText(text string) (Comment, error) {

	classifier := bayesian.NewClassifier(positive, negative) //classları belirleme

	classifier.Learn(service.positiveReviewWords, positive) //classları atama
	classifier.Learn(service.negativeReviewWords, negative)
	classifier.ConvertTermsFreqToTfIdf()

	sentenceWords := preProcessSentence(text)

	_, result, _ := classifier.ProbScores(sentenceWords)

	ratios, _, _ := classifier.ProbScores(sentenceWords)

	variable := ""
	if result == 0 {
		variable = "positive"
	}
	if result == 1 {
		variable = "negative"
	}

	return Comment{
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
