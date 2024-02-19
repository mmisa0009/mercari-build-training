package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "images"
)

var nextID int=1

type Response struct {
	Message string `json:"message"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	c.Logger().Infof("Receive item: %s, Category: %s", name, category)

	items, err:= loadItems()
	if err!=nil {
		res:= Response{Message: "Error loading items"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	id := nextID
	nextID++

	newItem:= map[string]interface{}{"id": id, "name": name, "category": category}
	items["items"] = append(items["items"].([]map[string]interface{}), newItem)

	err = saveItems(items)
	if err != nil {
		res:= Response{Message: "Error saving item"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	message := fmt.Sprintf("item received: %s, Category:%s, ID: %d", name, category, id)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func loadItems() (map[string]interface{}, error) {
	file, err:= os.ReadFile("items.json")
	if err != nil {
		return nil, err
	}

	var items map[string]interface{}
	if err:= json.Unmarshal(file, &items); err != nil{
		return nil, err
	}

	return items, nil
}

func saveItems(items map[string]interface{}) error {
	data, err:= json.MarshalIndent(items, "","  ")
	if err != nil {
		return err
	}

	err = os.WriteFile("items.json", data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func getItems(c echo.Context) error {
	items, err:= loadItems()
	if err!= nil {
		res := Response{Message: "Error loading items"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	return c.JSON(http.StatusOK, items)
}

func getItemDetails(c echo.Context) error {
	itemID := c.Param("item_id")

	items, err:= loadItems()
	if err != nil {
		res:= Response{Message: "Error loading items"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	for _, item := range items["items"].([]map[string]interface{}) {
		if id, ok := item["id"].(string); ok && id == itemID {
			return c.JSON(http.StatusOK, item)
		}
	}

	res := Response{Message: "Item not found"}
	return c.JSON(http.StatusNotFound, res)
}
				
func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func hashImage(file *multipart.FileHeader) string {
	f, err:= file.Open()
	if err != nil {
		log.Error(err)
		return ""
	}
	defer f.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		log.Error(err)
		return ""
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	front_url := os.Getenv("FRONT_URL")
	if front_url == "" {
		front_url = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{front_url},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.POST("/items", addItem)
	e.GET("/items", getItems)
	e.GET("/items/:item_id", getItemDetails)
	e.GET("/image/:imageFilename", getImg)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
