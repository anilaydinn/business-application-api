package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
)

func main() {

	csvfile, err := os.Open("data/IMDBDataset.csv") //dosyayı al

	if err != nil {
		fmt.Println("csv açılamadi")
	}

	defer csvfile.Close() //program sonunda dosyayı kapa

	csvLines, err := csv.NewReader(csvfile).ReadAll() //dosyayı oku

	if err != nil {
		fmt.Println(err)
	}

	reviews := []reviewData{}

	for _, line := range csvLines {
		reviews = append(reviews, reviewData{
			comment: line[0],
			class:   line[1], //'1'. sutun
		})
	}

	positiveReview := []string{}
	negativeReview := []string{}

	for _, item := range reviews { //sadece reviewleri alma ve ayırma
		if item.class == "positive" {
			positiveReview = append(positiveReview, item.comment)
		}
		if item.class == "negative" {
			negativeReview = append(negativeReview, item.comment)
		}
	}
	positiveReviewWords := preProcessReviews(positiveReview)

	negativeReviewWords := preProcessReviews(negativeReview)

	repository := NewRepository()
	service := NewService(repository, positiveReviewWords, negativeReviewWords)
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
