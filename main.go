package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"regexp"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/phettb/rmc/model"
	rawmaterial "github.com/phettb/rmc/model/rawMaterial"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	client, err := model.Connect()
	if err != nil {
		log.Panic(err)
	}

	db := client.Database("rmc")

	app := fiber.New()

	app.Use(cors.New())

	app.Get("/Hello/:name", func(c *fiber.Ctx) error {
		result := "Hello,You name is " + c.Params("name")
		return c.JSON(result)
	})

	app.Get("/RM/", func(c *fiber.Ctx) error {
		result, err := rawmaterial.List(context.Background(), db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		return c.JSON(result)
	})

	app.Get("/RM/Get/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		pid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		result, err := rawmaterial.Read(context.Background(), db, pid)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		return c.JSON(result)
	})

	app.Post("/RM/", func(c *fiber.Ctx) error {
		var form rawmaterial.RawMaterial
		err := c.BodyParser(&form)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("invalid body request")
		}

		id, err := form.Create(context.Background(), db)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		return c.JSON(id)
	})

	app.Put("/RM/", func(c *fiber.Ctx) error {
		var form rawmaterial.RawMaterial
		err := c.BodyParser(&form)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("invalid body request")
		}

		result, err := form.Update(context.Background(), db, form.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		return c.JSON(result)
	})

	app.Delete("/RM/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		pid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		result, err := rawmaterial.Delete(context.Background(), db, pid)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		return c.JSON(result)
	})

	app.Post("/api/image", func(c *fiber.Ctx) error {
		// Check if file is present in request body or not
		fileHeader, err := c.FormFile("image")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}

		// Check if file is of type image or not
		fileExtension := regexp.MustCompile(`\.[a-zA-Z0-9]+$`).FindString(fileHeader.Filename)
		if fileExtension != ".jpg" && fileExtension != ".jpeg" && fileExtension != ".png" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": true,
				"msg":   "Invalid file type",
			})
		}

		// Read file content
		file, err := fileHeader.Open()
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}
		content, err := io.ReadAll(file)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}

		// Create bucket
		bucket, err := gridfs.NewBucket(db, options.GridFSBucket().SetName("images"))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}

		// Upload file to GridFS bucket
		uploadStream, err := bucket.OpenUploadStream(fileHeader.Filename, options.GridFSUpload().SetMetadata(fiber.Map{"ext": fileExtension}))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}

		// Close upload stream
		fieldId := uploadStream.FileID
		defer uploadStream.Close()

		// Write file content to upload stream
		fileSize, err := uploadStream.Write(content)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}

		// Return response
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"error": false,
			"msg":   "Image uploaded successfully",
			"image": fiber.Map{
				"id":   fieldId,
				"name": fileHeader.Filename,
				"size": fileSize,
			},
		})
	})

	app.Get("/api/image/id/:id", func(c *fiber.Ctx) error {
		// Get image id from request params and convert it to ObjectID
		id, err := primitive.ObjectIDFromHex(c.Params("id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}

		// Create variable to store image metadata
		var avatarMetadata bson.M

		// Get image metadata from GridFS bucket
		if err := db.Collection("images.files").FindOne(c.Context(), fiber.Map{"_id": id}).Decode(&avatarMetadata); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": true,
				"msg":   "Avatar not found",
			})
		}

		// Create buffer to store image content
		var buffer bytes.Buffer
		// Create bucket
		bucket, _ := gridfs.NewBucket(db, options.GridFSBucket().SetName("images"))
		// Download image from GridFS bucket to buffer
		bucket.DownloadToStream(id, &buffer)

		// Set required headers
		setResponseHeaders(c, buffer, avatarMetadata["metadata"].(bson.M)["ext"].(string))

		// Return image
		return c.Send(buffer.Bytes())
	})

	app.Delete("/api/image/id/:id", func(c *fiber.Ctx) error {
		// Get image id from request params and convert it to ObjectID
		id, err := primitive.ObjectIDFromHex(c.Params("id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}

		// Create bucket
		bucket, _ := gridfs.NewBucket(db, options.GridFSBucket().SetName("images"))

		// Delete image from GridFS bucket
		if err := bucket.Delete(id); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}

		// Return success message
		return c.JSON(fiber.Map{
			"error": false,
			"msg":   "Image deleted successfully",
		})
	})

	app.Listen(":8080")
}

func setResponseHeaders(c *fiber.Ctx, buff bytes.Buffer, ext string) error {
	switch ext {
	case ".png":
		c.Set("Content-Type", "image/png")
	case ".jpg":
		c.Set("Content-Type", "image/jpeg")
	case ".jpeg":
		c.Set("Content-Type", "image/jpeg")
	}

	c.Set("Cache-Control", "public, max-age=31536000")
	c.Set("Content-Length", strconv.Itoa(len(buff.Bytes())))

	return c.Next()
}
