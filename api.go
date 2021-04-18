package main

import (
	"github.com/gofiber/fiber/v2"
)

type CommentDTO struct {
	Text string `json:"text"`
}

type API struct {
	service *Service
}

func NewAPI(service *Service) API {
	return API{
		service: service,
	}
}

func (api *API) GetCommentsHandler(c *fiber.Ctx) error {

	comments, err := api.service.GetComments()

	switch err {
	case nil:
		c.JSON(comments)
		c.Status(fiber.StatusOK)
	default:
		c.Status(fiber.StatusInternalServerError)
	}
	return nil
}

func (api *API) GetCommentHandler(c *fiber.Ctx) error {

	commentID := c.Params("id")

	comment, err := api.service.GetComment(commentID)

	switch err {
	case nil:
		c.JSON(comment)
		c.Status(fiber.StatusOK)
	case CommentNotFoundError:
		c.Status(fiber.StatusNotFound)
	default:
		c.Status(fiber.StatusInternalServerError)
	}

	return nil
}

func (api *API) AddCommentHandler(c *fiber.Ctx) error {

	commentDTO := CommentDTO{}
	err := c.BodyParser(&commentDTO)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
	}

	comment, err := api.service.AddComment(commentDTO.Text)

	switch err {
	case nil:
		c.JSON(comment)
		c.Status(fiber.StatusCreated)
	}
	return nil
}

func (api *API) DeleteCommentHandler(c *fiber.Ctx) error {

	commentID := c.Params("id")

	err := api.service.DeleteComment(commentID)

	switch err {
	case nil:
		c.Status(fiber.StatusNoContent)
	default:
		c.Status(fiber.StatusInternalServerError)
	}

	return nil
}

func (api *API) UpdateCommentHandler(c *fiber.Ctx) error {
	commentID := c.Params("id")
	commentDTO := CommentDTO{}

	err := c.BodyParser(&commentDTO)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
	}

	comment, err := api.service.UpdateComment(commentID, commentDTO.Text)

	switch err {
	case nil:
		c.JSON(comment)
		c.Status(fiber.StatusOK)
	default:
		c.Status(fiber.StatusInternalServerError)
	}
	return nil
}

func (api *API) AnalyzeTextHandler(c *fiber.Ctx) error {

	commentDTO := CommentDTO{}
	err := c.BodyParser(&commentDTO)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
	}

	comment, err := api.service.AnalyzeText(commentDTO.Text)

	switch err {
	case nil:
		c.JSON(comment)
		c.Status(fiber.StatusOK)
	default:
		c.Status(fiber.StatusInternalServerError)
	}
	return nil

}

func (api *API) GetProductsHandler(c *fiber.Ctx) error {

	products, err := api.service.GetProducts()

	switch err {
	case nil:
		c.JSON(products)
		c.Status(fiber.StatusOK)
	default:
		c.Status(fiber.StatusInternalServerError)
	}
	return nil
}

func (api *API) GetProductHandler(c *fiber.Ctx) error {

	productID := c.Params("id")

	product, err := api.service.GetProduct(productID)

	switch err {
	case nil:
		c.JSON(product)
		c.Status(fiber.StatusOK)
	case ProductNotFoundError:
		c.Status(fiber.StatusBadRequest)
	default:
		c.Status(fiber.StatusInternalServerError)
	}
	return nil

}
