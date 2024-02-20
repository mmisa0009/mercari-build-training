package main

import (
	"crypto/sha256"
	"database/sql"
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
	_ "github.com/mattn/go-sqlite3"
)

const (
	ImgDir = "images"
)

type ImageDetails struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Response struct {
	Message string `json:"message"`
	ImageDetails ImageDetails `json:"imageDetails"`
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

    file, err:= c.FormFile("image")
    if err != nil {
	    res:= Response{Message: "Error getting file from form"}
	    return c.JSON(http.StatusBadRequest, res)
    }

    fileHash := hashImage(file)

    db, err := initDB()
    if err != nil {
        res := Response{Message: "Error initializing database"}
        return c.JSON(http.StatusInternalServerError, res)
    }
    defer db.Close()

    _, err = db.Exec("INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)",
		name, category, fileHash+".jpg")
	if err != nil {
		res := Response{Message: "Error saving item to database"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	message := fmt.Sprintf("item received: %s, Category:%s", name, category)
	imageDetails := ImageDetails{Name: file.Filename, Path: ImgDir + "/" + fileHash + ".jpg"}
	res := Response{Message: message, ImageDetails: imageDetails}
	return c.JSON(http.StatusOK, res)
}
    
func loadItems() ([]map[string]interface{}, error) {
    file, err := os.ReadFile("items.json")
    if err != nil {
        return nil, err
    }

    var items []map[string]interface{}
    if err := json.Unmarshal(file, &items); err != nil {
        fmt.Println("Error unmarshalling JSON:", err)
        return nil, err
    }

    return items, nil
}

func saveItems(items []map[string]interface{}) error {
    data, err := json.MarshalIndent(items, "", "  ")
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
	db, err:= initDB()
	if err!= nil {
		res := Response{Message: "Error loading items"}
		return c.JSON(http.StatusInternalServerError, res)
	}

	defer db.Close()

	rows, err:= db.Query("SELECT * FROM items")
	if err != nil {
		res := Response{Message: "Error retrieveing items from database"}
		return c.JSON(http.StatusInternalServerError, res)
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id int
		var name, category, imageName string
		if err:= rows.Scan(&id, &name, &category, &imageName); err != nil {
			res := Response{Message: "Error scanning rows"}
			return c.JSON(http.StatusInternalServerError, res)
		}
		newItem := map[string]interface{}{
			"id": id,
			"name": name,
			"category": category,
			"image_name": imageName,
		}
		items= append(items, newItem)
	}

	return c.JSON(http.StatusOK, items)
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "mercari.sqlite3")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func getItemDetails(c echo.Context) error {
    itemID := c.Param("item_id")

    db, err := initDB()
    if err != nil {
        res := Response{Message: "Error initializing database"}
        return c.JSON(http.StatusInternalServerError, res)
    }
    defer db.Close()

    row := db.QueryRow("SELECT * FROM items WHERE id = ?", itemID)

    var id int
    var name, category, imageName string
    err = row.Scan(&id, &name, &category, &imageName)
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

	imageName := path.Base(imgPath)

	res := Response{Message: "Image received", ImageDetails: ImageDetails{Name: imageName, Path: imgPath}}
	return c.JSON(http.StatusOK, res)
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
