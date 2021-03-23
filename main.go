package main

import "github.com/gofiber/fiber/v2"

func main() {

	repository := NewRepository()
	service := NewService(repository)
	api := NewAPI(&service)

	app := SetupApp(&api)

	app.Listen(":3001")
}

func SetupApp(api *API) *fiber.App {
	app := fiber.New()

	app.Post("/analyze", api.AnalyzeTextHandler)
	app.Post("/comments", api.AddCommentHandler)
	app.Get("/comments", api.GetCommentsHandler)
	app.Get("/comments/:id", api.GetCommentHandler)
	app.Delete("/comments/:id", api.DeleteCommentHandler)
	app.Patch("/comments/:id", api.UpdateCommentHandler)

	return app
}
