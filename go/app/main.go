package main

import (
	"crypto/sha256"
	"database/sql"
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
	_ "github.com/mattn/go-sqlite3"
)

const (
	ImgDir = "images"
)

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

	file, err := c.FormFile("image")
	if err != nil {
		res := Response{Message: "Error getting file from form"}
		return c.JSON(http.StatusBadRequest, res)
	}

	fileHash := hashImage(file)

	db := connectDB(c)
	if err != nil {
		res := Response{Message: "Error initializing database"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	defer db.Close()

	// Get or insert the category
	var categoryID int64
	err = db.QueryRow("SELECT id FROM categories WHERE name = ?", category).Scan(&categoryID)
	if err == sql.ErrNoRows {
		// Category does not exist, insert it
		result, err := db.Exec("INSERT INTO categories (name) VALUES (?)", category)
		if err != nil {
			res := Response{Message: "Error saving category to database"}
			return c.JSON(http.StatusInternalServerError, res)
		}
		categoryID, _ = result.LastInsertId()
	} else if err != nil {
		res := Response{Message: "Error retrieving category from database"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	// Insert the item
	_, err = db.Exec("INSERT INTO items (name, category_id, image_name) VALUES (?, ?, ?)",
		name, categoryID, fileHash+".jpg")
	if err != nil {
		res := Response{Message: "Error saving item to database"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	message := fmt.Sprintf("item received: %s, Category:%s", name, category)
	return c.JSON(http.StatusOK, Response{Message: message})
}

func getItems(c echo.Context) error {
	db := connectDB(c)
	defer db.Close()
	query := `
	SELECT items.id, items.name, categories.name AS category, items.image_name
	FROM items
	JOIN categories ON items.category_id = categories.id
`

	c.Logger().Infof("Executing query: %s", query)

	rows, err := db.Query(query)

	if err != nil {
		c.Logger().Errorf("Error executing query: %v", err)
		res := Response{Message: fmt.Sprintf("Error executing query: %v", err)}
		return c.JSON(http.StatusInternalServerError, res)
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id int
		var name, category, imageName string
		if err := rows.Scan(&id, &name, &category, &imageName); err != nil {
			c.Logger().Errorf("Error scanning rows: %v", err)
			res := Response{Message: fmt.Sprintf("Error scanning rows: %v", err)}
			return c.JSON(http.StatusInternalServerError, res)
		}
		newItem := map[string]interface{}{
			"id":         id,
			"name":       name,
			"category":   category,
			"image_name": imageName,
		}
		items = append(items, newItem)
	}

	return c.JSON(http.StatusOK, items)
}

func connectDB(c echo.Context) *sql.DB {
	// dbOpen
	db, err := sql.Open("sqlite3", "/app/db/mercari.sqlite3")
	if err != nil {
		c.Logger().Errorf("No connection to the database: %v", err)
	}
	return db
}

func getItemDetails(c echo.Context) error {
	itemID := c.Param("item_id")

	db := connectDB(c)
	defer db.Close()

	row := db.QueryRow("SELECT * FROM items WHERE id = ?", itemID)

	var id int
	var name, category, imageName string
	err := row.Scan(&id, &name, &category, &imageName)
	if err != nil {
		res := Response{Message: "Error retrieving item details from database"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	itemDetails := map[string]interface{}{
		"id":         id,
		"name":       name,
		"category":   category,
		"image_name": imageName,
	}

	return c.JSON(http.StatusOK, itemDetails)
}

func searchItems(c echo.Context) error {
	keyword := c.QueryParam("keyword")

	db := connectDB(c)
	defer db.Close()

	query := `
    SELECT items.name, categories.name AS category
    FROM items
    JOIN categories ON items.category_id = categories.id
    WHERE items.name LIKE '%' || ? || '%' COLLATE NOCASE`

	rows, err := db.Query(query, "%"+keyword+"%")
	if err != nil {
		log.Errorf("Error executing query: %v", err)
		return c.JSON(http.StatusInternalServerError, Response{Message: "Error searching items in the database"})
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var itemName, categoryName string
		if err := rows.Scan(&itemName, &categoryName); err != nil {
			log.Errorf("Error scanning rows: %v", err)
			return c.JSON(http.StatusInternalServerError, Response{Message: "Error scanning rows"})
		}
		newItem := map[string]interface{}{
			"name":     itemName,
			"category": categoryName,
		}
		items = append(items, newItem)
	}

	response := map[string]interface{}{"items": items}
	return c.JSON(http.StatusOK, response)
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

	res := Response{Message: "Image received"}
	return c.JSON(http.StatusOK, res)
}

func hashImage(file *multipart.FileHeader) string {
	f, err := file.Open()
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

	frontURL := os.Getenv("FRONT_URL")
	if frontURL == "" {
		frontURL = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{frontURL},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/", root)
	e.POST("/items", addItem)
	e.GET("/items", getItems)
	e.GET("/items/:item_id", getItemDetails)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/search", searchItems)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
