package main

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/bbalet/stopwords"
	"github.com/dgrijalva/jwt-go"
	"github.com/euskadi31/go-tokenizer"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/navossoc/bayesian"
	"golang.org/x/crypto/bcrypt"
)

type Token struct {
	Token string `json:"token"`
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
}

type Product struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Image         []byte    `json:"image"`
	Price         float64   `json:"price"`
	Description   string    `json:"description"`
	Comments      []Comment `json:"comments,omitempty"`
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
var WrongPasswordError error = errors.New("Wrong password!")
var CommentNotBeEmptyString error = errors.New("Comment not be empty string!")

const SecretKey = "14465375-b4a8-47fa-9692-c986d4a825ee"

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

	product.Comments, err = service.repository.GetCommentsByIDList(product.CommentIDList)

	if err != nil {
		return nil, err
	}

	return product, nil
}

func (service *Service) AddProduct(productDTO ProductDTO) (*Product, error) {

	product := &Product{
		ID:          GenerateUUID(8),
		Name:        productDTO.Name,
		Image:       productDTO.Image,
		Description: productDTO.Description,
		Price:       productDTO.Price,
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

func (service *Service) AddProductComment(productID string, commentDTO CommentDTO) (*Product, error) {

	if commentDTO.Text == "" {
		return nil, CommentNotBeEmptyString
	}

	comment, err := service.AddComment(commentDTO.Text)

	if err != nil {
		return nil, err
	}

	product, err := service.GetProduct(productID)

	if err != nil {
		return nil, err
	}

	product.CommentIDList = append(product.CommentIDList, comment.ID)

	product, err = service.repository.UpdateProduct(*product)

	if err != nil {
		return nil, err
	}

	product.Comments, err = service.repository.GetCommentsByIDList(product.CommentIDList)

	if err != nil {
		return nil, err
	}

	return product, nil
}

func (service *Service) Login(userCredentials UserCredentialsDTO) (*Token, *fiber.Cookie, error) {

	user, err := service.repository.GetUserByUsername(userCredentials.Username)

	if err != nil {
		return nil, nil, err
	}

	if user == nil {
		return nil, nil, UserNotFoundError
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userCredentials.Password)); err != nil {
		return nil, nil, WrongPasswordError
	}

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
	})

	token, err := claims.SignedString([]byte(SecretKey))

	if err != nil {
		return nil, nil, err
	}

	cookie := fiber.Cookie{
		Name:    "user-token",
		Value:   token,
		Expires: time.Now().Add(time.Hour * 24),
	}

	return &Token{
		Token: token,
	}, &cookie, nil
}

func (service *Service) AnalyzeText(text string) (Comment, error) {

	classifier := bayesian.NewClassifier(positive, negative) //classlar?? belirleme

	classifier.Learn(service.positiveReviewWords, positive) //classlar?? atama
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
			PN:            variable,
			PositiveRatio: ratios[0],
			NegativeRatio: ratios[1],
		},
	}, nil
}

func preProcessSentence(sentence string) (sentenceWords []string) { //sentenceleri kelimelere ay??rarak tokenle??tirme
	re := regexp.MustCompile("[^a-zA-Z 0-9]+") //harf olmayanlar?? sil
	t := tokenizer.New()
	newSentence := re.ReplaceAllString(strings.ToLower(sentence), "")  //harf olmayanlar?? sil
	cleadnedSentence := stopwords.CleanString(newSentence, "en", true) //stopword sil
	tokenizedSentence := t.Tokenize(cleadnedSentence)
	for _, word := range tokenizedSentence {
		sentenceWords = append(sentenceWords, word)
	}
	return sentenceWords
}

func preProcessReviews(reviews []string) (reviewWords []string) { //reviewleri kelimelere ay??rma
	re := regexp.MustCompile("[^a-zA-Z 0-9]+") //harf olmayanlar?? sil
	t := tokenizer.New()
	for _, sentence := range reviews {
		newSentence := re.ReplaceAllString(strings.ToLower(sentence), "")  //harf olmayanlar?? sil
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
