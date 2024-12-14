package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"product-management/database"
	"product-management/models"

	"github.com/jackc/pgtype"

	_ "image/jpeg" // To handle JPEG images
	_ "image/png"  // To handle PNG images
)

func processImages() {
	channel, err := database.GetRabbitMQChannel()
	if err != nil {
		log.Fatal("Failed to create RabbitMQ channel:", err)
	}
	defer channel.Close()

	msgs, err := channel.Consume(
		"image_queue", // Queue name
		"",            // Consumer name
		true,          // Auto-ack
		false,         // Exclusive
		false,         // No-local
		false,         // No-wait
		nil,           // Args
	)
	if err != nil {
		log.Fatal("Failed to register consumer:", err)
	}

	forever := make(chan bool)

	go func() {
		for msg := range msgs {
			imageURL := string(msg.Body)
			log.Println("Processing image:", imageURL)

			// Add your image processing logic here
			compressImage(imageURL)
		}
	}()

	log.Println("Waiting for messages...")
	<-forever
}

func compressImage(imageURL string) {
	// Download the image
	resp, err := http.Get(imageURL)
	if err != nil {
		log.Println("Failed to download image:", err)
		return
	}
	defer resp.Body.Close()

	// Decode the image
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Println("Failed to decode image:", err)
		return
	}

	// Create a buffer to store the compressed image
	var compressedImage bytes.Buffer

	// Compress the image with a quality of 75
	err = jpeg.Encode(&compressedImage, img, &jpeg.Options{Quality: 75})
	if err != nil {
		log.Println("Failed to compress image:", err)
		return
	}

	// Save the compressed image to a file (you can update this logic to save to cloud storage)
	fileName := "compressed_image.jpg" // Replace with a unique name for each image
	err = os.WriteFile(fileName, compressedImage.Bytes(), 0644)
	if err != nil {
		log.Println("Failed to save compressed image:", err)
		return
	}

	log.Println("Compressed image saved successfully:", fileName)

	// Optionally, update the database (if required)
	compressedImageArray := pgtype.TextArray{
		Elements:   []pgtype.Text{{String: fileName, Status: pgtype.Present}},
		Dimensions: []pgtype.ArrayDimension{{Length: 1, LowerBound: 1}},
		Status:     pgtype.Present,
	}

	// Update the database with the compressed image array
	err = database.DB.Model(&models.Product{}).
		Where("product_images @> ?", `{`+imageURL+`}`).
		Update("compressed_product_images", compressedImageArray).Error
	if err != nil {
		log.Println("Failed to update compressed image in database:", err)
	}
}
