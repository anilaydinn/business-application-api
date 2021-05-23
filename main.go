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

	app.Post("api/analyze", api.AnalyzeTextHandler)

	app.Post("/api/comments", api.AddCommentHandler)
	app.Get("/api/comments", api.GetCommentsHandler)
	app.Get("/api/comments/:id", api.GetCommentHandler)
	app.Delete("/api/comments/:id", api.DeleteCommentHandler)
	app.Patch("/api/comments/:id", api.UpdateCommentHandler)

	app.Post("/api/products", api.AddProductHandler)
	app.Get("/api/products", api.GetProductsHandler)
	app.Get("/api/products/:id", api.GetProductHandler)
	app.Patch("/api/products/:id", api.UpdateProductHandler)
	app.Patch("/api/products/:id/comments", api.AddProductCommentHandler)

	app.Post("/api/users/register", api.CreateUserHandler)
	app.Post("/api/users/login", api.LoginHandler)
	app.Get("/api/user", api.UserHandler)
	app.Post("/api/logout", api.LogoutHandler)

	return app
}
