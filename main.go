package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	repository := NewRepository()

	reviews, err := repository.GetReviewsData()

	if err != nil {
		return
	}

	for _, line := range reviews {
		reviews = append(reviews, ReviewData{
			Comment: line.Comment,
			Class:   line.Class, //'1'. sutun
		})
	}

	positiveReview := []string{}
	negativeReview := []string{}

	for _, item := range reviews { //sadece reviewleri alma ve ayırma
		if item.Class == "positive" {
			positiveReview = append(positiveReview, item.Comment)
		}
		if item.Class == "negative" {
			negativeReview = append(negativeReview, item.Comment)
		}
	}
	positiveReviewWords := preProcessReviews(positiveReview)

	negativeReviewWords := preProcessReviews(negativeReview)

	service := NewService(repository, positiveReviewWords, negativeReviewWords)
	api := NewAPI(&service)

	app := SetupApp(&api)

	app.Listen(":3001")
}

func SetupApp(api *API) *fiber.App {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	app.Post("/analyze", api.AnalyzeTextHandler)

	app.Post("/comments", api.AddCommentHandler)
	app.Get("/comments", api.GetCommentsHandler)
	app.Get("/comments/:id", api.GetCommentHandler)
	app.Delete("/comments/:id", api.DeleteCommentHandler)
	app.Patch("/comments/:id", api.UpdateCommentHandler)

	app.Post("/products", api.AddProductHandler)
	app.Get("/products", api.GetProductsHandler)
	app.Get("/products/:id", api.GetProductHandler)
	app.Patch("/products/:id", api.UpdateProductHandler)
	app.Patch("/products/:id/comments", api.AddProductCommentHandler)

	app.Post("/users", api.CreateUserHandler)
	app.Post("/users/login", api.LoginHandler)

	return app
}
