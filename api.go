package main

import (
	"github.com/gofiber/fiber/v2"
)

type UserCredentialsDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type UserDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ProductDTO struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

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

func (api *API) AddProductHandler(c *fiber.Ctx) error {

	productDTO := ProductDTO{}
	err := c.BodyParser(&productDTO)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
	}

	product, err := api.service.AddProduct(productDTO)

	switch err {
	case nil:
		c.JSON(product)
		c.Status(fiber.StatusCreated)
	default:
		c.Status(fiber.StatusInternalServerError)
	}
	return nil
}

func (api *API) UpdateProductHandler(c *fiber.Ctx) error {

	productID := c.Params("id")
	productDTO := ProductDTO{}
	err := c.BodyParser(&productDTO)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
	}

	product, err := api.service.UpdateProduct(productID, productDTO)

	switch err {
	case nil:
		c.JSON(product)
		c.Status(fiber.StatusOK)
	default:
		c.Status(fiber.StatusInternalServerError)
	}
	return nil
}

func (api *API) CreateUserHandler(c *fiber.Ctx) error {

	userDTO := UserDTO{}
	err := c.BodyParser(&userDTO)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
	}

	user, err := api.service.CreateUser(userDTO)

	switch err {
	case nil:
		c.JSON(user)
		c.Status(fiber.StatusCreated)
	case UserAlreadyRegisteredError:
		c.Status(fiber.StatusBadRequest)
	default:
		c.Status(fiber.StatusInternalServerError)
	}
	return nil
}

func (api *API) AddProductCommentHandler(c *fiber.Ctx) error {

	productID := c.Params("id")
	commentDTO := CommentDTO{}
	err := c.BodyParser(&commentDTO)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
	}

	product, err := api.service.AddProductComment(productID, commentDTO)

	switch err {
	case nil:
		c.JSON(product)
		c.Status(fiber.StatusCreated)
	default:
		c.Status(fiber.StatusInternalServerError)
	}
	return nil
}

func (api *API) LoginHandler(c *fiber.Ctx) error {
	userCredentials := UserCredentialsDTO{}
	err := c.BodyParser(&userCredentials)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
	}

	token, cookie, err := api.service.Login(userCredentials)

	switch err {
	case nil:
		c.JSON(token)
		c.Cookie(cookie)
		c.Status(fiber.StatusOK)
	case UserNotFoundError:
		c.Status(fiber.StatusNotFound)
	case WrongPasswordError:
		c.Status(fiber.StatusBadRequest)
	default:
		c.Status(fiber.StatusInternalServerError)
	}
	return nil
}
