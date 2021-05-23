package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetComments(t *testing.T) {

	Convey("Given comments", t, func() {
		repository := GetCleanTestRepository()
		service := NewService(repository, nil, nil)
		api := NewAPI(&service)

		comment1 := Comment{
			ID:   "1",
			Text: "Bu bir yorumdur",
		}

		comment2 := Comment{
			ID:   "2",
			Text: "Test yorum",
		}
		repository.AddComment(comment1)
		repository.AddComment(comment2)

		Convey("When the get comments request sent", func() {
			app := SetupApp(&api)
			req, _ := http.NewRequest(http.MethodGet, "/api/comments", nil)

			resp, err := app.Test(req)
			So(err, ShouldBeNil)

			Convey("Then status code should be 200", func() {
				So(resp.StatusCode, ShouldEqual, fiber.StatusOK)
			})

			Convey("Then all comments should return", func() {
				actualResult := []Comment{}
				actualResponseBody, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(actualResponseBody, &actualResult)
				So(err, ShouldBeNil)

				So(actualResult, ShouldHaveLength, 2)
				So(actualResult[0].ID, ShouldEqual, "1")
				So(actualResult[1].ID, ShouldEqual, "2")
			})
		})
	})
}

func TestGetComment(t *testing.T) {

	Convey("Given comments", t, func() {
		repository := GetCleanTestRepository()
		service := NewService(repository, nil, nil)
		api := NewAPI(&service)

		comment1 := Comment{
			ID:   "1",
			Text: "Bu bir yorumdur",
		}

		comment2 := Comment{
			ID:   "2",
			Text: "Test yorum",
		}
		repository.AddComment(comment1)
		repository.AddComment(comment2)

		Convey("When the get comment request sent with comment id", func() {
			app := SetupApp(&api)
			req, _ := http.NewRequest(http.MethodGet, "/api/comments/"+comment2.ID, nil)

			resp, err := app.Test(req)
			So(err, ShouldBeNil)

			Convey("Then status code should be 200", func() {
				So(resp.StatusCode, ShouldEqual, fiber.StatusOK)
			})

			Convey("Then called comment should return", func() {
				actualResult := Comment{}
				actualResponseBody, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(actualResponseBody, &actualResult)
				So(err, ShouldBeNil)

				So(actualResult.ID, ShouldEqual, "2")
			})
		})

		Convey("When the get comment request sent with non existing comment id", func() {
			app := SetupApp(&api)
			req, _ := http.NewRequest(http.MethodGet, "/api/comments/654dasasd56d", nil)

			resp, err := app.Test(req)
			So(err, ShouldBeNil)

			Convey("Then status code should be 404", func() {
				So(resp.StatusCode, ShouldEqual, fiber.StatusNotFound)
			})
		})
	})
}

func TestAddComment(t *testing.T) {
	Convey("Given valid comment details", t, func() {
		repository := GetCleanTestRepository()

		csvfile, err := os.Open("data/IMDBDataset.csv") //dosyayı al

		if err != nil {
			fmt.Println("csv açılamadi")
		}

		defer csvfile.Close() //program sonunda dosyayı kapa

		csvLines, err := csv.NewReader(csvfile).ReadAll() //dosyayı oku

		if err != nil {
			fmt.Println(err)
		}

		reviews := []ReviewData{}

		for _, line := range csvLines {
			reviews = append(reviews, ReviewData{
				Comment: line[0],
				Class:   line[1], //'1'. sutun
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

		Convey("When add user request sent", func() {

			commentDTO := CommentDTO{
				Text: "I don't like this movie. I hate it. I don't watch again.",
			}

			reqBody, err := json.Marshal(commentDTO)
			So(err, ShouldBeNil)

			req, _ := http.NewRequest(http.MethodPost, "/api/comments", bytes.NewReader(reqBody))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Set("Content-Length", strconv.Itoa(len(reqBody)))

			app := SetupApp(&api)

			resp, err := app.Test(req, 30000)
			So(err, ShouldBeNil)

			Convey("Then status code should be 201", func() {
				So(resp.StatusCode, ShouldEqual, fiber.StatusCreated)
			})
			Convey("Then added user should be returned as response", func() {
				actualResponseBody, _ := ioutil.ReadAll(resp.Body)
				actualResult := Comment{}
				err := json.Unmarshal(actualResponseBody, &actualResult)

				So(err, ShouldBeNil)
				So(actualResult.ID, ShouldNotBeNil)
				So(actualResult.Text, ShouldEqual, commentDTO.Text)
				So(actualResult.PNModel.PN, ShouldEqual, "negative")
				So(actualResult.PNModel.NegativeRatio, ShouldNotBeNil)
				So(actualResult.PNModel.PositiveRatio, ShouldNotBeNil)
			})
		})
	})
}

func TestDeleteComment(t *testing.T) {
	Convey("Given comment delete request", t, func() {
		repository := GetCleanTestRepository()

		service := NewService(repository, nil, nil)
		api := NewAPI(&service)

		comment := Comment{
			ID:   GenerateUUID(8),
			Text: "Comment text",
		}
		repository.AddComment(comment)

		Convey("When comment delete request sent", func() {
			app := SetupApp(&api)
			req, _ := http.NewRequest(http.MethodDelete, "/api/comments/"+comment.ID, nil)
			resp, err := app.Test(req, 30000)
			So(err, ShouldBeNil)

			Convey("Then status code should be 204", func() {
				So(resp.StatusCode, ShouldEqual, fiber.StatusNoContent)
			})

			Convey("Then comment should be delete", func() {
				user, err := repository.GetComment(comment.ID)
				So(err, ShouldNotBeNil)
				So(user, ShouldBeNil)
			})
		})
	})
}

func TestUpdateComment(t *testing.T) {
	Convey("Given comment details", t, func() {
		repository := GetCleanTestRepository()

		service := NewService(repository, nil, nil)
		api := NewAPI(&service)

		commentID := GenerateUUID(8)

		comment := Comment{
			ID:   commentID,
			Text: "Test Comment",
		}
		repository.AddComment(comment)

		Convey("When request sent with new comment details", func() {
			app := SetupApp(&api)
			commentDTO := CommentDTO{
				Text: "Update Test Comment",
			}
			reqBody, err := json.Marshal(commentDTO)
			So(err, ShouldBeNil)

			req, err := http.NewRequest(http.MethodPatch, "/api/comments/"+comment.ID, bytes.NewReader(reqBody))
			req.Header.Add("Content-type", "application/json")
			req.Header.Set("Content-Length", strconv.Itoa(len(reqBody)))
			So(err, ShouldBeNil)

			resp, err := app.Test(req, 30000)
			So(err, ShouldBeNil)

			Convey("Then status code should be 200", func() {
				So(resp.StatusCode, ShouldEqual, fiber.StatusOK)
			})

			Convey("Then user should be updated", func() {
				actualResult := Comment{}
				actualResponseBody, _ := ioutil.ReadAll(resp.Body)
				err = json.Unmarshal(actualResponseBody, &actualResult)
				So(err, ShouldBeNil)

				So(actualResult.ID, ShouldEqual, comment.ID)
				So(actualResult.Text, ShouldEqual, commentDTO.Text)
			})
		})
	})
}

func TestAnalyzeText(t *testing.T) {
	Convey("Given a text", t, func() {
		repository := GetCleanTestRepository()

		csvfile, err := os.Open("data/IMDBDataset.csv") //dosyayı al

		if err != nil {
			fmt.Println("csv açılamadi")
		}

		defer csvfile.Close() //program sonunda dosyayı kapa

		csvLines, err := csv.NewReader(csvfile).ReadAll() //dosyayı oku

		if err != nil {
			fmt.Println(err)
		}

		reviews := []ReviewData{}

		for _, line := range csvLines {
			reviews = append(reviews, ReviewData{
				Comment: line[0],
				Class:   line[1], //'1'. sutun
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

		Convey("When request sent with comment details", func() {
			app := SetupApp(&api)

			commentDTO := CommentDTO{
				Text: "This is very bad!",
			}

			reqBody, err := json.Marshal(commentDTO)
			So(err, ShouldBeNil)

			req, err := http.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader(reqBody))
			req.Header.Add("Content-type", "application/json")
			req.Header.Set("Content-Length", strconv.Itoa(len(reqBody)))
			So(err, ShouldBeNil)

			resp, err := app.Test(req, 30000)
			So(err, ShouldBeNil)

			Convey("Then status code should be 200", func() {
				So(resp.StatusCode, ShouldEqual, 200)
			})

			Convey("Then text analyzed comment response should return", func() {
				actualResult := Comment{}
				actualResponseBody, _ := ioutil.ReadAll(resp.Body)
				err = json.Unmarshal(actualResponseBody, &actualResult)
				So(err, ShouldBeNil)

				So(actualResult.Text, ShouldEqual, "This is very bad!")
				So(actualResult.PNModel, ShouldNotBeNil)
				So(actualResult.PNModel.PN, ShouldEqual, "negative")
			})
		})
	})
}

func TestGetProducts(t *testing.T) {

	Convey("Given products", t, func() {
		repository := GetCleanTestRepository()
		service := NewService(repository, nil, nil)
		api := NewAPI(&service)

		app := SetupApp(&api)

		comment1 := Comment{
			ID:   "123",
			Text: "Test comment1",
		}

		comment2 := Comment{
			ID:   "456",
			Text: "Test comment2",
		}

		comment3 := Comment{
			ID:   "789",
			Text: "Test comment3",
		}

		product1 := Product{
			ID:    GenerateUUID(8),
			Name:  "product1",
			Price: 50.0,
			CommentIDList: []string{
				"123",
				"456",
			},
		}

		product2 := Product{
			ID:    GenerateUUID(8),
			Name:  "product2",
			Price: 100.0,
			CommentIDList: []string{
				"789",
			},
		}

		repository.AddComment(comment1)
		repository.AddComment(comment2)
		repository.AddComment(comment3)
		repository.AddProduct(product1)
		repository.AddProduct(product2)

		Convey("When get products request sent", func() {
			req, _ := http.NewRequest(http.MethodGet, "/api/products", nil)

			resp, err := app.Test(req, 30000)
			So(err, ShouldBeNil)

			Convey("Then status code should be 200", func() {
				So(resp.StatusCode, ShouldEqual, fiber.StatusOK)
			})

			Convey("Then all products should return", func() {
				actualResult := []Product{}
				actualResponseBody, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(actualResponseBody, &actualResult)
				So(err, ShouldBeNil)

				So(actualResult, ShouldHaveLength, 2)
				So(actualResult[0].ID, ShouldEqual, product1.ID)
				So(actualResult[0].Name, ShouldEqual, product1.Name)
				So(actualResult[0].Price, ShouldEqual, product1.Price)
				So(actualResult[0].Comments, ShouldHaveLength, 2)
				So(actualResult[0].Comments[0].ID, ShouldEqual, comment1.ID)
				So(actualResult[0].Comments[0].Text, ShouldEqual, comment1.Text)
				So(actualResult[0].Comments[1].ID, ShouldEqual, comment2.ID)
				So(actualResult[0].Comments[1].Text, ShouldEqual, comment2.Text)
				So(actualResult[1].ID, ShouldEqual, product2.ID)
				So(actualResult[1].Name, ShouldEqual, product2.Name)
				So(actualResult[1].Price, ShouldEqual, product2.Price)
				So(actualResult[1].Comments, ShouldHaveLength, 1)
				So(actualResult[1].Comments[0].ID, ShouldEqual, comment3.ID)
				So(actualResult[1].Comments[0].Text, ShouldEqual, comment3.Text)
			})
		})
	})
}

func TestGetProduct(t *testing.T) {
	Convey("Given a product", t, func() {
		repository := GetCleanTestRepository()
		service := NewService(repository, nil, nil)
		api := NewAPI(&service)

		app := SetupApp(&api)

		product1 := Product{
			ID:    GenerateUUID(8),
			Name:  "Test Product1",
			Price: 10.0,
		}

		product2 := Product{
			ID:    GenerateUUID(8),
			Name:  "Test Product2",
			Price: 15.0,
		}

		repository.AddProduct(product1)
		repository.AddProduct(product2)

		Convey("When the get product request sent with id params", func() {
			req, _ := http.NewRequest(http.MethodGet, "/api/products/"+product2.ID, nil)

			resp, err := app.Test(req, 30000)
			So(err, ShouldBeNil)

			Convey("Then status code should be 200", func() {
				So(resp.StatusCode, ShouldEqual, fiber.StatusOK)
			})

			Convey("Then requested product should return", func() {
				actualResult := Product{}
				actualResponseBody, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(actualResponseBody, &actualResult)
				So(err, ShouldBeNil)

				So(actualResult.ID, ShouldEqual, product2.ID)
				So(actualResult.Name, ShouldEqual, product2.Name)
				So(actualResult.Price, ShouldEqual, product2.Price)
			})
		})
	})
}

func TestAddProduct(t *testing.T) {
	Convey("Given a valid product", t, func() {
		repository := GetCleanTestRepository()
		service := NewService(repository, nil, nil)
		api := NewAPI(&service)

		app := SetupApp(&api)

		Convey("When the new product request sent with product data", func() {

			productDTO := ProductDTO{
				Name:        "New Product Name",
				Image:       getSampleImage(),
				Description: "This is product description",
				Price:       11.0,
			}

			reqBody, err := json.Marshal(productDTO)
			So(err, ShouldBeNil)

			req, _ := http.NewRequest(http.MethodPost, "/api/products", bytes.NewReader(reqBody))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Set("Content-Length", strconv.Itoa(len(reqBody)))

			resp, err := app.Test(req, 30000)
			So(err, ShouldBeNil)

			Convey("Then status code should be 201", func() {
				So(resp.StatusCode, ShouldEqual, fiber.StatusCreated)
			})

			Convey("Then created product should return", func() {
				actualResult := Product{}
				actualResponseBody, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(actualResponseBody, &actualResult)
				So(err, ShouldBeNil)

				So(actualResult.ID, ShouldNotBeNil)
				So(actualResult.Name, ShouldEqual, productDTO.Name)
				So(len(actualResult.Image), ShouldBeGreaterThan, 0)
				So(actualResult.Description, ShouldEqual, productDTO.Description)
				So(actualResult.Price, ShouldEqual, productDTO.Price)
			})
		})
	})
}

func TestUpdateProduct(t *testing.T) {
	Convey("Given a product", t, func() {
		repository := GetCleanTestRepository()
		service := NewService(repository, nil, nil)
		api := NewAPI(&service)

		app := SetupApp(&api)

		product := Product{
			ID:    GenerateUUID(8),
			Name:  "Test Product Name",
			Price: 99.90,
		}
		repository.AddProduct(product)

		Convey("When the update product request sent with new product data", func() {

			productDTO := ProductDTO{
				Name:  "Test New Product Name",
				Price: 50.00,
			}

			reqBody, err := json.Marshal(productDTO)
			So(err, ShouldBeNil)

			req, _ := http.NewRequest(http.MethodPatch, "/api/products/"+product.ID, bytes.NewReader(reqBody))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Set("Content-Length", strconv.Itoa(len(reqBody)))

			resp, err := app.Test(req, 300000)
			So(err, ShouldBeNil)

			Convey("Then status code should be 200", func() {
				So(resp.StatusCode, ShouldEqual, fiber.StatusOK)
			})

			Convey("Then updated product should return", func() {
				actualResult := Product{}
				actualResponseBody, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(actualResponseBody, &actualResult)
				So(err, ShouldBeNil)

				So(actualResult.ID, ShouldEqual, product.ID)
				So(actualResult.Name, ShouldEqual, productDTO.Name)
				So(actualResult.Price, ShouldEqual, productDTO.Price)
			})
		})
	})
}

func TestCreateUser(t *testing.T) {
	Convey("Given a valid user details", t, func() {
		repository := GetCleanTestRepository()
		service := NewService(repository, nil, nil)
		api := NewAPI(&service)

		app := SetupApp(&api)

		user := User{
			ID:       GenerateUUID(8),
			Username: "test-user",
			Password: "test-pw",
		}
		repository.CreateUser(user)

		Convey("When the new user request sent with user data", func() {

			userDTO := UserDTO{
				Username: "thracian",
				Password: "123123",
			}

			reqBody, err := json.Marshal(userDTO)
			So(err, ShouldBeNil)

			req, _ := http.NewRequest(http.MethodPost, "/api/users/register", bytes.NewReader(reqBody))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Set("Content-Length", strconv.Itoa(len(reqBody)))

			resp, err := app.Test(req, 300000)
			So(err, ShouldBeNil)

			Convey("Then status code should be 201", func() {
				So(resp.StatusCode, ShouldEqual, fiber.StatusCreated)
			})

			Convey("Then created user should return", func() {
				actualResult := User{}
				actualResponseBody, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(actualResponseBody, &actualResult)
				So(err, ShouldBeNil)

				So(actualResult.ID, ShouldNotBeNil)
				So(actualResult.Username, ShouldEqual, userDTO.Username)
			})
		})

		Convey("When already registered username user data request sent", func() {
			userDTO := UserDTO{
				Username: "test-user",
				Password: "asdasd",
			}

			reqBody, err := json.Marshal(userDTO)
			So(err, ShouldBeNil)

			req, _ := http.NewRequest(http.MethodPost, "/api/users/register", bytes.NewReader(reqBody))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Set("Content-Length", strconv.Itoa(len(reqBody)))

			resp, err := app.Test(req, 300000)
			So(err, ShouldBeNil)

			Convey("Then status code should be 400", func() {
				So(resp.StatusCode, ShouldEqual, fiber.StatusBadRequest)
			})
		})

	})
}

func TestAddCommentProduct(t *testing.T) {
	Convey("Given a product with doesn't have comments", t, func() {
		repository := GetCleanTestRepository()

		csvfile, err := os.Open("data/IMDBDataset.csv") //dosyayı al

		if err != nil {
			fmt.Println("csv açılamadi")
		}

		defer csvfile.Close() //program sonunda dosyayı kapa

		csvLines, err := csv.NewReader(csvfile).ReadAll() //dosyayı oku

		if err != nil {
			fmt.Println(err)
		}

		reviews := []ReviewData{}

		for _, line := range csvLines {
			reviews = append(reviews, ReviewData{
				Comment: line[0],
				Class:   line[1], //'1'. sutun
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

		product := Product{
			ID:          GenerateUUID(8),
			Name:        "Test Product",
			Price:       55.55,
			Description: "Test Description",
		}
		repository.AddProduct(product)

		Convey("When comment data sent with product id", func() {

			commentDTO := CommentDTO{
				Text: "Comment",
			}

			reqBody, err := json.Marshal(commentDTO)
			So(err, ShouldBeNil)

			req, _ := http.NewRequest(http.MethodPatch, "/api/products/"+product.ID+"/comments", bytes.NewReader(reqBody))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Set("Content-Length", strconv.Itoa(len(reqBody)))

			resp, err := app.Test(req, 300000)
			So(err, ShouldBeNil)

			Convey("Then status code should be 201", func() {
				So(resp.StatusCode, ShouldEqual, fiber.StatusCreated)
			})

			Convey("Then product with new comment should return", func() {
				actualResult := Product{}
				actualResponseBody, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(actualResponseBody, &actualResult)

				So(err, ShouldBeNil)

				So(actualResult.ID, ShouldEqual, product.ID)
				So(actualResult.Name, ShouldEqual, product.Name)
				So(actualResult.Description, ShouldEqual, product.Description)
				So(actualResult.Price, ShouldEqual, product.Price)
				So(actualResult.Comments, ShouldHaveLength, 1)
				So(actualResult.Comments[0].Text, ShouldEqual, "Comment")
			})
		})
	})

}

func TestUserLogin(t *testing.T) {
	Convey("Given a registered user", t, func() {
		repository := GetCleanTestRepository()
		service := NewService(repository, nil, nil)
		api := NewAPI(&service)

		app := SetupApp(&api)

		user := User{
			ID:       GenerateUUID(8),
			Username: "test",
			Password: "$2a$10$xknCQfdvbnX9H7Tk7vvsQeiOKB0OnnySLUw/CyzICymcX0iHgRi3q",
		}
		repository.CreateUser(user)

		Convey("When user login request sent with user credentials", func() {

			userCredentialsDTO := UserCredentialsDTO{
				Username: "test",
				Password: "123",
			}

			reqBody, err := json.Marshal(userCredentialsDTO)
			So(err, ShouldBeNil)

			req, _ := http.NewRequest(http.MethodPost, "/api/users/login", bytes.NewReader(reqBody))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Set("Content-Length", strconv.Itoa(len(reqBody)))

			resp, err := app.Test(req, 300000)
			So(err, ShouldBeNil)

			Convey("Then status code should be 200", func() {
				So(resp.StatusCode, ShouldEqual, fiber.StatusOK)
			})

			Convey("Then user token should returned", func() {
				actualResult := Token{}
				actualResponseBody, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(actualResponseBody, &actualResult)
				So(err, ShouldBeNil)

				So(actualResult.Token, ShouldNotBeNil)
				So(len(actualResult.Token), ShouldBeGreaterThan, 0)
			})
		})
	})
}

func GetCleanTestRepository() *Repository {

	repository := NewTestRepository()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	businessDB := repository.client.Database("business")
	businessDB.Drop(ctx)

	return repository
}
