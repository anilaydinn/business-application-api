package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserEntity struct {
	ID       string `bson:"id"`
	Username string `bson:"username"`
	Password string `bson:"password"`
}

type ProductEntity struct {
	ID          string   `bson:"id"`
	Name        string   `bson:"name"`
	Image       []byte   `bson:"image"`
	Description string   `bson:"description"`
	Price       float64  `bson:"price"`
	Comments    []string `bson:"comments"`
}

type CommentEntity struct {
	ID      string  `bson:"id"`
	Text    string  `bson:"text"`
	PNModel PNModel `bson:"pnModel"`
}

type ReviewDataEntity struct {
	Comment string `bson:"comment"`
	Class   string `bson:"class"`
}

type Repository struct {
	client *mongo.Client
}

func NewRepository() *Repository {
	uri := "mongodb+srv://admin:sk45dgnFHltHOKiG@cluster0.ljcyz.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	client.Connect(ctx)

	if err != nil {
		log.Fatal(err)
	}

	return &Repository{client}
}

func NewTestRepository() *Repository {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

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

func (repository *Repository) AddProduct(product Product) (*Product, error) {
	collection := repository.client.Database("business").Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	productEntity := convertProductModelToProductEntity(product)

	_, err := collection.InsertOne(ctx, productEntity)

	if err != nil {
		return nil, err
	}

	return repository.GetProduct(productEntity.ID)
}

func (repository *Repository) GetProduct(productID string) (*Product, error) {
	collection := repository.client.Database("business").Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur := collection.FindOne(ctx, bson.M{"id": productID})

	if cur.Err() != nil {
		return nil, cur.Err()
	}

	if cur == nil {
		return nil, ProductNotFoundError
	}

	productEntity := ProductEntity{}
	err := cur.Decode(&productEntity)

	if err != nil {
		return nil, err
	}

	product := convertProductEntityToProductModel(productEntity)

	return &product, nil
}

func (repository *Repository) GetProducts() ([]Product, error) {
	collection := repository.client.Database("business").Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, bson.M{})

	if err != nil {
		return nil, err
	}

	products := []Product{}
	for cur.Next(ctx) {
		productEntity := ProductEntity{}
		err := cur.Decode(&productEntity)
		if err != nil {
			return nil, err
		}
		products = append(products, convertProductEntityToProductModel(productEntity))
	}

	return products, nil
}

func (repository *Repository) GetCommentsByIDList(commentIDList []string) ([]Comment, error) {
	collection := repository.client.Database("business").Collection("comments")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"id": bson.M{
			"$in": commentIDList,
		},
	}

	cur, err := collection.Find(ctx, filter)

	if err != nil {
		return nil, err
	}

	comments := []Comment{}
	for cur.Next(ctx) {
		commentEntity := CommentEntity{}
		err := cur.Decode(&commentEntity)
		if err != nil {
			return nil, err
		}
		comments = append(comments, convertCommentEntityToModel(commentEntity))
	}

	return comments, nil
}

func (repository *Repository) UpdateProduct(product Product) (*Product, error) {
	collection := repository.client.Database("business").Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	productEntity := convertProductModelToProductEntity(product)

	result := collection.FindOneAndReplace(ctx, bson.M{"id": product.ID}, productEntity)

	if result.Err() != nil {
		return nil, result.Err()
	}

	return repository.GetProduct(product.ID)
}

func (repository *Repository) CreateUser(user User) (*User, error) {
	collection := repository.client.Database("business").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	userEntity := convertUserModelToUserEntity(user)

	_, err := collection.InsertOne(ctx, userEntity)

	if err != nil {
		return nil, err
	}

	return repository.GetUser(userEntity.ID)
}

func (repository *Repository) GetUser(userID string) (*User, error) {
	collection := repository.client.Database("business").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	cur := collection.FindOne(ctx, bson.M{"id": userID})

	if cur.Err() != nil {
		return nil, cur.Err()
	}

	if cur == nil {
		return nil, UserNotFoundError
	}

	userEntity := UserEntity{}
	err := cur.Decode(&userEntity)

	if err != nil {
		return nil, err
	}

	user := convertUserEntityToUserModel(userEntity)

	return &user, nil
}

func (repository *Repository) GetUserByUsername(username string) (*User, error) {
	collection := repository.client.Database("business").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	cur := collection.FindOne(ctx, bson.M{"username": username})

	if cur.Err() != nil {
		return nil, cur.Err()
	}

	if cur == nil {
		return nil, UserNotFoundError
	}

	userEntity := UserEntity{}
	err := cur.Decode(&userEntity)

	if err != nil {
		return nil, err
	}

	user := convertUserEntityToUserModel(userEntity)

	return &user, nil
}

func (repository *Repository) GetReviewsData() ([]ReviewData, error) {
	collection := repository.client.Database("test").Collection("reviews")
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, bson.M{})

	if err != nil {
		return nil, err
	}

	reviews := []ReviewData{}
	for cur.Next(ctx) {
		reviewDataEntity := ReviewDataEntity{}
		err := cur.Decode(&reviewDataEntity)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, convertReviewDataEntityToReviewData(reviewDataEntity))
	}

	return reviews, nil
}

func convertCommentModelToEntity(comment *Comment) CommentEntity {
	return CommentEntity{
		ID:   comment.ID,
		Text: comment.Text,
		PNModel: PNModel{
			PN:            comment.PNModel.PN,
			PositiveRatio: comment.PNModel.PositiveRatio,
			NegativeRatio: comment.PNModel.NegativeRatio,
		},
	}
}

func convertCommentEntityToModel(commentEntity CommentEntity) Comment {
	return Comment{
		ID:   commentEntity.ID,
		Text: commentEntity.Text,
		PNModel: PNModel{
			PN:            commentEntity.PNModel.PN,
			PositiveRatio: commentEntity.PNModel.PositiveRatio,
			NegativeRatio: commentEntity.PNModel.NegativeRatio,
		},
	}
}

func convertCommentEntitiesToCommentModels(commentEntities []CommentEntity) []Comment {
	comments := []Comment{}
	for _, commentEntity := range commentEntities {
		comments = append(comments, convertCommentEntityToModel(commentEntity))
	}
	return comments
}

func convertProductModelToProductEntity(product Product) ProductEntity {

	if product.CommentIDList == nil {
		product.CommentIDList = []string{}
	}

	return ProductEntity{
		ID:          product.ID,
		Name:        product.Name,
		Image:       product.Image,
		Description: product.Description,
		Price:       product.Price,
		Comments:    product.CommentIDList,
	}
}

func convertProductEntityToProductModel(productEntity ProductEntity) Product {
	return Product{
		ID:            productEntity.ID,
		Name:          productEntity.Name,
		Image:         productEntity.Image,
		Description:   productEntity.Description,
		Price:         productEntity.Price,
		CommentIDList: productEntity.Comments,
	}
}

func convertReviewDataEntityToReviewData(reviewDataEntity ReviewDataEntity) ReviewData {
	return ReviewData{
		Comment: reviewDataEntity.Comment,
		Class:   reviewDataEntity.Class,
	}
}

func convertUserModelToUserEntity(user User) UserEntity {
	return UserEntity{
		ID:       user.ID,
		Username: user.Username,
		Password: user.Password,
	}
}

func convertUserEntityToUserModel(userEntity UserEntity) User {
	return User{
		ID:       userEntity.ID,
		Username: userEntity.Username,
		Password: userEntity.Password,
	}
}
